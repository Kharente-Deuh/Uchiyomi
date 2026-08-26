// SPDX-License-Identifier: AGPL-3.0-or-later

package http_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
	asurascanshttp "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/transport/http"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func testDeps(httpClient *http.Client) asurascanshttp.Deps {
	return asurascanshttp.Deps{Http: httpClient, Logger: discardLogger()}
}

func newTestClient(t *testing.T, handler http.HandlerFunc) *asurascanshttp.Client {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c, err := asurascanshttp.New(testDeps(srv.Client()), asurascanshttp.Config{AsuraURL: srv.URL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return c
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		cfg     asurascanshttp.Config
		wantErr string
	}{
		"valide":       {cfg: asurascanshttp.Config{AsuraURL: "https://api.example.com"}},
		"empty URL":    {cfg: asurascanshttp.Config{}, wantErr: "asuraURL is not a valid URL"},
		"URL relative": {cfg: asurascanshttp.Config{AsuraURL: "example.com"}, wantErr: "asuraURL is not a valid URL"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := tc.cfg
			err := cfg.Validate()

			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}

				return
			}

			if err == nil || err.Error() != tc.wantErr {
				t.Errorf("Validate() = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestDepsValidate(t *testing.T) {
	t.Parallel()

	var deps asurascanshttp.Deps
	if err := deps.Validate(); err == nil {
		t.Error("Validate() without HTTP client must fail")
	}

	deps = asurascanshttp.Deps{Http: http.DefaultClient}
	if err := deps.Validate(); err == nil {
		t.Error("Validate() without logger must fail")
	}

	deps = testDeps(http.DefaultClient)
	if err := deps.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

func TestNewRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	if _, err := asurascanshttp.New(asurascanshttp.Deps{}, asurascanshttp.Config{AsuraURL: "https://x.dev"}); err == nil {
		t.Error("New without HTTP client must fail")
	}

	if _, err := asurascanshttp.New(testDeps(http.DefaultClient), asurascanshttp.Config{}); err == nil {
		t.Error("New with empty config must fail")
	}

	_, err := asurascanshttp.New(
		asurascanshttp.Deps{Http: http.DefaultClient},
		asurascanshttp.Config{AsuraURL: "https://x.dev"},
	)

	if err == nil {
		t.Error("New without logger must fail")
	}
}

func TestSearchRequestShape(t *testing.T) {
	t.Parallel()

	var got *url.URL

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		got = r.URL

		if accept := r.Header.Get("Accept"); accept != "application/json" {
			t.Errorf("Accept = %q, want %q", accept, "application/json")
		}

		_, _ = w.Write([]byte(`{"data":[],"meta":{}}`))
	})

	if _, err := c.Search(context.Background(), domain.SearchCacheOpts{Page: 1}); err != nil {
		t.Fatalf("Search: %v", err)
	}

	if got == nil {
		t.Fatal("no request received")
	}

	if got.Path != "/series" {
		t.Errorf("path = %q, want %q", got.Path, "/series")
	}

	q := got.Query()

	for key, want := range map[string]string{
		"offset": "0",
		"limit":  "20",
		"sort":   string(domain.SortTypePopular),
		"order":  string(domain.SortOrderDesc),
	} {
		if q.Get(key) != want {
			t.Errorf("query[%s] = %q, want %q", key, q.Get(key), want)
		}
	}
}

func TestSearchQueryParameters(t *testing.T) {
	t.Parallel()

	var got url.Values

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query()
		_, _ = w.Write([]byte(`{"data":[],"meta":{}}`))
	}))
	t.Cleanup(srv.Close)

	c, err := asurascanshttp.New(testDeps(srv.Client()), asurascanshttp.Config{AsuraURL: srv.URL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	opts := domain.SearchCacheOpts{
		Page:      3,
		Search:    "one piece",
		Sort:      domain.SortTypeTitle,
		SortOrder: domain.SortOrderAsc,
		Status:    sources.SeriesStatusOngoing,
		Type:      "manga",
		Artist:    "oda",
		Genres:    []string{"action", "adventure"},
	}

	if _, err := c.Search(context.Background(), opts); err != nil {
		t.Fatalf("Search: %v", err)
	}

	tests := map[string]string{
		"offset": "40",
		"limit":  "20",
		"search": "one piece",
		"sort":   "title",
		"order":  "asc",
		"status": string(sources.SeriesStatusOngoing),
		"type":   "manga",
		"artist": "oda",
		"genres": "action,adventure",
	}

	for key, want := range tests {
		if got.Get(key) != want {
			t.Errorf("query[%s] = %q, want %q", key, got.Get(key), want)
		}
	}
}

func TestSearchDefaultOrderFollowsSort(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		sort      domain.SortType
		wantOrder string
	}{
		"tri par titre => ascendant":       {sort: domain.SortTypeTitle, wantOrder: "asc"},
		"sort by popularity => descending": {sort: domain.SortTypePopular, wantOrder: "desc"},
		"default sort => descending":       {sort: "", wantOrder: "desc"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got url.Values

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = r.URL.Query()
				_, _ = w.Write([]byte(`{"data":[],"meta":{}}`))
			}))
			t.Cleanup(srv.Close)

			c, err := asurascanshttp.New(testDeps(srv.Client()), asurascanshttp.Config{AsuraURL: srv.URL})
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			if _, err := c.Search(context.Background(), domain.SearchCacheOpts{Sort: tc.sort}); err != nil {
				t.Fatalf("Search: %v", err)
			}

			if got.Get("order") != tc.wantOrder {
				t.Errorf("order = %q, want %q", got.Get("order"), tc.wantOrder)
			}
		})
	}
}

func TestSearchDecodesResponse(t *testing.T) {
	t.Parallel()

	const body = `{
		"data": [{
			"id": 12,
			"slug": "one-piece",
			"title": "One Piece",
			"description": "pirates",
			"cover": "cover.jpg",
			"status": "ongoing",
			"type": "manga",
			"author": "Oda",
			"artist": "Oda",
			"rating": 9.5,
			"chapter_count": 1100,
			"genres": [{"id": 1, "name": "Action", "slug": "action"}],
			"latest_chapters": [{"id": 9, "number": 1100, "slug": "1100", "title": "Chapitre 1100"}]
		}],
		"meta": {"total": 1, "per_page": 20, "has_more": true}
	}`

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	})

	res, err := c.Search(context.Background(), domain.SearchCacheOpts{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if !res.Meta.HasNextPage {
		t.Errorf("meta = %+v", res.Meta)
	}

	if len(res.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(res.Items))
	}

	item := res.Items[0]
	if item.ID != 12 || item.Slug != "one-piece" || item.Title != "One Piece" {
		t.Errorf("item = %+v", item)
	}

	if item.Status != sources.SeriesStatusOngoing {
		t.Errorf("status = %q, want %q", item.Status, sources.SeriesStatusOngoing)
	}

	if len(item.Genres) != 1 || item.Genres[0] != string("action") {
		t.Errorf("genres = %v", item.Genres)
	}

	if len(item.LatestChapters) != 1 || item.LatestChapters[0].Number != 1100 {
		t.Errorf("latestChapters = %+v", item.LatestChapters)
	}
}

func TestSearchFractionalChapterNumber(t *testing.T) {
	t.Parallel()

	const body = `{
		"data": [{
			"id": 1,
			"slug": "test",
			"title": "Test",
			"description": "desc",
			"cover": "cover.jpg",
			"status": "ongoing",
			"type": "manga",
			"rating": 8,
			"chapter_count": 1,
			"genres": [],
			"latest_chapters": [{"id": 1, "number": 112.7, "slug": "chapter-112-7"}]
		}],
		"meta": {"total": 1, "per_page": 20, "has_more": false}
	}`

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	})

	res, err := c.Search(context.Background(), domain.SearchCacheOpts{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if res.Items[0].LatestChapters[0].Number != 112.7 {
		t.Errorf("chapter number = %v, want 112.7", res.Items[0].LatestChapters[0].Number)
	}
}

func TestSearchServerErrors(t *testing.T) {
	t.Parallel()

	tests := map[string]int{
		"500": http.StatusInternalServerError,
		"404": http.StatusNotFound,
		"403": http.StatusForbidden,
	}

	for name, status := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
			})

			if _, err := c.Search(context.Background(), domain.SearchCacheOpts{}); err == nil {
				t.Errorf("Search on %d = nil, want error", status)
			}
		})
	}
}

func TestSearchMalformedJSON(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data": [`))
	})

	if _, err := c.Search(context.Background(), domain.SearchCacheOpts{}); err == nil {
		t.Error("Search on truncated JSON = nil, want error")
	}
}

func TestSearchHonoursContext(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[],"meta":{}}`))
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := c.Search(ctx, domain.SearchCacheOpts{}); !errors.Is(err, context.Canceled) {
		t.Errorf("Search = %v, want context.Canceled", err)
	}
}

func TestGetInfosBySlug(t *testing.T) {
	t.Parallel()

	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"series":{"id":3,"slug":"solo-leveling","title":"Solo Leveling"}}`))
	}))
	t.Cleanup(srv.Close)

	c, err := asurascanshttp.New(testDeps(srv.Client()), asurascanshttp.Config{AsuraURL: srv.URL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if _, err := c.GetInfosBySlug(context.Background(), "solo-leveling"); err != nil {
		t.Fatalf("GetInfosBySlug: %v", err)
	}

	if want := "/series/solo-leveling"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestGetInfosBySlugNotFound(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	if _, err := c.GetInfosBySlug(context.Background(), "inconnu"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetInfosBySlug = %v, want domain.ErrNotFound", err)
	}
}

func TestGetChaptersListBySerie(t *testing.T) {
	t.Parallel()

	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":[
			{"id":1,"slug":"chapter-1","number":1,"title":"The beginning"},
			{"id":2,"slug":"chapter-2","number":2}
		]}`))
	}))
	t.Cleanup(srv.Close)

	c, err := asurascanshttp.New(testDeps(srv.Client()), asurascanshttp.Config{AsuraURL: srv.URL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	chaptersPtr, err := c.GetChaptersListBySerie(context.Background(), "one-piece")
	if err != nil {
		t.Fatalf("GetChaptersListBySerie: %v", err)
	}

	if chaptersPtr == nil {
		t.Fatal("GetChaptersListBySerie returned nil result without error")
	}

	chapters := *chaptersPtr

	if want := "/series/one-piece/chapters"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}

	if len(chapters) != 2 {
		t.Fatalf("len(chapters) = %d, want 2", len(chapters))
	}

	if chapters[0].ID != "chapter-1" {
		t.Errorf("chapters[0].ID = %q, want %q", chapters[0].ID, "chapter-1")
	}

	if chapters[0].Number != 1 || chapters[0].Title != "The beginning" {
		t.Errorf("chapters[0] = %+v", chapters[0])
	}

	if chapters[1].Title != "" {
		t.Errorf("chapters[1].Title = %q, want empty", chapters[1].Title)
	}
}

func TestGetChaptersListBySerieNotFound(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	if _, err := c.GetChaptersListBySerie(context.Background(), "inconnu"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetChaptersListBySerie = %v, want domain.ErrNotFound", err)
	}
}

func TestGetImageURLsByChapter(t *testing.T) {
	t.Parallel()

	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"data":{"chapter":{"id":7,"pages":[
			{"url":"https://cdn/1.webp","width":800,"height":1200},
			{"url":"https://cdn/2.webp","width":800,"height":1200}
		]}}}`))
	}))
	t.Cleanup(srv.Close)

	c, err := asurascanshttp.New(testDeps(srv.Client()), asurascanshttp.Config{AsuraURL: srv.URL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	urlsPtr, err := c.GetImageURLsByChapter(context.Background(), domain.GetImageURLsByChapterOpts{
		SeriesSlug: "one-piece",
		ChapterID:  "1100",
	})

	if err != nil {
		t.Fatalf("GetImageURLsByChapter: %v", err)
	}

	if urlsPtr == nil {
		t.Fatal("GetImageURLsByChapter returned nil result without error")
	}

	urls := *urlsPtr

	if want := "/series/one-piece/chapters/1100"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}

	want := []string{"https://cdn/1.webp", "https://cdn/2.webp"}
	if len(urls) != len(want) {
		t.Fatalf("len(urls) = %d, want %d", len(urls), len(want))
	}

	for i := range want {
		if urls[i] != want[i] {
			t.Errorf("urls[%d] = %q, want %q", i, urls[i], want[i])
		}
	}
}

func TestGetImageURLsByChapterEmptyPages(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"chapter":{"id":7,"pages":[]}}}`))
	})

	urlsPtr, err := c.GetImageURLsByChapter(context.Background(), domain.GetImageURLsByChapterOpts{
		SeriesSlug: "s",
		ChapterID:  "1",
	})

	if err != nil {
		t.Fatalf("GetImageURLsByChapter: %v", err)
	}

	if urlsPtr == nil {
		t.Fatal("GetImageURLsByChapter returned nil result without error")
	}

	if len(*urlsPtr) != 0 {
		t.Errorf("urls = %v, want empty", *urlsPtr)
	}
}

func TestGetImageURLsByChapterNotFound(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := c.GetImageURLsByChapter(context.Background(), domain.GetImageURLsByChapterOpts{
		SeriesSlug: "s",
		ChapterID:  "1",
	})

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetImageURLsByChapter = %v, want domain.ErrNotFound", err)
	}
}

func TestSearchNotFound(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	if _, err := c.Search(context.Background(), domain.SearchCacheOpts{}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Search on 404 = %v, want domain.ErrNotFound", err)
	}
}
