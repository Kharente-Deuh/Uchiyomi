// SPDX-License-Identifier: AGPL-3.0-or-later

//nolint:goconst,lll
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
	coredomain "github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	kingofshojocore "github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/core"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/domain"
	kingofshojohttp "github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/parse"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/fncache"
)

const testExternalCoverURL = "https://external.example/cover.webp"

type searchResponseItem struct {
	Cover  string  `json:"cover"`
	Slug   string  `json:"slug"`
	Title  string  `json:"title"`
	Status string  `json:"status"`
	Rating float64 `json:"rating"`
}

type stubComicsRepository struct{}

func (stubComicsRepository) GetByID(context.Context, comics.GetByIDOpts) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (stubComicsRepository) FindByID(context.Context, uuid.UUID) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (stubComicsRepository) GetBySourceSlug(
	context.Context, comics.GetBySourceSlugOpts,
) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (stubComicsRepository) FindBySourceSlug(
	context.Context, comics.FindBySourceSlugOpts,
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

func (stubComicsRepository) GetMany(context.Context, comics.GetManyOpts) (comics.Page, error) {
	return comics.Page{}, nil
}

func (stubComicsRepository) ListByStatuses(context.Context, comics.ListByStatusesOpts) ([]comics.Comic, error) {
	return nil, nil
}

func (stubComicsRepository) UpdateStatusAndChapterCount(context.Context, comics.UpdateStatusAndChapterCountOpts) error {
	return nil
}

func (stubComicsRepository) UpdateType(context.Context, comics.UpdateTypeOpts) error {
	return nil
}

func coverURLBuilder(source, slug string) string {
	return "/api/sources/cover/" + slug + "?source=" + source
}

func authenticatedRequest(target string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	user := &users.User{ID: uuid.New()}

	return req.WithContext(sessionshttp.WithAuth(req.Context(), user, sessions.Session{}))
}

func newTestKingOfShojoApp(t *testing.T, deps kingofshojocore.Deps) *kingofshojocore.App {
	t.Helper()

	if deps.Logger == nil {
		deps.Logger = slog.New(slog.DiscardHandler)
	}

	if deps.SearchCache == nil {
		searchCache, err := fncache.New(
			fncache.Config[domain.SearchCacheOpts, domain.SearchCacheResult]{
				Name: "search",
				Fn: func(context.Context, domain.SearchCacheOpts) (*domain.SearchCacheResult, error) {
					return &domain.SearchCacheResult{
						Meta: domain.SearchResultMeta{HasNextPage: true},
						Items: []domain.SearchCacheResultItem{{
							Slug:  "solo-shojo",
							Title: "Solo Shojo",
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
			fncache.Deps{Logger: deps.Logger},
		)
		if err != nil {
			t.Fatalf("fncache.New(search): %v", err)
		}

		deps.SearchCache = searchCache
	}

	if deps.SeriesCache == nil {
		seriesCache, err := fncache.New(
			fncache.Config[string, parse.SeriesPage]{
				Name: "series",
				Fn: func(_ context.Context, slug string) (*parse.SeriesPage, error) {
					return &parse.SeriesPage{}, nil
				},
				Key:           func(slug string) string { return slug },
				TTL:           time.Minute,
				ErrorTTL:      time.Minute,
				FetchTimeout:  time.Minute,
				CleanInterval: time.Minute,
				MaxEntries:    1,
			},
			fncache.Deps{Logger: deps.Logger},
		)
		if err != nil {
			t.Fatalf("fncache.New(series): %v", err)
		}

		deps.SeriesCache = seriesCache
	}

	if deps.GetImageURLsByChapter == nil {
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
			fncache.Deps{Logger: deps.Logger},
		)
		if err != nil {
			t.Fatalf("fncache.New(images): %v", err)
		}

		deps.GetImageURLsByChapter = imagesCache
	}

	if deps.ComicsRepository == nil {
		deps.ComicsRepository = stubComicsRepository{}
	}

	app, err := kingofshojocore.New(kingofshojocore.Config{
		SourceName: sources.SourceKingOfShojo,
		BaseURL:    "https://kingofshojo.com",
	}, deps)
	if err != nil {
		t.Fatalf("kingofshojocore.New: %v", err)
	}

	return app
}

func newTestController(t *testing.T, app *kingofshojocore.App) *kingofshojohttp.Controller {
	t.Helper()

	endpoint := "/" + string(sources.SourceKingOfShojo)
	ctrl, err := kingofshojohttp.New(
		kingofshojohttp.Config{Endpoint: endpoint},
		kingofshojohttp.Deps{
			KingOfShojoApp:  app,
			Logger:          slog.New(slog.DiscardHandler),
			CoverURLBuilder: coverURLBuilder,
		},
	)
	if err != nil {
		t.Fatalf("kingofshojohttp.New: %v", err)
	}

	return ctrl
}

func TestSearchUnauthorizedWithoutSession(t *testing.T) {
	t.Parallel()

	ctrl := newTestController(t, newTestKingOfShojoApp(t, kingofshojocore.Deps{}))

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/kingofshojo/search", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestSearchBadRequestInvalidSort(t *testing.T) {
	t.Parallel()

	ctrl := newTestController(t, newTestKingOfShojoApp(t, kingofshojocore.Deps{}))

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, authenticatedRequest("/kingofshojo/search?sort=rating"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestSearchOK(t *testing.T) {
	t.Parallel()

	ctrl := newTestController(t, newTestKingOfShojoApp(t, kingofshojocore.Deps{}))

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, authenticatedRequest("/kingofshojo/search"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body struct {
		Items       []searchResponseItem `json:"items"`
		HasNextPage bool                 `json:"hasNextPage"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if !body.HasNextPage || len(body.Items) != 1 {
		t.Fatalf("hasNextPage/items = %v/%d, want true/1", body.HasNextPage, len(body.Items))
	}

	wantCover := coverURLBuilder(string(sources.SourceKingOfShojo), "solo-shojo")
	if body.Items[0].Cover != wantCover {
		t.Errorf("cover = %q, want %q", body.Items[0].Cover, wantCover)
	}

	if body.Items[0].Rating != 0 {
		t.Errorf("rating = %v, want 0", body.Items[0].Rating)
	}

	if body.Items[0].Status != "" {
		t.Errorf("status = %q, want empty", body.Items[0].Status)
	}
}

func TestSearchRejectsNonIntegerPage(t *testing.T) {
	t.Parallel()

	ctrl := newTestController(t, newTestKingOfShojoApp(t, kingofshojocore.Deps{}))
	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, authenticatedRequest("/kingofshojo/search?page=x"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestSearchClampsPageBelowOne(t *testing.T) {
	t.Parallel()

	for _, query := range []string{"", "?page=0", "?page=-1"} {
		t.Run(query, func(t *testing.T) {
			t.Parallel()

			var got domain.SearchCacheOpts

			searchCache, err := fncache.New(
				fncache.Config[domain.SearchCacheOpts, domain.SearchCacheResult]{
					Name: "search",
					Fn: func(_ context.Context, opts domain.SearchCacheOpts) (*domain.SearchCacheResult, error) {
						got = opts

						return &domain.SearchCacheResult{}, nil
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
				t.Fatalf("fncache.New: %v", err)
			}

			ctrl := newTestController(t, newTestKingOfShojoApp(t, kingofshojocore.Deps{SearchCache: searchCache}))
			r := chi.NewRouter()
			ctrl.InitRouter(r)

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, authenticatedRequest("/kingofshojo/search"+query))

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}

			if got.Page != 1 {
				t.Errorf("page = %d, want 1", got.Page)
			}
		})
	}
}

func TestGetInfosBySlugReturnsProxiedCoverURL(t *testing.T) {
	t.Parallel()

	seriesCache, err := fncache.New(
		fncache.Config[string, parse.SeriesPage]{
			Name: "series",
			Fn: func(context.Context, string) (*parse.SeriesPage, error) {
				return &parse.SeriesPage{
					Infos: parse.SeriesInfos{
						Slug:  "solo-shojo",
						Title: "Solo Shojo",
						Cover: testExternalCoverURL,
						Type:  sources.SeriesTypeManhwa,
					},
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
		t.Fatalf("fncache.New(series): %v", err)
	}

	ctrl := newTestController(t, newTestKingOfShojoApp(t, kingofshojocore.Deps{
		SeriesCache: seriesCache,
	}))

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, authenticatedRequest("/kingofshojo/series/solo-shojo"))

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

	want := coverURLBuilder(string(sources.SourceKingOfShojo), "solo-shojo")
	if body.Cover != want {
		t.Errorf("cover = %q, want %q", body.Cover, want)
	}

	if body.Cover == testExternalCoverURL {
		t.Error("external cover URL leaked in response")
	}
}

func TestSearchServiceUnavailableOnChallenge(t *testing.T) {
	t.Parallel()

	searchCache, err := fncache.New(
		fncache.Config[domain.SearchCacheOpts, domain.SearchCacheResult]{
			Name: "search",
			Fn: func(context.Context, domain.SearchCacheOpts) (*domain.SearchCacheResult, error) {
				return nil, domain.ErrChallenge
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

	ctrl := newTestController(t, newTestKingOfShojoApp(t, kingofshojocore.Deps{
		SearchCache: searchCache,
	}))

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, authenticatedRequest("/kingofshojo/search"))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestGetInfosBySlugNotFoundUnsupportedType(t *testing.T) {
	t.Parallel()

	seriesCache, err := fncache.New(
		fncache.Config[string, parse.SeriesPage]{
			Name: "series",
			Fn: func(context.Context, string) (*parse.SeriesPage, error) {
				return nil, parse.ErrUnsupportedType
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
		t.Fatalf("fncache.New(series): %v", err)
	}

	ctrl := newTestController(t, newTestKingOfShojoApp(t, kingofshojocore.Deps{
		SeriesCache: seriesCache,
	}))

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, authenticatedRequest("/kingofshojo/series/unsupported-type"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
