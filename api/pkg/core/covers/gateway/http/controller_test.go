// SPDX-License-Identifier: AGPL-3.0-or-later

package http_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
	covershttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/covers/gateway/http"
	coredomain "github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/imgcache"
)

const testExternalCoverURL = "https://cdn.example/cover.webp"

type fakeResolver struct {
	err   error
	fetch func(context.Context, string) (io.ReadCloser, error)
	url   string
}

func (f *fakeResolver) ResolveExternalURL(context.Context, string) (string, error) {
	if f.err != nil {
		return "", f.err
	}

	return f.url, nil
}

func (f *fakeResolver) Fetch(ctx context.Context, externalURL string) (io.ReadCloser, error) {
	if f.fetch != nil {
		return f.fetch(ctx, externalURL)
	}

	return io.NopCloser(bytes.NewReader([]byte("cover-bytes"))), nil
}

type stubFinder struct {
	err error
	id  uuid.UUID
}

func (f stubFinder) FindBySourceSlug(context.Context, string, string) (uuid.UUID, error) {
	if f.err != nil {
		return uuid.Nil, f.err
	}

	return f.id, nil
}

func newTestControllerWithDirs(
	t *testing.T,
	resolvers map[string]covers.CoverResolver,
	finder covers.LocalComicFinder,
) (*covershttp.Controller, string, string) {
	t.Helper()

	cacheDir := t.TempDir()
	downloadsDir := t.TempDir()

	cache, err := imgcache.New(imgcache.Config{
		Dir:           cacheDir,
		FetchFn:       covers.NewFetchFn(resolvers),
		ErrorCacheTTL: time.Minute,
		MinInterval:   time.Millisecond,
		Logger:        slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("imgcache.New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = cache.Run(ctx)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for !cache.IsReady() {
		if time.Now().After(deadline) {
			t.Fatal("cache never became ready")
		}

		time.Sleep(time.Millisecond)
	}

	svc, err := covers.NewService(
		covers.ServiceConfig{
			ProxyPathPrefix: "/api/sources/cover",
			DownloadsDir:    downloadsDir,
		},
		covers.ServiceDeps{
			Cache:      cache,
			Resolvers:  resolvers,
			HTTPClient: &http.Client{Timeout: time.Second},
			Logger:     slog.New(slog.DiscardHandler),
			Finder:     finder,
		},
	)
	if err != nil {
		t.Fatalf("covers.NewService: %v", err)
	}

	ctrl, err := covershttp.New(
		covershttp.Config{Endpoint: "/cover"},
		covershttp.Deps{Service: svc, Logger: slog.New(slog.DiscardHandler)},
	)
	if err != nil {
		t.Fatalf("covershttp.New: %v", err)
	}

	return ctrl, cacheDir, downloadsDir
}

func newTestController(t *testing.T, resolvers map[string]covers.CoverResolver) *covershttp.Controller {
	t.Helper()

	ctrl, _, _ := newTestControllerWithDirs(t, resolvers, stubFinder{err: coredomain.ErrNotFound})

	return ctrl
}

func serveCover(t *testing.T, ctrl *covershttp.Controller, path string) *httptest.ResponseRecorder {
	t.Helper()

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

	return rec
}

func TestServeCoverPrefersLocalWithoutFillingCache(t *testing.T) {
	t.Parallel()

	var fetchCount int
	comicID := uuid.New()

	resolvers := map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{
			url: testExternalCoverURL,
			fetch: func(context.Context, string) (io.ReadCloser, error) {
				fetchCount++

				return io.NopCloser(bytes.NewReader([]byte("cdn"))), nil
			},
		},
	}

	ctrl, cacheDir, downloadsDir := newTestControllerWithDirs(t, resolvers, stubFinder{id: comicID})

	coverPath := filepath.Join(downloadsDir, comicID.String(), "cover.webp")
	if err := os.MkdirAll(filepath.Dir(coverPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.WriteFile(coverPath, []byte("local-bytes"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	rec := serveCover(t, ctrl, "/cover/solo-leveling?source=asurascans")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "image/webp" {
		t.Errorf("Content-Type = %q, want image/webp", ct)
	}

	if !bytes.Equal(rec.Body.Bytes(), []byte("local-bytes")) {
		t.Errorf("body = %q, want local-bytes", rec.Body.Bytes())
	}

	if fetchCount != 0 {
		t.Errorf("fetchCount = %d, want 0", fetchCount)
	}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("cache filled: %v", entries)
	}
}

func TestServeCoverDownloadsAndCaches(t *testing.T) {
	t.Parallel()

	var fetchCount int

	resolvers := map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{
			url: testExternalCoverURL,
			fetch: func(context.Context, string) (io.ReadCloser, error) {
				fetchCount++

				return io.NopCloser(bytes.NewReader([]byte("webp-data"))), nil
			},
		},
	}

	ctrl := newTestController(t, resolvers)

	rec := serveCover(t, ctrl, "/cover/solo-leveling?source=asurascans")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "image/webp" {
		t.Errorf("Content-Type = %q, want image/webp", ct)
	}

	if !bytes.Equal(rec.Body.Bytes(), []byte("webp-data")) {
		t.Errorf("body = %q, want webp-data", rec.Body.Bytes())
	}

	rec = serveCover(t, ctrl, "/cover/solo-leveling?source=asurascans")
	if rec.Code != http.StatusOK {
		t.Fatalf("second status = %d, want %d", rec.Code, http.StatusOK)
	}

	if fetchCount != 1 {
		t.Errorf("fetchCount = %d, want 1", fetchCount)
	}
}

func TestServeCoverUnknownSource(t *testing.T) {
	t.Parallel()

	ctrl := newTestController(t, map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{url: "https://cdn.example/cover.webp"},
	})

	rec := serveCover(t, ctrl, "/cover/solo-leveling?source=unknown")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestServeCoverSeriesNotFound(t *testing.T) {
	t.Parallel()

	resolvers := map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{err: covers.ErrSeriesNotFound},
	}

	ctrl := newTestController(t, resolvers)

	rec := serveCover(t, ctrl, "/cover/missing?source=asurascans")
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestServeCoverDownloadFailure(t *testing.T) {
	t.Parallel()

	resolvers := map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{
			url: testExternalCoverURL,
			fetch: func(context.Context, string) (io.ReadCloser, error) {
				return nil, errors.New("network down")
			},
		},
	}

	ctrl := newTestController(t, resolvers)

	rec := serveCover(t, ctrl, "/cover/solo-leveling?source=asurascans")
	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadGateway)
	}
}

func TestServeCoverMissingSourceParam(t *testing.T) {
	t.Parallel()

	ctrl := newTestController(t, map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{url: "https://cdn.example/cover.webp"},
	})

	rec := serveCover(t, ctrl, "/cover/solo-leveling")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestServeCoverUsesDiskCachePath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	cache, err := imgcache.New(imgcache.Config{
		Dir: dir,
		FetchFn: func(context.Context, string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte("cached"))), nil
		},
		ErrorCacheTTL: time.Minute,
		MinInterval:   time.Millisecond,
	})
	if err != nil {
		t.Fatalf("imgcache.New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = cache.Run(ctx)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for !cache.IsReady() {
		if time.Now().After(deadline) {
			t.Fatal("cache never became ready")
		}

		time.Sleep(time.Millisecond)
	}

	resolvers := map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{url: "https://cdn.example/cover.jpg"},
	}

	svc, err := covers.NewService(
		covers.ServiceConfig{
			ProxyPathPrefix: "/api/sources/cover",
			DownloadsDir:    t.TempDir(),
		},
		covers.ServiceDeps{
			Cache:      cache,
			Resolvers:  resolvers,
			HTTPClient: &http.Client{Timeout: time.Second},
			Logger:     slog.New(slog.DiscardHandler),
			Finder:     stubFinder{err: coredomain.ErrNotFound},
		},
	)
	if err != nil {
		t.Fatalf("covers.NewService: %v", err)
	}

	ctrl, err := covershttp.New(
		covershttp.Config{Endpoint: "/cover"},
		covershttp.Deps{Service: svc, Logger: slog.New(slog.DiscardHandler)},
	)
	if err != nil {
		t.Fatalf("covershttp.New: %v", err)
	}

	serveCover(t, ctrl, "/cover/one-piece?source=asurascans")

	wantPath := filepath.Join(dir, "asurascans", "one-piece.jpg")
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("os.Stat(%s): %v", wantPath, err)
	}
}
