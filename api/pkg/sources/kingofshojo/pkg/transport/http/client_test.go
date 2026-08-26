// SPDX-License-Identifier: AGPL-3.0-or-later

package http_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/domain"
	koshttp "github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/transport/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/challengesolver"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func testDeps(httpClient *http.Client, solver domain.Solver) koshttp.Deps {
	return koshttp.Deps{HTTP: httpClient, Logger: discardLogger(), Solver: solver}
}

func newTestClient(t *testing.T, handler http.HandlerFunc) *koshttp.Client {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c, err := koshttp.New(koshttp.Config{BaseURL: srv.URL}, testDeps(srv.Client(), nil))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return c
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "parse", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}

	return raw
}

func TestSearchReturnsItemsAndTotal(t *testing.T) {
	t.Parallel()

	body := readFixture(t, "search.html")

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/manga/" {
			t.Errorf("path = %q, want /manga/", r.URL.Path)
		}

		_, _ = w.Write(body)
	})

	res, err := c.Search(context.Background(), domain.SearchCacheOpts{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if res.Meta.Total != 21 {
		t.Errorf("total = %d, want 21", res.Meta.Total)
	}

	if len(res.Items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(res.Items))
	}

	if res.Items[0].Slug != "tears-on-a-withered-flower" {
		t.Errorf("item0 slug = %q", res.Items[0].Slug)
	}
}

func TestSearchTotalWithoutNextIsOffsetPlusLen(t *testing.T) {
	t.Parallel()

	body := `<!DOCTYPE html><html><body><div class="listupd">` +
		`<div class="bsx"><a href="/manga/only/"><div class="tt">Only</div><div class="adds">Chapter 1</div></a></div>` +
		`</div><div class="hpage"></div></body></html>`

	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	})

	res, err := c.Search(context.Background(), domain.SearchCacheOpts{Offset: 40, Limit: 20})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(res.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(res.Items))
	}

	if res.Meta.Total != 41 {
		t.Errorf("total = %d, want 41", res.Meta.Total)
	}
}

func TestSearchQueryOrderOnly(t *testing.T) {
	t.Parallel()

	var got url.Values

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query()
		_, _ = w.Write(readFixture(t, "search.html"))
	})

	if _, err := c.Search(context.Background(), domain.SearchCacheOpts{Sort: domain.SortTypePopular}); err != nil {
		t.Fatalf("Search: %v", err)
	}

	if got.Get("order") != "popular" {
		t.Errorf("order = %q, want popular", got.Get("order"))
	}

	for _, key := range []string{"genre", "status", "type", "title"} {
		if _, ok := got[key]; ok {
			t.Errorf("query[%s] present, want absent", key)
		}
	}
}

func TestSearchChallengeWithoutSolver(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`<html><body>cf-browser-verification</body></html>`))
	})

	_, err := c.Search(context.Background(), domain.SearchCacheOpts{})
	if !errors.Is(err, domain.ErrChallenge) {
		t.Errorf("Search = %v, want ErrChallenge", err)
	}
}

type fakeSolver struct {
	client *http.Client
	calls  atomic.Int32
}

func (f *fakeSolver) Session(
	_ context.Context,
	_ string,
	_ ...challengesolver.Request,
) (*http.Client, *challengesolver.Solution, error) {
	f.calls.Add(1)

	return f.client, &challengesolver.Solution{}, nil
}

func TestSearchChallengeWithSolverRetriesOnce(t *testing.T) {
	t.Parallel()

	body := readFixture(t, "search.html")
	var attempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`<html><body>Just a moment</body></html>`))

			return
		}

		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	solver := &fakeSolver{client: srv.Client()}

	c, err := koshttp.New(koshttp.Config{BaseURL: srv.URL}, testDeps(srv.Client(), solver))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := c.Search(context.Background(), domain.SearchCacheOpts{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(res.Items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(res.Items))
	}

	if solver.calls.Load() != 1 {
		t.Errorf("solver calls = %d, want 1", solver.calls.Load())
	}
}

func TestGetSeriesPageNotFound(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := c.GetSeriesPage(context.Background(), "missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetSeriesPage = %v, want ErrNotFound", err)
	}
}

func TestGetImageURLsByChapter(t *testing.T) {
	t.Parallel()

	body := readFixture(t, "reader.html")
	var gotPath string

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write(body)
	})

	urlsPtr, err := c.GetImageURLsByChapter(context.Background(), domain.GetImageURLsByChapterOpts{
		SeriesSlug: "tears-on-a-withered-flower",
		ChapterID:  "tears-on-a-withered-flower-chapter-115",
	})
	if err != nil {
		t.Fatalf("GetImageURLsByChapter: %v", err)
	}

	if gotPath != "/tears-on-a-withered-flower-chapter-115/" {
		t.Errorf("path = %q", gotPath)
	}

	want := []string{
		"https://cdn.example/page1.jpg",
		"https://cdn.example/page2.jpg",
		"https://cdn.example/page3.jpg",
		"https://cdn.example/page6.jpg",
	}

	if urlsPtr == nil {
		t.Fatal("nil urls")
	}

	urls := *urlsPtr
	if len(urls) != len(want) {
		t.Fatalf("len(urls) = %d, want %d: %v", len(urls), len(want), urls)
	}

	for i := range want {
		if urls[i] != want[i] {
			t.Errorf("urls[%d] = %q, want %q", i, urls[i], want[i])
		}
	}
}
