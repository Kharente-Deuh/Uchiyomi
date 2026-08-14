// SPDX-License-Identifier: AGPL-3.0-or-later

package asurascans_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
	asurascans "github.com/kharente-deuh/uchiyomi-server/pkg/core/covers/source/asurascans"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	asuradomain "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
)

type fakeGetter struct {
	err   error
	infos *sources.GetInfosBySlugResponse
}

func (f fakeGetter) GetInfosBySlug(
	_ context.Context,
	opts sources.GetInfosBySlugOpts,
) (*sources.GetInfosBySlugResponse, error) {
	return f.infos, f.err
}

func newResolver(t *testing.T, getter asurascans.InfosBySlugGetter, client *http.Client) *asurascans.Resolver {
	t.Helper()

	r, err := asurascans.New(
		asurascans.Config{CDNBaseURL: asurascans.DefaultCDNBaseURL},
		asurascans.Deps{
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

	r := newResolver(t, fakeGetter{infos: &sources.GetInfosBySlugResponse{
		Cover: "https://cdn.example/cover.webp",
	}}, &http.Client{})

	got, err := r.ResolveExternalURL(context.Background(), "solo-leveling")
	if err != nil {
		t.Fatalf("ResolveExternalURL: %v", err)
	}

	if got != "https://cdn.example/cover.webp" {
		t.Errorf("url = %q", got)
	}
}

func TestResolveExternalURLRelative(t *testing.T) {
	t.Parallel()

	r := newResolver(t, fakeGetter{infos: &sources.GetInfosBySlugResponse{
		Cover: "/storage/cover.webp",
	}}, &http.Client{})

	got, err := r.ResolveExternalURL(context.Background(), "solo-leveling")
	if err != nil {
		t.Fatalf("ResolveExternalURL: %v", err)
	}

	if got != asurascans.DefaultCDNBaseURL+"/storage/cover.webp" {
		t.Errorf("url = %q", got)
	}
}

func TestResolveExternalURLNotFound(t *testing.T) {
	t.Parallel()

	r := newResolver(t, fakeGetter{err: asuradomain.ErrNotFound}, &http.Client{})

	_, err := r.ResolveExternalURL(context.Background(), "missing")
	if !errors.Is(err, covers.ErrSeriesNotFound) {
		t.Errorf("ResolveExternalURL = %v, want ErrSeriesNotFound", err)
	}
}

func TestResolveExternalURLMissingCover(t *testing.T) {
	t.Parallel()

	r := newResolver(t, fakeGetter{infos: &sources.GetInfosBySlugResponse{}}, &http.Client{})

	_, err := r.ResolveExternalURL(context.Background(), "solo-leveling")
	if err == nil {
		t.Fatal("empty cover must fail")
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

	_, err := asurascans.New(asurascans.Config{}, asurascans.Deps{})
	if err == nil {
		t.Fatal("New with empty config must fail")
	}
}
