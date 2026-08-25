// SPDX-License-Identifier: AGPL-3.0-or-later

//nolint:goconst,lll
package core_test

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	coredomain "github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/core"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/parse"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/fncache"
)

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

type libraryComicsRepository struct {
	err    error
	comics []comics.Comic
}

func (r libraryComicsRepository) GetByID(context.Context, comics.GetByIDOpts) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (r libraryComicsRepository) FindByID(context.Context, uuid.UUID) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (r libraryComicsRepository) GetBySourceSlug(
	context.Context, comics.GetBySourceSlugOpts,
) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (r libraryComicsRepository) FindBySourceSlug(
	context.Context, comics.FindBySourceSlugOpts,
) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (r libraryComicsRepository) Create(context.Context, comics.CreateComicOpts) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (r libraryComicsRepository) GetBySlugsAndSource(
	context.Context, comics.GetBySlugsAndSource,
) ([]comics.Comic, error) {
	return r.comics, r.err
}

func (r libraryComicsRepository) Delete(context.Context, uuid.UUID) error {
	return coredomain.ErrNotFound
}

func (r libraryComicsRepository) GetMany(context.Context, comics.GetManyOpts) (comics.Page, error) {
	return comics.Page{}, nil
}

func (r libraryComicsRepository) ListByStatuses(context.Context, comics.ListByStatusesOpts) ([]comics.Comic, error) {
	return nil, nil
}

func (r libraryComicsRepository) UpdateStatusAndChapterCount(context.Context, comics.UpdateStatusAndChapterCountOpts) error {
	return nil
}

type inLibraryComicsRepository struct {
	comic *comics.Comic
	err   error
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
	if r.err != nil {
		return nil, r.err
	}

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
	err      error
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
	return s.chapters, s.err
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

func testConfig() core.Config {
	return core.Config{
		SourceName: sources.SourceKingOfShojo,
		BaseURL:    "https://kingofshojo.com",
	}
}

func defaultSeriesPage(slug string) *parse.SeriesPage {
	return &parse.SeriesPage{
		Infos: parse.SeriesInfos{
			Slug: slug,
			Type: sources.SeriesTypeMangatoon,
		},
		Chapters: []parse.SeriesChapter{{ID: "c1", Number: 1}},
	}
}

func fullDeps(t *testing.T) core.Deps {
	t.Helper()

	return core.Deps{
		Logger: slog.New(slog.DiscardHandler),
		SearchCache: newCache(t, func(opts domain.SearchCacheOpts) string { return opts.CacheKey() },
			func(context.Context, domain.SearchCacheOpts) (*domain.SearchCacheResult, error) {
				return &domain.SearchCacheResult{}, nil
			}),
		SeriesCache: newCache(t, identityKey,
			func(_ context.Context, slug string) (*parse.SeriesPage, error) {
				return defaultSeriesPage(slug), nil
			}),
		GetImageURLsByChapter: newCache(t, domain.GetImageURLsByChapterOpts.CacheKey,
			func(context.Context, domain.GetImageURLsByChapterOpts) (*[]string, error) {
				urls := []string{"https://example.test/1.jpg"}

				return &urls, nil
			}),
		ComicsRepository: stubComicsRepository{},
	}
}

func TestDepsValidate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		drop    func(*core.Deps)
		wantErr string
	}{
		"complet": {drop: func(*core.Deps) {}},
		"without logger": {
			drop:    func(d *core.Deps) { d.Logger = nil },
			wantErr: "logger is required",
		},
		"without searchCache": {
			drop:    func(d *core.Deps) { d.SearchCache = nil },
			wantErr: "searchCache is required",
		},
		"without seriesCache": {
			drop:    func(d *core.Deps) { d.SeriesCache = nil },
			wantErr: "seriesCache is required",
		},
		"without getImageURLsByChapter": {
			drop:    func(d *core.Deps) { d.GetImageURLsByChapter = nil },
			wantErr: "getImageURLsByChapter is required",
		},
		"without comicsRepository": {
			drop:    func(d *core.Deps) { d.ComicsRepository = nil },
			wantErr: "comicsRepository is required",
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

	app, err := core.New(testConfig(), core.Deps{})
	if err == nil {
		t.Fatal("New with empty dependencies must fail")
	}

	if app != nil {
		t.Error("New returned an App in addition to the error")
	}

	if want := "deps.Validate: logger is required"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
}

func TestNewSucceeds(t *testing.T) {
	t.Parallel()

	app, err := core.New(testConfig(), fullDeps(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if app == nil {
		t.Fatal("New returned nil App without error")
	}
}

func TestAppSearchDelegatesToCache(t *testing.T) {
	t.Parallel()

	called := make(chan domain.SearchCacheOpts, 1)

	deps := fullDeps(t)
	deps.SearchCache = newCache(t, func(opts domain.SearchCacheOpts) string { return opts.CacheKey() },
		func(_ context.Context, opts domain.SearchCacheOpts) (*domain.SearchCacheResult, error) {
			called <- opts

			return &domain.SearchCacheResult{Meta: domain.SearchResultMeta{Total: 3}}, nil
		})

	app, err := core.New(testConfig(), deps)
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

	app, err := core.New(testConfig(), fullDeps(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	first, err := app.GetChaptersListBySeries(context.Background(), "slug", uuid.Nil)
	if err != nil {
		t.Fatalf("GetChaptersListBySeries: %v", err)
	}

	if len(first) == 0 {
		t.Fatal("test cache must return at least one chapter")
	}

	first[0].ID = "corrompu"

	second, err := app.GetChaptersListBySeries(context.Background(), "slug", uuid.Nil)
	if err != nil {
		t.Fatalf("GetChaptersListBySeries (2nd call): %v", err)
	}

	if second[0].ID != "c1" {
		t.Errorf("second[0].ID = %q, want \"c1\": caller could modify cache entry", second[0].ID)
	}
}

func TestAppGetImageURLsByChapterReturnsACopy(t *testing.T) {
	t.Parallel()

	app, err := core.New(testConfig(), fullDeps(t))
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

func TestAppGetInfosBySlugWithoutLibraryEntry(t *testing.T) {
	t.Parallel()

	app, err := core.New(testConfig(), fullDeps(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := app.GetInfosBySlug(context.Background(), sources.GetInfosBySlugOpts{
		Slug:   "nano-machine",
		UserID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("GetInfosBySlug: %v", err)
	}

	if res.InternalID != nil {
		t.Errorf("InternalID = %v, want nil when comic is not in library", res.InternalID)
	}
}

func TestAppGetInfosBySlugFreshBypassesCache(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	deps := fullDeps(t)
	deps.SeriesCache = newCache(t, identityKey,
		func(context.Context, string) (*parse.SeriesPage, error) {
			calls.Add(1)

			return &parse.SeriesPage{Infos: parse.SeriesInfos{Title: "Nano"}}, nil
		})

	app, err := core.New(testConfig(), deps)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	opts := sources.GetInfosBySlugOpts{Slug: "nano-machine"}
	if _, err := app.GetInfosBySlug(context.Background(), opts); err != nil {
		t.Fatalf("GetInfosBySlug: %v", err)
	}

	opts.Fresh = true
	if _, err := app.GetInfosBySlug(context.Background(), opts); err != nil {
		t.Fatalf("GetInfosBySlug fresh: %v", err)
	}

	if got := calls.Load(); got != 2 {
		t.Errorf("cache Fn called %d times, want 2 (cached Get then Fresh Fetch)", got)
	}
}

func TestAppGetInfosAndChaptersShareOneSeriesFetch(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	slug := "solo-leveling"

	deps := fullDeps(t)
	deps.SeriesCache = newCache(t, identityKey,
		func(_ context.Context, s string) (*parse.SeriesPage, error) {
			calls.Add(1)

			return &parse.SeriesPage{
				Infos: parse.SeriesInfos{
					Slug: s,
					Type: sources.SeriesTypeMangatoon,
				},
				Chapters: []parse.SeriesChapter{{
					ID:     "solo-leveling-chapter-1",
					Number: 1,
				}},
			}, nil
		})

	app, err := core.New(testConfig(), deps)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx := context.Background()
	if _, err := app.GetInfosBySlug(ctx, sources.GetInfosBySlugOpts{Slug: slug}); err != nil {
		t.Fatalf("GetInfosBySlug: %v", err)
	}

	if _, err := app.GetChaptersBySlug(ctx, sources.GetChaptersBySlugOpts{Slug: slug}); err != nil {
		t.Fatalf("GetChaptersBySlug: %v", err)
	}

	if got := calls.Load(); got != 1 {
		t.Errorf("series cache Fn called %d times, want 1 shared fetch", got)
	}
}

func TestAppGetInfosBySlugMangaTypeIsMangatoon(t *testing.T) {
	t.Parallel()

	slug := "some-manga"
	deps := fullDeps(t)
	deps.SeriesCache = newCache(t, identityKey,
		func(_ context.Context, s string) (*parse.SeriesPage, error) {
			return &parse.SeriesPage{
				Infos: parse.SeriesInfos{
					Slug:  s,
					Type:  sources.SeriesTypeMangatoon,
					Title: "Some Manga",
				},
			}, nil
		})

	app, err := core.New(testConfig(), deps)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := app.GetInfosBySlug(context.Background(), sources.GetInfosBySlugOpts{Slug: slug})
	if err != nil {
		t.Fatalf("GetInfosBySlug: %v", err)
	}

	if res.Type != sources.SeriesTypeMangatoon {
		t.Errorf("Type = %q, want %q", res.Type, sources.SeriesTypeMangatoon)
	}
}

func TestAppSearchAttachesLibraryInternalID(t *testing.T) {
	t.Parallel()

	libraryID := uuid.New()
	inLibrarySlug := "solo-leveling"
	otherSlug := "one-piece"

	deps := fullDeps(t)
	deps.SearchCache = newCache(t, func(opts domain.SearchCacheOpts) string { return opts.CacheKey() },
		func(context.Context, domain.SearchCacheOpts) (*domain.SearchCacheResult, error) {
			return &domain.SearchCacheResult{
				Items: []domain.SearchCacheResultItem{
					{Slug: inLibrarySlug, Title: "Solo Leveling"},
					{Slug: otherSlug, Title: "One Piece"},
				},
			}, nil
		})
	deps.ComicsRepository = libraryComicsRepository{comics: []comics.Comic{{
		ID:   libraryID,
		Slug: inLibrarySlug,
	}}}

	app, err := core.New(testConfig(), deps)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := app.Search(context.Background(), domain.SearchOpts{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(res.Items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(res.Items))
	}

	if res.Items[0].InternalID == nil || *res.Items[0].InternalID != libraryID {
		t.Errorf("solo-leveling InternalID = %v, want %s", res.Items[0].InternalID, libraryID)
	}

	if res.Items[1].InternalID != nil {
		t.Errorf("one-piece InternalID = %v, want nil", res.Items[1].InternalID)
	}
}

func TestAppSearchLibraryLookupFailure(t *testing.T) {
	t.Parallel()

	deps := fullDeps(t)
	deps.SearchCache = newCache(t, func(opts domain.SearchCacheOpts) string { return opts.CacheKey() },
		func(context.Context, domain.SearchCacheOpts) (*domain.SearchCacheResult, error) {
			return &domain.SearchCacheResult{
				Items: []domain.SearchCacheResultItem{{Slug: "solo-leveling"}},
			}, nil
		})
	deps.ComicsRepository = libraryComicsRepository{err: errors.New("db down")}

	app, err := core.New(testConfig(), deps)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = app.Search(context.Background(), domain.SearchOpts{})
	if err == nil {
		t.Fatal("Search must fail when the library lookup fails")
	}
}

func TestAppGetChaptersListBySeriesEnrichesLibraryChapters(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chapterID := uuid.New()
	userID := uuid.New()
	seriesSlug := "solo-leveling"
	sourceChapterSlug := "solo-leveling-chapter-1"
	download := 42

	deps := fullDeps(t)
	deps.SeriesCache = newCache(t, identityKey,
		func(_ context.Context, slug string) (*parse.SeriesPage, error) {
			return &parse.SeriesPage{
				Infos: parse.SeriesInfos{Slug: slug},
				Chapters: []parse.SeriesChapter{{
					ID:     sourceChapterSlug,
					Number: 1,
					Title:  "Chapter 1",
				}, {
					ID:     "unknown-chapter",
					Number: 2,
					Title:  "Chapter 2",
				}},
			}, nil
		})
	deps.ComicsRepository = inLibraryComicsRepository{
		comic: &comics.Comic{ID: comicID, Slug: seriesSlug},
	}
	deps.ChaptersService = libraryChaptersService{chapters: []chapters.Chapter{{
		ID:                chapterID,
		ComicID:           comicID,
		SourceChapterSlug: sourceChapterSlug,
		Download:          download,
	}}}

	app, err := core.New(testConfig(), deps)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := app.GetChaptersListBySeries(context.Background(), seriesSlug, userID)
	if err != nil {
		t.Fatalf("GetChaptersListBySeries: %v", err)
	}

	if len(res) != 2 {
		t.Fatalf("len(chapters) = %d, want 2", len(res))
	}

	if res[0].InternalID == nil || *res[0].InternalID != chapterID {
		t.Errorf("chapter[0].InternalID = %v, want %s", res[0].InternalID, chapterID)
	}

	if res[0].Download == nil || *res[0].Download != download {
		t.Errorf("chapter[0].Download = %v, want %d", res[0].Download, download)
	}

	if res[1].InternalID != nil {
		t.Errorf("chapter[1].InternalID = %v, want nil", res[1].InternalID)
	}

	if res[1].Download != nil {
		t.Errorf("chapter[1].Download = %v, want nil", res[1].Download)
	}
}

func TestAppGetChaptersListBySeriesWithoutLibraryEntry(t *testing.T) {
	t.Parallel()

	deps := fullDeps(t)
	deps.SeriesCache = newCache(t, identityKey,
		func(_ context.Context, slug string) (*parse.SeriesPage, error) {
			return &parse.SeriesPage{
				Infos:    parse.SeriesInfos{Slug: slug},
				Chapters: []parse.SeriesChapter{{ID: "c1", Number: 1}},
			}, nil
		})

	app, err := core.New(testConfig(), deps)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := app.GetChaptersListBySeries(context.Background(), "solo-leveling", uuid.New())
	if err != nil {
		t.Fatalf("GetChaptersListBySeries: %v", err)
	}

	if len(res) != 1 {
		t.Fatalf("len(chapters) = %d, want 1", len(res))
	}

	if res[0].InternalID != nil {
		t.Errorf("InternalID = %v, want nil when comic is not in library", res[0].InternalID)
	}

	if res[0].Download != nil {
		t.Errorf("Download = %v, want nil when comic is not in library", res[0].Download)
	}
}

func TestAppRunReturnsOnContextCancel(t *testing.T) {
	t.Parallel()

	app, err := core.New(testConfig(), fullDeps(t))
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

func TestAppGetChaptersBySlug(t *testing.T) {
	t.Parallel()

	publishedAt := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	deps := fullDeps(t)
	deps.SeriesCache = newCache(t, identityKey,
		func(_ context.Context, slug string) (*parse.SeriesPage, error) {
			return &parse.SeriesPage{
				Infos: parse.SeriesInfos{Slug: slug},
				Chapters: []parse.SeriesChapter{{
					ID:          "chapter-1",
					Title:       "Chapter 1",
					Number:      1,
					PageCount:   20,
					PublishedAt: publishedAt,
				}},
			}, nil
		})

	app, err := core.New(testConfig(), deps)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	got, err := app.GetChaptersBySlug(context.Background(), sources.GetChaptersBySlugOpts{Slug: "solo-leveling"})
	if err != nil {
		t.Fatalf("GetChaptersBySlug: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}

	if got[0].SourceChapterSlug != "chapter-1" || got[0].Title != "Chapter 1" || got[0].PageCount != 20 {
		t.Errorf("GetChaptersBySlug() = %+v", got[0])
	}

	if !got[0].PublishedAt.Equal(publishedAt) {
		t.Errorf("PublishedAt = %v, want %v", got[0].PublishedAt, publishedAt)
	}
}

func TestAppGetPageURLsByChapter(t *testing.T) {
	t.Parallel()

	app, err := core.New(testConfig(), fullDeps(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	got, err := app.GetPageURLsByChapter(context.Background(), sources.GetPageURLsByChapterOpts{
		SeriesSlug:  "solo-leveling",
		ChapterSlug: "c1",
	})
	if err != nil {
		t.Fatalf("GetPageURLsByChapter: %v", err)
	}

	if len(got) != 1 || got[0] != "https://example.test/1.jpg" {
		t.Errorf("GetPageURLsByChapter() = %v", got)
	}
}

func TestAppBindChaptersServiceEnrichesLibraryChapters(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chapterID := uuid.New()
	userID := uuid.New()
	seriesSlug := "solo-leveling"
	sourceChapterSlug := "solo-leveling-chapter-1"

	deps := fullDeps(t)
	deps.SeriesCache = newCache(t, identityKey,
		func(_ context.Context, slug string) (*parse.SeriesPage, error) {
			return &parse.SeriesPage{
				Infos:    parse.SeriesInfos{Slug: slug},
				Chapters: []parse.SeriesChapter{{ID: sourceChapterSlug, Number: 1, Title: "Chapter 1"}},
			}, nil
		})
	deps.ComicsRepository = inLibraryComicsRepository{
		comic: &comics.Comic{ID: comicID, Slug: seriesSlug},
	}

	app, err := core.New(testConfig(), deps)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	before, err := app.GetChaptersListBySeries(context.Background(), seriesSlug, userID)
	if err != nil {
		t.Fatalf("GetChaptersListBySeries before bind: %v", err)
	}

	if before[0].InternalID != nil {
		t.Errorf("InternalID before bind = %v, want nil", before[0].InternalID)
	}

	app.BindChaptersService(libraryChaptersService{chapters: []chapters.Chapter{{
		ID:                chapterID,
		ComicID:           comicID,
		SourceChapterSlug: sourceChapterSlug,
		Download:          7,
	}}})

	after, err := app.GetChaptersListBySeries(context.Background(), seriesSlug, userID)
	if err != nil {
		t.Fatalf("GetChaptersListBySeries after bind: %v", err)
	}

	if after[0].InternalID == nil || *after[0].InternalID != chapterID {
		t.Errorf("InternalID after bind = %v, want %s", after[0].InternalID, chapterID)
	}
}
