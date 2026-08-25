// SPDX-License-Identifier: AGPL-3.0-or-later

package kingofshojo_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
	kingofshojo "github.com/kharente-deuh/uchiyomi-server/pkg/core/covers/source/kingofshojo"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	kosdomain "github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/domain"
)

type fakeGetter struct {
	err   error
	infos *sources.GetInfosBySlugResponse
}

func (f fakeGetter) GetInfosBySlug(
	_ context.Context,
	_ sources.GetInfosBySlugOpts,
) (*sources.GetInfosBySlugResponse, error) {
	return f.infos, f.err
}

func newResolver(t *testing.T, getter kingofshojo.InfosBySlugGetter, client *http.Client) *kingofshojo.Resolver {
	t.Helper()

	r, err := kingofshojo.New(
		kingofshojo.Config{},
		kingofshojo.Deps{
			Getter:     getter,
			HTTPClient: client,
			Logger:     slog.New(slog.DiscardHandler),
		},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return r
}

func TestResolveExternalURLAbsolute(t *testing.T) {
	t.Parallel()

	want := "https://cdn.example/cover.webp"
	r := newResolver(t, fakeGetter{infos: &sources.GetInfosBySlugResponse{
		Cover: want,
	}}, &http.Client{})

	got, err := r.ResolveExternalURL(context.Background(), "solo-shojo")
	if err != nil {
		t.Fatalf("ResolveExternalURL: %v", err)
	}

	if got != want {
		t.Errorf("url = %q, want %q", got, want)
	}
}

func TestResolveExternalURLNotFound(t *testing.T) {
	t.Parallel()

	r := newResolver(t, fakeGetter{err: kosdomain.ErrNotFound}, &http.Client{})

	_, err := r.ResolveExternalURL(context.Background(), "missing")
	if !errors.Is(err, covers.ErrSeriesNotFound) {
		t.Errorf("ResolveExternalURL = %v, want ErrSeriesNotFound", err)
	}
}

func TestResolveExternalURLMissingCover(t *testing.T) {
	t.Parallel()

	r := newResolver(t, fakeGetter{infos: &sources.GetInfosBySlugResponse{}}, &http.Client{})

	_, err := r.ResolveExternalURL(context.Background(), "solo-shojo")
	if err == nil {
		t.Fatal("empty cover must fail")
	}
}

func TestResolveExternalURLRelative(t *testing.T) {
	t.Parallel()

	r := newResolver(t, fakeGetter{infos: &sources.GetInfosBySlugResponse{
		Cover: "/wp-content/cover.webp",
	}}, &http.Client{})

	_, err := r.ResolveExternalURL(context.Background(), "solo-shojo")
	if err == nil {
		t.Fatal("relative cover must fail")
	}
}

func TestFetchOK(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("webp"))
	}))
	t.Cleanup(srv.Close)

	r := newResolver(t, fakeGetter{infos: &sources.GetInfosBySlugResponse{Cover: "x"}}, srv.Client())

	rc, err := r.Fetch(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	t.Cleanup(func() { _ = rc.Close() })

	body, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	if string(body) != "webp" {
		t.Errorf("body = %q", body)
	}
}

func TestFetchNotFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	r := newResolver(t, fakeGetter{infos: &sources.GetInfosBySlugResponse{Cover: "x"}}, srv.Client())

	_, err := r.Fetch(context.Background(), srv.URL)
	if !errors.Is(err, covers.ErrSeriesNotFound) {
		t.Errorf("Fetch = %v, want ErrSeriesNotFound", err)
	}
}

func TestNewValidates(t *testing.T) {
	t.Parallel()

	_, err := kingofshojo.New(kingofshojo.Config{}, kingofshojo.Deps{})
	if err == nil {
		t.Fatal("New with empty deps must fail")
	}
}
