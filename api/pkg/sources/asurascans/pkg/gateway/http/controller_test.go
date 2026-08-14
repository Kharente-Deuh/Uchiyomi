// SPDX-License-Identifier: AGPL-3.0-or-later

package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	sessionshttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
	coredomain "github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	asura "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/core"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
	asurahttp "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/fncache"
)

const testExternalCoverURL = "https://external.example/cover.webp"

type stubComicsRepository struct{}

func (stubComicsRepository) GetByID(context.Context, comics.GetByIDOpts) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (stubComicsRepository) GetBySourceSlug(
	context.Context, comics.GetBySourceSlugOpts,
) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (stubComicsRepository) Create(context.Context, comics.CreateComicOpts) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (stubComicsRepository) GetBySlugsAndSource(
	context.Context, comics.GetBySlugsAndSource,
) ([]comics.Comic, error) {
	return nil, nil
}

func (stubComicsRepository) Delete(context.Context, uuid.UUID) error {
	return coredomain.ErrNotFound
}

func (stubComicsRepository) GetMany(context.Context, comics.GetManyOpts) ([]comics.Comic, error) {
	return nil, nil
}

func newTestAsuraApp(t *testing.T) *asura.App {
	t.Helper()

	searchCache, err := fncache.New(
		fncache.Config[domain.SearchCacheOpts, domain.SearchCacheResult]{
			Name: "search",
			Fn: func(context.Context, domain.SearchCacheOpts) (*domain.SearchCacheResult, error) {
				return &domain.SearchCacheResult{
					Meta: domain.SearchResultMeta{Total: 1},
					Items: []domain.SearchCacheResultItem{{
						Slug:  "solo-leveling",
						Title: "Solo Leveling",
						Cover: testExternalCoverURL,
					}},
				}, nil
			},
			Key:           func(domain.SearchCacheOpts) string { return "k" },
			TTL:           time.Minute,
			ErrorTTL:      time.Minute,
			FetchTimeout:  time.Minute,
			CleanInterval: time.Minute,
			MaxEntries:    1,
		},
		fncache.Deps{Logger: slog.New(slog.DiscardHandler)},
	)
	if err != nil {
		t.Fatalf("fncache.New(search): %v", err)
	}

	infosCache, err := fncache.New(
		fncache.Config[string, domain.GetInfosBySlugResponse]{
			Name: "infos",
			Fn: func(_ context.Context, slug string) (*domain.GetInfosBySlugResponse, error) {
				return &domain.GetInfosBySlugResponse{
					Slug:  slug,
					Title: "Solo Leveling",
					Cover: testExternalCoverURL,
				}, nil
			},
			Key:           func(slug string) string { return slug },
			TTL:           time.Minute,
			ErrorTTL:      time.Minute,
			FetchTimeout:  time.Minute,
			CleanInterval: time.Minute,
			MaxEntries:    1,
		},
		fncache.Deps{Logger: slog.New(slog.DiscardHandler)},
	)
	if err != nil {
		t.Fatalf("fncache.New(infos): %v", err)
	}

	chaptersCache, err := fncache.New(
		fncache.Config[string, []domain.Chapter]{
			Name:          "chapters",
			Fn:            func(context.Context, string) (*[]domain.Chapter, error) { return nil, errors.New("n/a") },
			Key:           func(slug string) string { return slug },
			TTL:           time.Minute,
			ErrorTTL:      time.Minute,
			FetchTimeout:  time.Minute,
			CleanInterval: time.Minute,
			MaxEntries:    1,
		},
		fncache.Deps{Logger: slog.New(slog.DiscardHandler)},
	)
	if err != nil {
		t.Fatalf("fncache.New(chapters): %v", err)
	}

	imagesCache, err := fncache.New(
		fncache.Config[domain.GetImageURLsByChapterOpts, []string]{
			Name: "images",
			Fn: func(context.Context, domain.GetImageURLsByChapterOpts) (*[]string, error) {
				return nil, errors.New("n/a")
			},
			Key:           func(domain.GetImageURLsByChapterOpts) string { return "k" },
			TTL:           time.Minute,
			ErrorTTL:      time.Minute,
			FetchTimeout:  time.Minute,
			CleanInterval: time.Minute,
			MaxEntries:    1,
		},
		fncache.Deps{Logger: slog.New(slog.DiscardHandler)},
	)
	if err != nil {
		t.Fatalf("fncache.New(images): %v", err)
	}

	app, err := asura.New(asura.Config{SourceName: sources.SourceAsuraScans}, asura.Deps{
		Logger:                       slog.New(slog.DiscardHandler),
		SearchCache:                  searchCache,
		GetInfosBySlugCache:          infosCache,
		GetChaptersListBySeriesCache: chaptersCache,
		GetImageURLsByChapter:        imagesCache,
		ComicsRepository:             stubComicsRepository{},
	})
	if err != nil {
		t.Fatalf("asura.New: %v", err)
	}

	return app
}

func coverURLBuilder(source, slug string) string {
	return "/api/sources/cover/" + slug + "?source=" + source
}

func authenticatedRequest(method, target string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	user := &users.User{ID: uuid.New()}

	return req.WithContext(sessionshttp.WithAuth(req.Context(), user, sessions.Session{}))
}

func TestSearchReturnsProxiedCoverURL(t *testing.T) {
	t.Parallel()

	ctrl, err := asurahttp.New(
		asurahttp.Config{Endpoint: "/asura"},
		asurahttp.Deps{
			AsuraApp:        newTestAsuraApp(t),
			Logger:          slog.New(slog.DiscardHandler),
			CoverURLBuilder: coverURLBuilder,
		},
	)
	if err != nil {
		t.Fatalf("asurahttp.New: %v", err)
	}

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, authenticatedRequest(http.MethodGet, "/asura/search?sort=popular&order=desc"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body struct {
		Items []struct {
			Cover string `json:"cover"`
			Slug  string `json:"slug"`
		} `json:"items"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if len(body.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(body.Items))
	}

	want := coverURLBuilder(covers.SourceAsuraScans, "solo-leveling")
	if body.Items[0].Cover != want {
		t.Errorf("cover = %q, want %q", body.Items[0].Cover, want)
	}

	if body.Items[0].Cover == testExternalCoverURL {
		t.Error("external cover URL leaked in response")
	}
}

func TestGetInfosBySlugReturnsProxiedCoverURL(t *testing.T) {
	t.Parallel()

	ctrl, err := asurahttp.New(
		asurahttp.Config{Endpoint: "/asura"},
		asurahttp.Deps{
			AsuraApp:        newTestAsuraApp(t),
			Logger:          slog.New(slog.DiscardHandler),
			CoverURLBuilder: coverURLBuilder,
		},
	)
	if err != nil {
		t.Fatalf("asurahttp.New: %v", err)
	}

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, authenticatedRequest(http.MethodGet, "/asura/series/solo-leveling"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body struct {
		Cover string `json:"cover"`
		Slug  string `json:"slug"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	want := coverURLBuilder(covers.SourceAsuraScans, "solo-leveling")
	if body.Cover != want {
		t.Errorf("cover = %q, want %q", body.Cover, want)
	}
}
