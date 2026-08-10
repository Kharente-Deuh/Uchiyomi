// SPDX-License-Identifier: AGPL-3.0-or-later

package core_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/core"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/fncache"
)

func newCache[P any, T any](
	t *testing.T, key func(P) string, fn func(context.Context, P) (*T, error),
) *fncache.Cache[P, T] {
	t.Helper()

	c, err := fncache.New(fncache.Config[P, T]{
		Name:          "test",
		Fn:            fn,
		Key:           key,
		TTL:           time.Hour,
		ErrorTTL:      time.Minute,
		FetchTimeout:  time.Minute,
		CleanInterval: time.Hour,
		MaxEntries:    16,
	}, fncache.Deps{Logger: slog.New(slog.DiscardHandler)})
	if err != nil {
		t.Fatalf("fncache.New: %v", err)
	}

	return c
}

func identityKey(s string) string { return s }

func fullDeps(t *testing.T) core.Dependencies {
	t.Helper()

	return core.Dependencies{
		Logger: slog.New(slog.DiscardHandler),
		SearchCache: newCache(t, domain.SearchOpts.CacheKey,
			func(context.Context, domain.SearchOpts) (*domain.SearchResult, error) {
				return &domain.SearchResult{}, nil
			}),
		GetInfosBySlugCache: newCache(t, identityKey,
			func(context.Context, string) (*domain.GetInfosBySlugResponse, error) {
				return &domain.GetInfosBySlugResponse{}, nil
			}),
		GetChaptersListBySeriesCache: newCache(t, identityKey,
			func(context.Context, string) (*[]domain.Chapter, error) {
				chapters := []domain.Chapter{{ID: "c1", Number: 1}}

				return &chapters, nil
			}),
		GetImageURLsByChapter: newCache(t, domain.GetImageURLsByChapterOpts.CacheKey,
			func(context.Context, domain.GetImageURLsByChapterOpts) (*[]string, error) {
				urls := []string{"https://example.test/1.jpg"}

				return &urls, nil
			}),
	}
}

func TestDependenciesValidate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		drop    func(*core.Dependencies)
		wantErr string
	}{
		"complet": {drop: func(*core.Dependencies) {}},
		"without searchCache": {
			drop:    func(d *core.Dependencies) { d.SearchCache = nil },
			wantErr: "searchCache is required",
		},
		"without getInfosBySlugCache": {
			drop:    func(d *core.Dependencies) { d.GetInfosBySlugCache = nil },
			wantErr: "getInfosBySlugCache is required",
		},
		"without getChaptersListBySeriesCache": {
			drop:    func(d *core.Dependencies) { d.GetChaptersListBySeriesCache = nil },
			wantErr: "getChaptersListBySeriesCache is required",
		},
		"without getImageURLsByChapter": {
			drop:    func(d *core.Dependencies) { d.GetImageURLsByChapter = nil },
			wantErr: "getImageURLsByChapter is required",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			deps := fullDeps(t)
			tc.drop(&deps)

			err := deps.Validate()

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

func TestNewRejectsIncompleteDeps(t *testing.T) {
	t.Parallel()

	app, err := core.New(core.Dependencies{})
	if err == nil {
		t.Fatal("New with empty dependencies must fail")
	}

	if app != nil {
		t.Error("New returned an App in addition to the error")
	}

	if want := "deps.Validate: searchCache is required"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
}

func TestNewSucceeds(t *testing.T) {
	t.Parallel()

	app, err := core.New(fullDeps(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if app == nil {
		t.Fatal("New returned nil App without error")
	}
}

func TestAppSearchDelegatesToCache(t *testing.T) {
	t.Parallel()

	called := make(chan domain.SearchOpts, 1)

	deps := fullDeps(t)
	deps.SearchCache = newCache(t, domain.SearchOpts.CacheKey,
		func(_ context.Context, opts domain.SearchOpts) (*domain.SearchResult, error) {
			called <- opts

			return &domain.SearchResult{Meta: domain.SearchResultMeta{Total: 3}}, nil
		})

	app, err := core.New(deps)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := app.Search(context.Background(), domain.SearchOpts{Search: "one piece"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if res == nil || res.Meta.Total != 3 {
		t.Errorf("result = %+v", res)
	}

	if opts := <-called; opts.Search != "one piece" {
		t.Errorf("opts transmises au cache = %+v", opts)
	}
}

func TestAppGetChaptersListBySeriesReturnsACopy(t *testing.T) {
	t.Parallel()

	app, err := core.New(fullDeps(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	first, err := app.GetChaptersListBySeries(context.Background(), "slug")
	if err != nil {
		t.Fatalf("GetChaptersListBySeries: %v", err)
	}

	if len(first) == 0 {
		t.Fatal("test cache must return at least one chapter")
	}

	first[0].ID = "corrompu"

	second, err := app.GetChaptersListBySeries(context.Background(), "slug")
	if err != nil {
		t.Fatalf("GetChaptersListBySeries (2nd call): %v", err)
	}

	if second[0].ID != "c1" {
		t.Errorf("second[0].ID = %q, want \"c1\": caller could modify cache entry", second[0].ID)
	}
}

func TestAppGetImageURLsByChapterReturnsACopy(t *testing.T) {
	t.Parallel()

	app, err := core.New(fullDeps(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	opts := domain.GetImageURLsByChapterOpts{SeriesSlug: "slug", ChapterID: "c1"}

	first, err := app.GetImageURLsByChapter(context.Background(), opts)
	if err != nil {
		t.Fatalf("GetImageURLsByChapter: %v", err)
	}

	if len(first) == 0 {
		t.Fatal("test cache must return at least one URL")
	}

	first[0] = "corrompu"

	second, err := app.GetImageURLsByChapter(context.Background(), opts)
	if err != nil {
		t.Fatalf("GetImageURLsByChapter (2nd call): %v", err)
	}

	if second[0] != "https://example.test/1.jpg" {
		t.Errorf("second[0] = %q: caller could modify cache entry", second[0])
	}
}

func TestAppRunReturnsOnContextCancel(t *testing.T) {
	t.Parallel()

	app, err := core.New(fullDeps(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)

	go func() { done <- app.Run(ctx) }()

	select {
	case err := <-done:
		if err == nil {
			t.Error("Run() = nil, want context error")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run never returned after context cancellation")
	}
}
