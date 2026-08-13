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
	"testing"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
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

func startService(t *testing.T, resolvers map[string]covers.CoverResolver) *covers.Service {
	t.Helper()

	cache, err := imgcache.New(imgcache.Config{
		Dir:           t.TempDir(),
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
		covers.ServiceConfig{ProxyPathPrefix: "/api/sources/cover"},
		covers.ServiceDeps{
			Cache:      cache,
			Resolvers:  resolvers,
			HTTPClient: &http.Client{Timeout: time.Second},
			Logger:     slog.New(slog.DiscardHandler),
		},
	)
	if err != nil {
		t.Fatalf("covers.NewService: %v", err)
	}

	return svc
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
