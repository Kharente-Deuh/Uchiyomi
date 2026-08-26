// SPDX-License-Identifier: AGPL-3.0-or-later

//nolint:goconst,lll,unparam
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
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
	coredomain "github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	asurascans "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/core"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
	asurascanshttp "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/fncache"
)

const testExternalCoverURL = "https://external.example/cover.webp"

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

type inLibraryComicsRepository struct {
	comic *comics.Comic
}

func (r inLibraryComicsRepository) GetByID(context.Context, comics.GetByIDOpts) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (r inLibraryComicsRepository) FindByID(context.Context, uuid.UUID) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (r inLibraryComicsRepository) GetBySourceSlug(
	_ context.Context, _ comics.GetBySourceSlugOpts,
) (*comics.Comic, error) {
	return r.comic, nil
}

func (r inLibraryComicsRepository) FindBySourceSlug(
	context.Context, comics.FindBySourceSlugOpts,
) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (r inLibraryComicsRepository) Create(context.Context, comics.CreateComicOpts) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (r inLibraryComicsRepository) GetBySlugsAndSource(
	context.Context, comics.GetBySlugsAndSource,
) ([]comics.Comic, error) {
	return nil, nil
}

func (r inLibraryComicsRepository) Delete(context.Context, uuid.UUID) error {
	return coredomain.ErrNotFound
}

func (r inLibraryComicsRepository) GetMany(context.Context, comics.GetManyOpts) (comics.Page, error) {
	return comics.Page{}, nil
}

func (r inLibraryComicsRepository) ListByStatuses(context.Context, comics.ListByStatusesOpts) ([]comics.Comic, error) {
	return nil, nil
}

func (r inLibraryComicsRepository) UpdateStatusAndChapterCount(context.Context, comics.UpdateStatusAndChapterCountOpts) error {
	return nil
}

type libraryChaptersService struct {
	chapters []chapters.Chapter
}

func (s libraryChaptersService) CreateAll(
	context.Context, uuid.UUID, []sources.SourceChapter,
) ([]chapters.Chapter, error) {
	return nil, nil
}

func (s libraryChaptersService) ListByComicID(
	context.Context, uuid.UUID,
) ([]chapters.Chapter, error) {
	return s.chapters, nil
}

func (s libraryChaptersService) EnqueueDownloadable(context.Context, []chapters.Chapter) error {
	return nil
}

func (s libraryChaptersService) EnqueueResumable(context.Context) error {
	return nil
}

func (s libraryChaptersService) ScanEarlyAccess(context.Context) error {
	return nil
}

func (s libraryChaptersService) CleanupComic(context.Context, uuid.UUID, []chapters.Chapter) error {
	return nil
}

func (s libraryChaptersService) RetryDownload(context.Context, chapters.RetryDownloadOpts) error {
	return nil
}

func (s libraryChaptersService) GetByIds(context.Context, chapters.GetByIdsOpts) ([]chapters.Chapter, error) {
	return nil, nil
}

func (s libraryChaptersService) ListForLibrary(context.Context, chapters.ListForLibraryOpts) ([]chapters.Chapter, error) {
	return nil, nil
}

func (s libraryChaptersService) GetForLibrary(context.Context, chapters.GetForLibraryOpts) (*chapters.Chapter, error) {
	return nil, nil
}

func (s libraryChaptersService) GetDetailForLibrary(
	context.Context, chapters.GetForLibraryOpts,
) (*chapters.ChapterDetail, error) {
	return nil, nil
}

func (s libraryChaptersService) ServePage(context.Context, chapters.ServePageOpts) (string, string, error) {
	return "", "", nil
}

func newTestAsuraScansApp(t *testing.T, deps ...asurascans.Deps) *asurascans.App {
	t.Helper()

	var partial asurascans.Deps
	if len(deps) > 0 {
		partial = deps[0]
	}

	searchCache := partial.SearchCache
	if searchCache == nil {
		var err error

		searchCache, err = fncache.New(
			fncache.Config[domain.SearchCacheOpts, domain.SearchCacheResult]{
				Name: "search",
				Fn: func(context.Context, domain.SearchCacheOpts) (*domain.SearchCacheResult, error) {
					return &domain.SearchCacheResult{
						Meta: domain.SearchResultMeta{HasNextPage: true},
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
	}

	infosCache := partial.GetInfosBySlugCache
	if infosCache == nil {
		var err error

		infosCache, err = fncache.New(
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
	}

	chaptersCache := partial.GetChaptersListBySeriesCache
	if chaptersCache == nil {
		var err error

		chaptersCache, err = fncache.New(
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
	}

	imagesCache := partial.GetImageURLsByChapter
	if imagesCache == nil {
		var err error

		imagesCache, err = fncache.New(
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
	}

	comicsRepo := partial.ComicsRepository
	if comicsRepo == nil {
		comicsRepo = stubComicsRepository{}
	}

	logger := partial.Logger
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	app, err := asurascans.New(asurascans.Config{SourceName: sources.SourceAsuraScans}, asurascans.Deps{
		Logger:                       logger,
		SearchCache:                  searchCache,
		GetInfosBySlugCache:          infosCache,
		GetChaptersListBySeriesCache: chaptersCache,
		GetImageURLsByChapter:        imagesCache,
		ComicsRepository:             comicsRepo,
		ChaptersService:              partial.ChaptersService,
	})
	if err != nil {
		t.Fatalf("asurascans.New: %v", err)
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

func TestSearchRejectsNonIntegerPage(t *testing.T) {
	t.Parallel()

	endpoint := "/" + string(sources.SourceAsuraScans)
	ctrl, err := asurascanshttp.New(
		asurascanshttp.Config{Endpoint: endpoint},
		asurascanshttp.Deps{
			AsuraScansApp:   newTestAsuraScansApp(t),
			Logger:          slog.New(slog.DiscardHandler),
			CoverURLBuilder: coverURLBuilder,
		},
	)
	if err != nil {
		t.Fatalf("asurascanshttp.New: %v", err)
	}

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, authenticatedRequest(http.MethodGet, endpoint+"/search?page=x"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestSearchClampsPageBelowOne(t *testing.T) {
	t.Parallel()

	endpoint := "/" + string(sources.SourceAsuraScans)

	for _, query := range []string{"?sort=popular&order=desc", "?sort=popular&order=desc&page=0", "?sort=popular&order=desc&page=-1"} {
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

			ctrl, err := asurascanshttp.New(
				asurascanshttp.Config{Endpoint: endpoint},
				asurascanshttp.Deps{
					AsuraScansApp:   newTestAsuraScansApp(t, asurascans.Deps{SearchCache: searchCache}),
					Logger:          slog.New(slog.DiscardHandler),
					CoverURLBuilder: coverURLBuilder,
				},
			)
			if err != nil {
				t.Fatalf("asurascanshttp.New: %v", err)
			}

			r := chi.NewRouter()
			ctrl.InitRouter(r)

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, authenticatedRequest(http.MethodGet, endpoint+"/search"+query))

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}

			if got.Page != 1 {
				t.Errorf("page = %d, want 1", got.Page)
			}
		})
	}
}

func TestSearchReturnsProxiedCoverURL(t *testing.T) {
	t.Parallel()

	endpoint := "/" + string(sources.SourceAsuraScans)
	ctrl, err := asurascanshttp.New(
		asurascanshttp.Config{Endpoint: endpoint},
		asurascanshttp.Deps{
			AsuraScansApp:   newTestAsuraScansApp(t),
			Logger:          slog.New(slog.DiscardHandler),
			CoverURLBuilder: coverURLBuilder,
		},
	)
	if err != nil {
		t.Fatalf("asurascanshttp.New: %v", err)
	}

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, authenticatedRequest(http.MethodGet, endpoint+"/search?sort=popular&order=desc"))

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

	endpoint := "/" + string(sources.SourceAsuraScans)
	ctrl, err := asurascanshttp.New(
		asurascanshttp.Config{Endpoint: endpoint},
		asurascanshttp.Deps{
			AsuraScansApp:   newTestAsuraScansApp(t),
			Logger:          slog.New(slog.DiscardHandler),
			CoverURLBuilder: coverURLBuilder,
		},
	)
	if err != nil {
		t.Fatalf("asurascanshttp.New: %v", err)
	}

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, authenticatedRequest(http.MethodGet, endpoint+"/series/solo-leveling"))

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

func newChaptersTestAsuraScansApp(t *testing.T, deps asurascans.Deps) *asurascans.App {
	t.Helper()

	if deps.Logger == nil {
		deps.Logger = slog.New(slog.DiscardHandler)
	}

	if deps.SearchCache == nil {
		searchCache, err := fncache.New(
			fncache.Config[domain.SearchCacheOpts, domain.SearchCacheResult]{
				Name: "search",
				Fn: func(context.Context, domain.SearchCacheOpts) (*domain.SearchCacheResult, error) {
					return &domain.SearchCacheResult{}, nil
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

	if deps.GetInfosBySlugCache == nil {
		infosCache, err := fncache.New(
			fncache.Config[string, domain.GetInfosBySlugResponse]{
				Name: "infos",
				Fn: func(context.Context, string) (*domain.GetInfosBySlugResponse, error) {
					return &domain.GetInfosBySlugResponse{}, nil
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
			t.Fatalf("fncache.New(infos): %v", err)
		}

		deps.GetInfosBySlugCache = infosCache
	}

	if deps.GetChaptersListBySeriesCache == nil {
		chaptersCache, err := fncache.New(
			fncache.Config[string, []domain.Chapter]{
				Name: "chapters",
				Fn: func(context.Context, string) (*[]domain.Chapter, error) {
					return &[]domain.Chapter{}, nil
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
			t.Fatalf("fncache.New(chapters): %v", err)
		}

		deps.GetChaptersListBySeriesCache = chaptersCache
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

	app, err := asurascans.New(asurascans.Config{SourceName: sources.SourceAsuraScans}, deps)
	if err != nil {
		t.Fatalf("asurascans.New: %v", err)
	}

	return app
}

func TestGetChaptersListBySeriesEnrichesLibraryChapters(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chapterID := uuid.New()
	sourceChapterSlug := "solo-leveling-chapter-1"
	download := 75

	chaptersCache, err := fncache.New(
		fncache.Config[string, []domain.Chapter]{
			Name: "chapters",
			Fn: func(context.Context, string) (*[]domain.Chapter, error) {
				chapterList := []domain.Chapter{{
					ID:     sourceChapterSlug,
					Number: 1,
					Title:  "Chapter 1",
				}}

				return &chapterList, nil
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
		t.Fatalf("fncache.New(chapters): %v", err)
	}

	endpoint := "/" + string(sources.SourceAsuraScans)
	ctrl, err := asurascanshttp.New(
		asurascanshttp.Config{Endpoint: endpoint},
		asurascanshttp.Deps{
			AsuraScansApp: newChaptersTestAsuraScansApp(t, asurascans.Deps{
				GetChaptersListBySeriesCache: chaptersCache,
				ComicsRepository: inLibraryComicsRepository{
					comic: &comics.Comic{ID: comicID, Slug: "solo-leveling"},
				},
				ChaptersService: libraryChaptersService{chapters: []chapters.Chapter{{
					ID:                chapterID,
					ComicID:           comicID,
					SourceChapterSlug: sourceChapterSlug,
					Download:          download,
				}}},
			}),
			Logger:          slog.New(slog.DiscardHandler),
			CoverURLBuilder: coverURLBuilder,
		},
	)
	if err != nil {
		t.Fatalf("asurascanshttp.New: %v", err)
	}

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, authenticatedRequest(http.MethodGet, endpoint+"/series/solo-leveling/chapters"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body []struct {
		InternalID *string `json:"internalId"`
		Download   *int    `json:"download"`
		ID         string  `json:"id"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if len(body) != 1 {
		t.Fatalf("len(chapters) = %d, want 1", len(body))
	}

	if body[0].ID != sourceChapterSlug {
		t.Errorf("id = %q, want %q", body[0].ID, sourceChapterSlug)
	}

	if body[0].InternalID == nil || *body[0].InternalID != chapterID.String() {
		t.Errorf("internalId = %v, want %q", body[0].InternalID, chapterID)
	}

	if body[0].Download == nil || *body[0].Download != download {
		t.Errorf("download = %v, want %d", body[0].Download, download)
	}
}

func TestGetChaptersListBySeriesWithoutLibraryEntry(t *testing.T) {
	t.Parallel()

	chaptersCache, err := fncache.New(
		fncache.Config[string, []domain.Chapter]{
			Name: "chapters",
			Fn: func(context.Context, string) (*[]domain.Chapter, error) {
				chapterList := []domain.Chapter{{ID: "c1", Number: 1, Title: "Chapter 1"}}

				return &chapterList, nil
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
		t.Fatalf("fncache.New(chapters): %v", err)
	}

	endpoint := "/" + string(sources.SourceAsuraScans)
	ctrl, err := asurascanshttp.New(
		asurascanshttp.Config{Endpoint: endpoint},
		asurascanshttp.Deps{
			AsuraScansApp: newChaptersTestAsuraScansApp(t, asurascans.Deps{
				GetChaptersListBySeriesCache: chaptersCache,
				ComicsRepository:             stubComicsRepository{},
			}),
			Logger:          slog.New(slog.DiscardHandler),
			CoverURLBuilder: coverURLBuilder,
		},
	)
	if err != nil {
		t.Fatalf("asurascanshttp.New: %v", err)
	}

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, authenticatedRequest(http.MethodGet, endpoint+"/series/solo-leveling/chapters"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body []struct {
		InternalID *string `json:"internalId"`
		Download   *int    `json:"download"`
		ID         string  `json:"id"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if len(body) != 1 {
		t.Fatalf("len(chapters) = %d, want 1", len(body))
	}

	if body[0].ID != "c1" {
		t.Errorf("id = %q, want %q", body[0].ID, "c1")
	}

	if body[0].InternalID != nil {
		t.Errorf("internalId = %v, want nil", body[0].InternalID)
	}

	if body[0].Download != nil {
		t.Errorf("download = %v, want nil", body[0].Download)
	}
}

func TestGetImageURLsByChapter(t *testing.T) {
	t.Parallel()

	imagesCache, err := fncache.New(
		fncache.Config[domain.GetImageURLsByChapterOpts, []string]{
			Name: "images",
			Fn: func(context.Context, domain.GetImageURLsByChapterOpts) (*[]string, error) {
				urls := []string{"https://cdn.example/1.webp", "https://cdn.example/2.webp"}

				return &urls, nil
			},
			Key:           domain.GetImageURLsByChapterOpts.CacheKey,
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

	endpoint := "/" + string(sources.SourceAsuraScans)
	ctrl, err := asurascanshttp.New(
		asurascanshttp.Config{Endpoint: endpoint},
		asurascanshttp.Deps{
			AsuraScansApp: newChaptersTestAsuraScansApp(t, asurascans.Deps{
				GetImageURLsByChapter: imagesCache,
			}),
			Logger:          slog.New(slog.DiscardHandler),
			CoverURLBuilder: coverURLBuilder,
		},
	)
	if err != nil {
		t.Fatalf("asurascanshttp.New: %v", err)
	}

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, authenticatedRequest(
		http.MethodGet,
		endpoint+"/series/solo-leveling/chapters/chapter-1",
	))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s, want %d", rec.Code, rec.Body.String(), http.StatusOK)
	}

	var body struct {
		URLs []string `json:"urls"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if len(body.URLs) != 2 || body.URLs[0] != "https://cdn.example/1.webp" {
		t.Errorf("urls = %v", body.URLs)
	}
}

func TestGetImageURLsByChapterNotFound(t *testing.T) {
	t.Parallel()

	imagesCache, err := fncache.New(
		fncache.Config[domain.GetImageURLsByChapterOpts, []string]{
			Name: "images",
			Fn: func(context.Context, domain.GetImageURLsByChapterOpts) (*[]string, error) {
				return nil, domain.ErrNotFound
			},
			Key:           domain.GetImageURLsByChapterOpts.CacheKey,
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

	endpoint := "/" + string(sources.SourceAsuraScans)
	ctrl, err := asurascanshttp.New(
		asurascanshttp.Config{Endpoint: endpoint},
		asurascanshttp.Deps{
			AsuraScansApp: newChaptersTestAsuraScansApp(t, asurascans.Deps{
				GetImageURLsByChapter: imagesCache,
			}),
			Logger:          slog.New(slog.DiscardHandler),
			CoverURLBuilder: coverURLBuilder,
		},
	)
	if err != nil {
		t.Fatalf("asurascanshttp.New: %v", err)
	}

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, authenticatedRequest(
		http.MethodGet,
		endpoint+"/series/solo-leveling/chapters/missing",
	))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestSearchNotMountedAtLegacyAsuraPath(t *testing.T) {
	t.Parallel()

	endpoint := "/" + string(sources.SourceAsuraScans)
	ctrl, err := asurascanshttp.New(
		asurascanshttp.Config{Endpoint: endpoint},
		asurascanshttp.Deps{
			AsuraScansApp:   newTestAsuraScansApp(t),
			Logger:          slog.New(slog.DiscardHandler),
			CoverURLBuilder: coverURLBuilder,
		},
	)
	if err != nil {
		t.Fatalf("asurascanshttp.New: %v", err)
	}

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, authenticatedRequest(http.MethodGet, "/asura/search"))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
