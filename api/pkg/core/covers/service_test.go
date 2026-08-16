// SPDX-License-Identifier: AGPL-3.0-or-later

package covers_test

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

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
	coredomain "github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/imgcache"
)

const testCoverURL = "https://cdn.example/cover.webp"

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

func startServiceWithDirs(
	t *testing.T,
	resolvers map[string]covers.CoverResolver,
	finder covers.LocalComicFinder,
) (*covers.Service, string, string) {
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

	return svc, cacheDir, downloadsDir
}

func startService(t *testing.T, resolvers map[string]covers.CoverResolver) *covers.Service {
	t.Helper()

	svc, _, _ := startServiceWithDirs(t, resolvers, stubFinder{err: coredomain.ErrNotFound})

	return svc
}

func startServiceFull(
	t *testing.T,
	resolvers map[string]covers.CoverResolver,
	finder covers.LocalComicFinder,
) (*covers.Service, string, string) {
	t.Helper()

	return startServiceWithDirs(t, resolvers, finder)
}

func TestBuildProxyURL(t *testing.T) {
	t.Parallel()

	svc := startService(t, map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{url: testCoverURL},
	})

	got := svc.BuildProxyURL(covers.SourceAsuraScans, "solo-leveling")
	want := "/api/sources/cover/solo-leveling?source=asurascans"
	if got != want {
		t.Errorf("BuildProxyURL = %q, want %q", got, want)
	}
}

func TestServeUnknownSource(t *testing.T) {
	t.Parallel()

	svc := startService(t, map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{url: testCoverURL},
	})

	_, _, err := svc.Serve(context.Background(), "unknown", "solo-leveling")
	if !errors.Is(err, covers.ErrUnknownSource) {
		t.Errorf("Serve = %v, want ErrUnknownSource", err)
	}
}

func TestServeSeriesNotFound(t *testing.T) {
	t.Parallel()

	svc := startService(t, map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{err: covers.ErrSeriesNotFound},
	})

	_, _, err := svc.Serve(context.Background(), covers.SourceAsuraScans, "missing")
	if !errors.Is(err, covers.ErrSeriesNotFound) {
		t.Errorf("Serve = %v, want ErrSeriesNotFound", err)
	}
}

func TestServeDownloadFailure(t *testing.T) {
	t.Parallel()

	svc := startService(t, map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{
			url: testCoverURL,
			fetch: func(context.Context, string) (io.ReadCloser, error) {
				return nil, errors.New("network down")
			},
		},
	})

	_, _, err := svc.Serve(context.Background(), covers.SourceAsuraScans, "solo-leveling")
	if !errors.Is(err, covers.ErrDownloadFailed) {
		t.Errorf("Serve = %v, want ErrDownloadFailed", err)
	}
}

func TestServeProbesExtensionWhenURLHasNone(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/webp")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	svc := startService(t, map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{url: upstream.URL + "/cover"},
	})

	path, contentType, err := svc.Serve(context.Background(), covers.SourceAsuraScans, "solo-leveling")
	if err != nil {
		t.Fatalf("Serve: %v", err)
	}

	if path == "" {
		t.Error("empty disk path")
	}

	if contentType != "image/webp" {
		t.Errorf("contentType = %q, want image/webp", contentType)
	}
}

func TestNewServiceValidates(t *testing.T) {
	t.Parallel()

	_, err := covers.NewService(covers.ServiceConfig{}, covers.ServiceDeps{})
	if err == nil {
		t.Fatal("NewService with empty config must fail")
	}
}

func TestNewFetchFnUnknownSource(t *testing.T) {
	t.Parallel()

	fn := covers.NewFetchFn(map[string]covers.CoverResolver{})

	_, err := fn(context.Background(), "asurascans/solo-leveling.webp")
	if !errors.Is(err, covers.ErrUnknownSource) {
		t.Errorf("NewFetchFn = %v, want ErrUnknownSource", err)
	}
}

func TestObtainLocalMovesCachedCover(t *testing.T) {
	t.Parallel()

	var fetchCount int
	resolvers := map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{
			url: testCoverURL,
			fetch: func(context.Context, string) (io.ReadCloser, error) {
				fetchCount++

				return io.NopCloser(bytes.NewReader([]byte("cover-bytes"))), nil
			},
		},
	}

	svc := startService(t, resolvers)
	ctx := context.Background()

	if _, _, err := svc.Serve(ctx, covers.SourceAsuraScans, "solo-leveling"); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	comicID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	if err := svc.ObtainLocal(ctx, comicID, covers.SourceAsuraScans, "solo-leveling"); err != nil {
		t.Fatalf("ObtainLocal: %v", err)
	}

	path, contentType, err := svc.ServeLocal(ctx, comicID)
	if err != nil {
		t.Fatalf("ServeLocal: %v", err)
	}

	if contentType != "image/webp" {
		t.Errorf("contentType = %q, want image/webp", contentType)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile: %v", err)
	}

	if string(got) != "cover-bytes" {
		t.Errorf("local cover = %q", got)
	}

	if fetchCount != 1 {
		t.Errorf("fetchCount = %d, want 1 (cache hit then move)", fetchCount)
	}
}

func TestObtainLocalDownloadsWhenCacheMisses(t *testing.T) {
	t.Parallel()

	var fetchCount int
	resolvers := map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{
			url: testCoverURL,
			fetch: func(context.Context, string) (io.ReadCloser, error) {
				fetchCount++

				return io.NopCloser(bytes.NewReader([]byte("fresh"))), nil
			},
		},
	}

	svc, cacheDir, _ := startServiceWithDirs(t, resolvers, stubFinder{err: coredomain.ErrNotFound})
	comicID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	if err := svc.ObtainLocal(context.Background(), comicID, covers.SourceAsuraScans, "solo-leveling"); err != nil {
		t.Fatalf("ObtainLocal: %v", err)
	}

	if fetchCount != 1 {
		t.Errorf("fetchCount = %d, want 1", fetchCount)
	}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		t.Fatalf("os.ReadDir cache: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("browse cache filled on miss: %v", entries)
	}

	path, _, err := svc.ServeLocal(context.Background(), comicID)
	if err != nil {
		t.Fatalf("ServeLocal: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile: %v", err)
	}

	if string(got) != "fresh" {
		t.Errorf("local cover = %q", got)
	}
}

func TestServeLocalMissing(t *testing.T) {
	t.Parallel()

	svc := startService(t, map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{url: testCoverURL},
	})

	_, _, err := svc.ServeLocal(context.Background(), uuid.New())
	if !errors.Is(err, covers.ErrLocalCoverMissing) {
		t.Errorf("ServeLocal = %v, want ErrLocalCoverMissing", err)
	}
}

func TestRemoveLocalDeletesComicDir(t *testing.T) {
	t.Parallel()

	resolvers := map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{url: testCoverURL},
	}
	svc := startService(t, resolvers)
	comicID := uuid.New()

	if err := svc.ObtainLocal(context.Background(), comicID, covers.SourceAsuraScans, "solo-leveling"); err != nil {
		t.Fatalf("ObtainLocal: %v", err)
	}

	if err := svc.RemoveLocal(comicID); err != nil {
		t.Fatalf("RemoveLocal: %v", err)
	}

	_, _, err := svc.ServeLocal(context.Background(), comicID)
	if !errors.Is(err, covers.ErrLocalCoverMissing) {
		t.Errorf("after RemoveLocal: %v", err)
	}
}

func TestServePrefersLocalCoverWithoutFillingCache(t *testing.T) {
	t.Parallel()

	var fetchCount int
	resolvers := map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{
			url: testCoverURL,
			fetch: func(context.Context, string) (io.ReadCloser, error) {
				fetchCount++

				return io.NopCloser(bytes.NewReader([]byte("cdn"))), nil
			},
		},
	}

	comicID := uuid.New()
	svc, cacheDir, downloadsDir := startServiceFull(t, resolvers, stubFinder{id: comicID})

	coverPath := filepath.Join(downloadsDir, comicID.String(), "cover.webp")
	if err := os.MkdirAll(filepath.Dir(coverPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.WriteFile(coverPath, []byte("local-bytes"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	path, contentType, err := svc.Serve(context.Background(), covers.SourceAsuraScans, "solo-leveling")
	if err != nil {
		t.Fatalf("Serve: %v", err)
	}

	if path != coverPath {
		t.Errorf("path = %q, want %q", path, coverPath)
	}

	if contentType != "image/webp" {
		t.Errorf("contentType = %q", contentType)
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

func TestNewFetchFnInvalidKey(t *testing.T) {
	t.Parallel()

	fn := covers.NewFetchFn(map[string]covers.CoverResolver{
		covers.SourceAsuraScans: &fakeResolver{url: testCoverURL},
	})

	_, err := fn(context.Background(), "badkey")
	if err == nil {
		t.Fatal("invalid cache key must fail")
	}
}
