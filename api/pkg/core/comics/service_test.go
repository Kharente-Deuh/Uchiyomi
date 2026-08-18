// SPDX-License-Identifier: AGPL-3.0-or-later

//nolint:goconst
package comics_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/library"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
)

const (
	testSlug         = "solo-leveling"
	testSource       = sources.SourceAsuraScans
	testChapterSlug  = "chapter-1"
	testChapter2Slug = "chapter-2"
	testChapterTitle = "Chapter 1"
)

type fakeTransactor struct {
	err   error
	calls int
}

func (f *fakeTransactor) WithinTx(ctx context.Context, _ transaction.TxOpts, fn func(context.Context) error) error {
	f.calls++

	if err := fn(ctx); err != nil {
		return err
	}

	return f.err
}

type fakeLocalCoverStore struct {
	obtainErr  error
	serveErr   error
	removeErr  error
	diskPath   string
	mime       string
	lastObtain struct {
		source string
		slug   string
		id     uuid.UUID
	}
	obtainCalls int
	removeCalls int
	serveCalls  int
	lastRemove  uuid.UUID
}

func (f *fakeLocalCoverStore) ObtainLocal(_ context.Context, comicID uuid.UUID, source, slug string) error {
	f.obtainCalls++
	f.lastObtain.id = comicID
	f.lastObtain.source = source
	f.lastObtain.slug = slug

	return f.obtainErr
}

func (f *fakeLocalCoverStore) ServeLocal(context.Context, uuid.UUID) (string, string, error) {
	f.serveCalls++

	return f.diskPath, f.mime, f.serveErr
}

func (f *fakeLocalCoverStore) RemoveLocal(comicID uuid.UUID) error {
	f.removeCalls++
	f.lastRemove = comicID

	return f.removeErr
}

type fakeComicsRepository struct {
	getByIDErr             error
	findBySourceSlugErr    error
	deleteErr              error
	createErr              error
	getManyErr             error
	listByStatusesErr      error
	updateStatusErr        error
	getByIDResult          *comics.Comic
	findBySourceSlugResult *comics.Comic
	findBySourceSlugSecond *comics.Comic
	createResult           *comics.Comic
	lastListByStatuses     comics.ListByStatusesOpts
	listByStatusesResult   []comics.Comic
	getManyResult          comics.Page
	lastUpdateStatus       comics.UpdateStatusAndChapterCountOpts
	lastCreateOpts         comics.CreateComicOpts
	findBySourceSlugCalls  int
	createCalls            int
	getByIDCalls           int
	getManyCalls           int
	deleteCalls            int
	listByStatusesCalls    int
	updateStatusCalls      int
	lastDeleteID           uuid.UUID
}

func (f *fakeComicsRepository) GetByID(_ context.Context, _ comics.GetByIDOpts) (*comics.Comic, error) {
	f.getByIDCalls++

	return f.getByIDResult, f.getByIDErr
}

func (f *fakeComicsRepository) FindByID(context.Context, uuid.UUID) (*comics.Comic, error) {
	panic("FindByID must not be called by the comics service")
}

func (f *fakeComicsRepository) GetBySourceSlug(context.Context, comics.GetBySourceSlugOpts) (*comics.Comic, error) {
	panic("GetBySourceSlug must not be called by the comics service")
}

func (f *fakeComicsRepository) FindBySourceSlug(
	_ context.Context,
	opts comics.FindBySourceSlugOpts,
) (*comics.Comic, error) {
	f.findBySourceSlugCalls++

	if f.findBySourceSlugCalls > 1 && f.findBySourceSlugSecond != nil {
		return f.findBySourceSlugSecond, nil
	}

	return f.findBySourceSlugResult, f.findBySourceSlugErr
}

func (f *fakeComicsRepository) Create(_ context.Context, opts comics.CreateComicOpts) (*comics.Comic, error) {
	f.createCalls++
	f.lastCreateOpts = opts

	if f.createErr != nil {
		return nil, f.createErr
	}

	if f.createResult != nil {
		return f.createResult, nil
	}

	return &comics.Comic{
		ID:     uuid.New(),
		Slug:   opts.Slug,
		Source: opts.Source,
		Title:  opts.Title,
		Type:   opts.Type,
		Status: opts.Status,
	}, nil
}

//nolint:lll
func (f *fakeComicsRepository) GetBySlugsAndSource(context.Context, comics.GetBySlugsAndSource) ([]comics.Comic, error) {
	panic("GetBySlugsAndSource must not be called by the comics service")
}

func (f *fakeComicsRepository) Delete(_ context.Context, id uuid.UUID) error {
	f.deleteCalls++
	f.lastDeleteID = id

	return f.deleteErr
}

func (f *fakeComicsRepository) GetMany(_ context.Context, _ comics.GetManyOpts) (comics.Page, error) {
	f.getManyCalls++

	return f.getManyResult, f.getManyErr
}

func (f *fakeComicsRepository) ListByStatuses(
	_ context.Context,
	opts comics.ListByStatusesOpts,
) ([]comics.Comic, error) {
	f.listByStatusesCalls++
	f.lastListByStatuses = opts

	return f.listByStatusesResult, f.listByStatusesErr
}

func (f *fakeComicsRepository) UpdateStatusAndChapterCount(
	_ context.Context,
	opts comics.UpdateStatusAndChapterCountOpts,
) error {
	f.updateStatusCalls++
	f.lastUpdateStatus = opts

	return f.updateStatusErr
}

type fakeLibraryRepository struct {
	createErr            error
	deleteErr            error
	existsByComicIDErr   error
	existsByComicID      bool
	createCalls          int
	deleteCalls          int
	existsByComicIDCalls int
	lastCreate           library.CreateOpts
	lastDelete           library.DeleteOpts
	lastExistsComicID    uuid.UUID
}

func (f *fakeLibraryRepository) Create(_ context.Context, opts library.CreateOpts) (*library.Entry, error) {
	f.createCalls++
	f.lastCreate = opts

	if f.createErr != nil {
		return nil, f.createErr
	}

	return &library.Entry{
		ID:      uuid.New(),
		UserID:  opts.UserID,
		ComicID: opts.ComicID,
	}, nil
}

func (f *fakeLibraryRepository) Delete(_ context.Context, opts library.DeleteOpts) error {
	f.deleteCalls++
	f.lastDelete = opts

	return f.deleteErr
}

func (f *fakeLibraryRepository) ExistsByComicID(_ context.Context, comicID uuid.UUID) (bool, error) {
	f.existsByComicIDCalls++
	f.lastExistsComicID = comicID

	return f.existsByComicID, f.existsByComicIDErr
}

func (f *fakeLibraryRepository) ExistsByUserAndComic(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return true, nil
}

type fakeChaptersService struct {
	createAllErr             error
	listByComicIDErr         error
	enqueueDownloadableErr   error
	cleanupComicErr          error
	lastCreateAllChapters    []sources.SourceChapter
	lastEnqueueChapters      []chapters.Chapter
	lastCleanupChapters      []chapters.Chapter
	listByComicIDResult      []chapters.Chapter
	createAllResult          []chapters.Chapter
	createAllCalls           int
	listByComicIDCalls       int
	enqueueDownloadableCalls int
	cleanupComicCalls        int
	lastCreateAllComicID     uuid.UUID
	lastListByComicID        uuid.UUID
	lastCleanupComicID       uuid.UUID
}

func (f *fakeChaptersService) CreateAll(
	_ context.Context,
	comicID uuid.UUID,
	srcChapters []sources.SourceChapter,
) ([]chapters.Chapter, error) {
	f.createAllCalls++
	f.lastCreateAllComicID = comicID
	f.lastCreateAllChapters = srcChapters

	if f.createAllErr != nil {
		return nil, f.createAllErr
	}

	if f.createAllResult != nil {
		return f.createAllResult, nil
	}

	created := make([]chapters.Chapter, len(srcChapters))
	for i, srcChapter := range srcChapters {
		created[i] = chapters.Chapter{
			ID:                uuid.New(),
			ComicID:           comicID,
			SourceChapterSlug: srcChapter.SourceChapterSlug,
			Number:            srcChapter.Number,
			Title:             srcChapter.Title,
			PagesNb:           srcChapter.PageCount,
			PublishedAt:       srcChapter.PublishedAt,
			EarlyAccessUntil:  srcChapter.EarlyAccessUntil,
		}
	}

	return created, nil
}

func (f *fakeChaptersService) ListByComicID(_ context.Context, comicID uuid.UUID) ([]chapters.Chapter, error) {
	f.listByComicIDCalls++
	f.lastListByComicID = comicID

	if f.listByComicIDErr != nil {
		return nil, f.listByComicIDErr
	}

	return f.listByComicIDResult, nil
}

func (f *fakeChaptersService) GetByIds(context.Context, chapters.GetByIdsOpts) ([]chapters.Chapter, error) {
	panic("GetByIds must not be called")
}

func (f *fakeChaptersService) EnqueueDownloadable(_ context.Context, chapterList []chapters.Chapter) error {
	f.enqueueDownloadableCalls++
	f.lastEnqueueChapters = chapterList

	return f.enqueueDownloadableErr
}

func (f *fakeChaptersService) EnqueueResumable(context.Context) error {
	return nil
}

func (f *fakeChaptersService) ScanEarlyAccess(context.Context) error {
	return nil
}

func (f *fakeChaptersService) CleanupComic(_ context.Context, comicID uuid.UUID, chapterList []chapters.Chapter) error {
	f.cleanupComicCalls++
	f.lastCleanupComicID = comicID
	f.lastCleanupChapters = chapterList

	return f.cleanupComicErr
}

func (f *fakeChaptersService) RetryDownload(context.Context, chapters.RetryDownloadOpts) error {
	return nil
}

//nolint:govet // fieldalignment on a test fake is not worth the unreadable field order
type fakeSource struct {
	infos             *sources.GetInfosBySlugResponse
	err               error
	chaptersErr       error
	chapters          []sources.SourceChapter
	lastSlug          string
	calls             int
	chaptersCalls     int
	lastInfosFresh    bool
	lastChaptersFresh bool
}

func (f *fakeSource) GetInfosBySlug(
	_ context.Context,
	opts sources.GetInfosBySlugOpts,
) (*sources.GetInfosBySlugResponse, error) {
	f.calls++
	f.lastSlug = opts.Slug
	f.lastInfosFresh = opts.Fresh

	if f.err != nil {
		return nil, f.err
	}

	return f.infos, nil
}

func (f *fakeSource) GetChaptersBySlug(
	_ context.Context,
	opts sources.GetChaptersBySlugOpts,
) ([]sources.SourceChapter, error) {
	f.chaptersCalls++
	f.lastSlug = opts.Slug
	f.lastChaptersFresh = opts.Fresh

	if f.chaptersErr != nil {
		return nil, f.chaptersErr
	}

	return f.chapters, nil
}

func (f *fakeSource) GetPageURLsByChapter(context.Context, sources.GetPageURLsByChapterOpts) ([]string, error) {
	return nil, nil
}

func validServiceDeps() comics.Deps {
	return comics.Deps{
		ComicsRepository:  &fakeComicsRepository{},
		LibraryRepository: &fakeLibraryRepository{},
		ChaptersService:   &fakeChaptersService{},
		LocalCoverStore:   &fakeLocalCoverStore{},
		Transactor:        &fakeTransactor{},
		Sources: sources.SourceMap{
			testSource: &fakeSource{
				infos: &sources.GetInfosBySlugResponse{
					Slug:         testSlug,
					Title:        "Solo Leveling",
					Status:       sources.SeriesStatusCompleted,
					Type:         sources.SeriesTypeManhwa,
					ChapterCount: 200,
				},
				chapters: []sources.SourceChapter{
					{
						SourceChapterSlug: "chapter-1",
						Number:            1,
						Title:             "Chapter 1",
						PageCount:         42,
					},
				},
			},
		},
	}
}

func TestNewServiceRejectsMissingDeps(t *testing.T) {
	t.Parallel()

	svc, err := comics.NewService(comics.Deps{})
	if err == nil {
		t.Fatal("NewService without deps must fail")
	}

	if svc != nil {
		t.Errorf("NewService returned a service (%v) in addition to the error", svc)
	}

	if got := err.Error(); got != "deps.Validate: comics repository is required" {
		t.Errorf("err = %q, want %q", got, "deps.Validate: comics repository is required")
	}
}

func TestNewServiceRejectsMissingChaptersService(t *testing.T) {
	t.Parallel()

	deps := validServiceDeps()
	deps.ChaptersService = nil

	svc, err := comics.NewService(deps)
	if err == nil {
		t.Fatal("NewService without chapters service must fail")
	}

	if svc != nil {
		t.Errorf("NewService returned a service (%v) in addition to the error", svc)
	}

	if got := err.Error(); got != "deps.Validate: chapters service is required" {
		t.Errorf("err = %q, want %q", got, "deps.Validate: chapters service is required")
	}
}

func TestCreateReturnsExistingComicAndAddsLibraryEntry(t *testing.T) {
	t.Parallel()

	userB := uuid.New()
	existing := &comics.Comic{ID: uuid.New(), Slug: testSlug, Source: testSource}
	existingChapters := []chapters.Chapter{
		{
			ID:                uuid.New(),
			ComicID:           existing.ID,
			SourceChapterSlug: "chapter-1",
			Number:            1,
			PagesNb:           42,
		},
	}

	comicsRepo := &fakeComicsRepository{
		findBySourceSlugResult: existing,
	}
	libraryRepo := &fakeLibraryRepository{}
	chaptersSvc := &fakeChaptersService{listByComicIDResult: existingChapters}
	coverStore := &fakeLocalCoverStore{}
	source := &fakeSource{}
	deps := validServiceDeps()
	deps.ComicsRepository = comicsRepo
	deps.LibraryRepository = libraryRepo
	deps.ChaptersService = chaptersSvc
	deps.LocalCoverStore = coverStore
	deps.Sources = sources.SourceMap{testSource: source}

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	got, err := svc.Create(context.Background(), comics.CreateOpts{
		UserID: userB,
		Source: testSource,
		Slug:   testSlug,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if got != existing {
		t.Errorf("Create() = %+v, want existing comic %+v", got, existing)
	}

	if comicsRepo.findBySourceSlugCalls != 1 {
		t.Errorf("FindBySourceSlug called %d times, want 1", comicsRepo.findBySourceSlugCalls)
	}

	if source.calls != 0 {
		t.Errorf("source.GetInfosBySlug called %d times, want 0", source.calls)
	}

	if source.chaptersCalls != 0 {
		t.Errorf("source.GetChaptersBySlug called %d times, want 0", source.chaptersCalls)
	}

	if libraryRepo.createCalls != 1 {
		t.Errorf("LibraryRepository.Create called %d times, want 1", libraryRepo.createCalls)
	}

	if comicsRepo.createCalls != 0 {
		t.Errorf("ComicsRepository.Create called %d times, want 0", comicsRepo.createCalls)
	}

	if chaptersSvc.createAllCalls != 0 {
		t.Errorf("ChaptersService.CreateAll called %d times, want 0", chaptersSvc.createAllCalls)
	}

	if libraryRepo.lastCreate.UserID != userB || libraryRepo.lastCreate.ComicID != existing.ID {
		t.Errorf("library CreateOpts = %+v", libraryRepo.lastCreate)
	}

	if chaptersSvc.listByComicIDCalls != 1 || chaptersSvc.lastListByComicID != existing.ID {
		t.Errorf("ListByComicID called %d times for comic %s", chaptersSvc.listByComicIDCalls, chaptersSvc.lastListByComicID)
	}

	if chaptersSvc.enqueueDownloadableCalls != 1 {
		t.Errorf("EnqueueDownloadable called %d times, want 1", chaptersSvc.enqueueDownloadableCalls)
	}

	if len(chaptersSvc.lastEnqueueChapters) != 1 {
		t.Errorf("enqueued chapters = %+v, want existing chapter list", chaptersSvc.lastEnqueueChapters)
	}

	if coverStore.obtainCalls != 0 {
		t.Errorf("ObtainLocal called %d times, want 0", coverStore.obtainCalls)
	}
}

func TestCreateFetchesSourceAndPersistsLibraryEntry(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicsRepo := &fakeComicsRepository{findBySourceSlugErr: domain.ErrNotFound}
	libraryRepo := &fakeLibraryRepository{}
	chaptersSvc := &fakeChaptersService{}
	coverStore := &fakeLocalCoverStore{}
	tx := &fakeTransactor{}
	source := &fakeSource{
		infos: &sources.GetInfosBySlugResponse{
			Slug:         testSlug,
			Title:        "Solo Leveling",
			Status:       sources.SeriesStatusCompleted,
			Type:         sources.SeriesTypeManhwa,
			ChapterCount: 200,
			Author:       "Chugong",
		},
		chapters: []sources.SourceChapter{
			{
				SourceChapterSlug: "chapter-1",
				Number:            1,
				Title:             "Chapter 1",
				PageCount:         42,
				PublishedAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	deps := validServiceDeps()
	deps.ComicsRepository = comicsRepo
	deps.LibraryRepository = libraryRepo
	deps.ChaptersService = chaptersSvc
	deps.LocalCoverStore = coverStore
	deps.Transactor = tx
	deps.Sources = sources.SourceMap{testSource: source}

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	got, err := svc.Create(context.Background(), comics.CreateOpts{
		UserID: userID,
		Source: testSource,
		Slug:   testSlug,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if got.Slug != testSlug || got.Source != testSource {
		t.Errorf("Create() = %+v", got)
	}

	if tx.calls != 1 {
		t.Errorf("WithinTx called %d times, want 1", tx.calls)
	}

	if comicsRepo.createCalls != 1 {
		t.Errorf("ComicsRepository.Create called %d times, want 1", comicsRepo.createCalls)
	}

	if comicsRepo.lastCreateOpts.Slug != testSlug || comicsRepo.lastCreateOpts.Author != "Chugong" {
		t.Errorf("CreateComicOpts = %+v", comicsRepo.lastCreateOpts)
	}

	if libraryRepo.createCalls != 1 {
		t.Errorf("LibraryRepository.Create called %d times, want 1", libraryRepo.createCalls)
	}

	if libraryRepo.lastCreate.UserID != userID || libraryRepo.lastCreate.ComicID != got.ID {
		t.Errorf("library CreateOpts = %+v, comic ID = %s", libraryRepo.lastCreate, got.ID)
	}

	if source.lastSlug != testSlug {
		t.Errorf("source slug = %q, want %q", source.lastSlug, testSlug)
	}

	if source.chaptersCalls != 1 {
		t.Errorf("source.GetChaptersBySlug called %d times, want 1", source.chaptersCalls)
	}

	if chaptersSvc.createAllCalls != 1 || chaptersSvc.lastCreateAllComicID != got.ID {
		t.Errorf("CreateAll called %d times for comic %s", chaptersSvc.createAllCalls, chaptersSvc.lastCreateAllComicID)
	}

	if len(chaptersSvc.lastCreateAllChapters) != 1 || chaptersSvc.lastCreateAllChapters[0].PageCount != 42 {
		t.Errorf("CreateAll chapters = %+v", chaptersSvc.lastCreateAllChapters)
	}

	if chaptersSvc.listByComicIDCalls != 1 {
		t.Errorf("ListByComicID called %d times, want 1", chaptersSvc.listByComicIDCalls)
	}

	if chaptersSvc.enqueueDownloadableCalls != 1 {
		t.Errorf("EnqueueDownloadable called %d times, want 1", chaptersSvc.enqueueDownloadableCalls)
	}

	if coverStore.obtainCalls != 1 {
		t.Errorf("ObtainLocal called %d times, want 1", coverStore.obtainCalls)
	}

	if coverStore.lastObtain.slug != testSlug {
		t.Errorf("ObtainLocal slug = %q, want %q", coverStore.lastObtain.slug, testSlug)
	}

	if comicsRepo.lastCreateOpts.ID != coverStore.lastObtain.id {
		t.Errorf("CreateComicOpts.ID = %s, want %s", comicsRepo.lastCreateOpts.ID, coverStore.lastObtain.id)
	}

	if comicsRepo.lastCreateOpts.ID == uuid.Nil {
		t.Error("CreateComicOpts.ID must not be uuid.Nil")
	}

	if coverStore.removeCalls != 0 {
		t.Errorf("RemoveLocal called %d times, want 0", coverStore.removeCalls)
	}
}

func TestCreateUnknownSource(t *testing.T) {
	t.Parallel()

	deps := validServiceDeps()
	deps.ComicsRepository = &fakeComicsRepository{findBySourceSlugErr: domain.ErrNotFound}

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	_, err = svc.Create(context.Background(), comics.CreateOpts{
		UserID: uuid.New(),
		Source: "unknown",
		Slug:   testSlug,
	})
	if err == nil {
		t.Fatal("Create with unknown source must fail")
	}

	if want := "source unknown not found"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
}

func TestCreatePropagatesChapterListError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("chapters unavailable")
	source := &fakeSource{
		infos: &sources.GetInfosBySlugResponse{
			Slug:  testSlug,
			Title: "Solo Leveling",
		},
		chaptersErr: sentinel,
	}
	deps := validServiceDeps()
	deps.ComicsRepository = &fakeComicsRepository{findBySourceSlugErr: domain.ErrNotFound}
	deps.Sources = sources.SourceMap{testSource: source}

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	_, err = svc.Create(context.Background(), comics.CreateOpts{
		UserID: uuid.New(),
		Source: testSource,
		Slug:   testSlug,
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("Create = %v, original error no longer reachable", err)
	}
}

func TestCreatePropagatesSourceError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("upstream unavailable")
	source := &fakeSource{err: sentinel}
	deps := validServiceDeps()
	deps.ComicsRepository = &fakeComicsRepository{findBySourceSlugErr: domain.ErrNotFound}
	deps.Sources = sources.SourceMap{testSource: source}

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	_, err = svc.Create(context.Background(), comics.CreateOpts{
		UserID: uuid.New(),
		Source: testSource,
		Slug:   testSlug,
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("Create = %v, original error no longer reachable", err)
	}
}

func TestGetByIDDelegatesToRepository(t *testing.T) {
	t.Parallel()

	comic := &comics.Comic{ID: uuid.New()}
	repo := &fakeComicsRepository{getByIDResult: comic}
	deps := validServiceDeps()
	deps.ComicsRepository = repo

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	got, err := svc.GetByID(context.Background(), comics.GetByIDOpts{
		UserID: uuid.New(),
		ID:     comic.ID,
	})
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got != comic {
		t.Errorf("GetByID() = %+v, want %+v", got, comic)
	}

	if repo.getByIDCalls != 1 {
		t.Errorf("GetByID called %d times, want 1", repo.getByIDCalls)
	}
}

func TestGetManyDelegatesToRepository(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	items := []comics.Comic{{ID: uuid.New(), Slug: testSlug}}
	repo := &fakeComicsRepository{getManyResult: comics.Page{Items: items}}
	deps := validServiceDeps()
	deps.ComicsRepository = repo

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	got, err := svc.GetMany(context.Background(), comics.GetManyOpts{UserID: &userID})
	if err != nil {
		t.Fatalf("GetMany: %v", err)
	}

	if len(got.Items) != 1 || got.Items[0].ID != items[0].ID {
		t.Errorf("GetMany() = %+v, want %+v", got.Items, items)
	}
}

func TestDeleteRemovesLibraryEntryAndComic(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicID := uuid.New()
	chapterID := uuid.New()
	comicChapters := []chapters.Chapter{
		{ID: chapterID, ComicID: comicID, Number: 1},
	}
	comicsRepo := &fakeComicsRepository{}
	libraryRepo := &fakeLibraryRepository{}
	chaptersSvc := &fakeChaptersService{listByComicIDResult: comicChapters}
	tx := &fakeTransactor{}

	deps := validServiceDeps()
	deps.ComicsRepository = comicsRepo
	deps.LibraryRepository = libraryRepo
	deps.ChaptersService = chaptersSvc
	deps.Transactor = tx

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	err = svc.Delete(context.Background(), comics.DeleteOpts{
		UserID: userID,
		ID:     comicID,
	})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if tx.calls != 1 {
		t.Errorf("WithinTx called %d times, want 1", tx.calls)
	}

	if libraryRepo.deleteCalls != 1 {
		t.Errorf("LibraryRepository.Delete called %d times, want 1", libraryRepo.deleteCalls)
	}

	if libraryRepo.lastDelete.UserID != userID || libraryRepo.lastDelete.ComicID != comicID {
		t.Errorf("library DeleteOpts = %+v", libraryRepo.lastDelete)
	}

	if chaptersSvc.listByComicIDCalls != 1 || chaptersSvc.lastListByComicID != comicID {
		t.Errorf("ListByComicID comic ID = %s, want %s", chaptersSvc.lastListByComicID, comicID)
	}

	if comicsRepo.deleteCalls != 1 || comicsRepo.lastDeleteID != comicID {
		t.Errorf("ComicsRepository.Delete comic ID = %s, want %s", comicsRepo.lastDeleteID, comicID)
	}

	if libraryRepo.existsByComicIDCalls != 1 || libraryRepo.lastExistsComicID != comicID {
		t.Errorf("ExistsByComicID comic ID = %s, want %s", libraryRepo.lastExistsComicID, comicID)
	}

	if chaptersSvc.cleanupComicCalls != 1 || chaptersSvc.lastCleanupComicID != comicID {
		t.Errorf("CleanupComic comic ID = %s, want %s", chaptersSvc.lastCleanupComicID, comicID)
	}

	if len(chaptersSvc.lastCleanupChapters) != 1 || chaptersSvc.lastCleanupChapters[0].ID != chapterID {
		t.Errorf("CleanupComic chapters = %+v", chaptersSvc.lastCleanupChapters)
	}
}

func TestDeleteKeepsComicWhenOtherUsersHaveLibraryEntry(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicID := uuid.New()
	comicsRepo := &fakeComicsRepository{}
	libraryRepo := &fakeLibraryRepository{existsByComicID: true}
	chaptersSvc := &fakeChaptersService{}
	tx := &fakeTransactor{}

	deps := validServiceDeps()
	deps.ComicsRepository = comicsRepo
	deps.LibraryRepository = libraryRepo
	deps.ChaptersService = chaptersSvc
	deps.Transactor = tx

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	err = svc.Delete(context.Background(), comics.DeleteOpts{
		UserID: userID,
		ID:     comicID,
	})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if libraryRepo.deleteCalls != 1 {
		t.Errorf("LibraryRepository.Delete called %d times, want 1", libraryRepo.deleteCalls)
	}

	if libraryRepo.existsByComicIDCalls != 1 {
		t.Errorf("ExistsByComicID called %d times, want 1", libraryRepo.existsByComicIDCalls)
	}

	if comicsRepo.deleteCalls != 0 {
		t.Errorf("ComicsRepository.Delete called %d times, want 0", comicsRepo.deleteCalls)
	}

	if chaptersSvc.listByComicIDCalls != 0 {
		t.Errorf("ListByComicID called %d times, want 0", chaptersSvc.listByComicIDCalls)
	}

	if chaptersSvc.cleanupComicCalls != 0 {
		t.Errorf("CleanupComic called %d times, want 0", chaptersSvc.cleanupComicCalls)
	}
}

func TestDeleteReturnsWhenLibraryEntryMissing(t *testing.T) {
	t.Parallel()

	deps := validServiceDeps()
	deps.LibraryRepository = &fakeLibraryRepository{deleteErr: domain.ErrNotFound}
	deps.ComicsRepository = &fakeComicsRepository{}

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	err = svc.Delete(context.Background(), comics.DeleteOpts{
		UserID: uuid.New(),
		ID:     uuid.New(),
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete = %v, want domain.ErrNotFound", err)
	}
}

func TestCreateDoesNotInsertWhenObtainLocalFails(t *testing.T) {
	t.Parallel()

	comicsRepo := &fakeComicsRepository{findBySourceSlugErr: domain.ErrNotFound}
	libraryRepo := &fakeLibraryRepository{}
	coverStore := &fakeLocalCoverStore{obtainErr: errors.New("cdn down")}
	deps := validServiceDeps()
	deps.ComicsRepository = comicsRepo
	deps.LibraryRepository = libraryRepo
	deps.LocalCoverStore = coverStore

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	_, err = svc.Create(context.Background(), comics.CreateOpts{
		UserID: uuid.New(),
		Source: testSource,
		Slug:   testSlug,
	})
	if err == nil {
		t.Fatal("Create must fail when ObtainLocal fails")
	}

	if comicsRepo.createCalls != 0 {
		t.Errorf("ComicsRepository.Create called %d times, want 0", comicsRepo.createCalls)
	}

	if libraryRepo.createCalls != 0 {
		t.Errorf("LibraryRepository.Create called %d times, want 0", libraryRepo.createCalls)
	}
}

func TestCreateUniqueRaceRemovesOrphanCoverAndAddsExisting(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	winner := &comics.Comic{ID: uuid.New(), Slug: testSlug, Source: testSource}
	winnerChapters := []chapters.Chapter{
		{ID: uuid.New(), ComicID: winner.ID, SourceChapterSlug: testChapterSlug, Number: 1},
	}

	comicsRepo := &fakeComicsRepository{
		findBySourceSlugErr:    domain.ErrNotFound,
		findBySourceSlugSecond: winner,
		createErr:              domain.ErrAlreadyExists,
	}
	libraryRepo := &fakeLibraryRepository{}
	chaptersSvc := &fakeChaptersService{listByComicIDResult: winnerChapters}
	coverStore := &fakeLocalCoverStore{}
	deps := validServiceDeps()
	deps.ComicsRepository = comicsRepo
	deps.LibraryRepository = libraryRepo
	deps.ChaptersService = chaptersSvc
	deps.LocalCoverStore = coverStore

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	got, err := svc.Create(context.Background(), comics.CreateOpts{
		UserID: userID,
		Source: testSource,
		Slug:   testSlug,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if got != winner {
		t.Errorf("Create() = %+v, want winner %+v", got, winner)
	}

	if coverStore.removeCalls != 1 {
		t.Errorf("RemoveLocal called %d times, want 1", coverStore.removeCalls)
	}

	if coverStore.lastRemove != coverStore.lastObtain.id {
		t.Errorf("RemoveLocal comic ID = %s, want %s", coverStore.lastRemove, coverStore.lastObtain.id)
	}

	if libraryRepo.lastCreate.ComicID != winner.ID {
		t.Errorf("library CreateOpts.ComicID = %s, want %s", libraryRepo.lastCreate.ComicID, winner.ID)
	}
}

func TestCreateRemovesCoverWhenTxFails(t *testing.T) {
	t.Parallel()

	comicsRepo := &fakeComicsRepository{
		findBySourceSlugErr: domain.ErrNotFound,
		createErr:           errors.New("db down"),
	}
	coverStore := &fakeLocalCoverStore{}
	deps := validServiceDeps()
	deps.ComicsRepository = comicsRepo
	deps.LocalCoverStore = coverStore

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	_, err = svc.Create(context.Background(), comics.CreateOpts{
		UserID: uuid.New(),
		Source: testSource,
		Slug:   testSlug,
	})
	if err == nil {
		t.Fatal("Create must fail when tx fails")
	}

	if coverStore.removeCalls != 1 {
		t.Errorf("RemoveLocal called %d times, want 1", coverStore.removeCalls)
	}
}

func TestServeCoverNotInLibrary(t *testing.T) {
	t.Parallel()

	comicsRepo := &fakeComicsRepository{getByIDErr: domain.ErrNotFound}
	coverStore := &fakeLocalCoverStore{}
	deps := validServiceDeps()
	deps.ComicsRepository = comicsRepo
	deps.LocalCoverStore = coverStore

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	_, _, err = svc.ServeCover(context.Background(), comics.GetByIDOpts{
		UserID: uuid.New(),
		ID:     uuid.New(),
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ServeCover = %v, want domain.ErrNotFound", err)
	}

	if coverStore.serveCalls != 0 {
		t.Errorf("ServeLocal called %d times, want 0", coverStore.serveCalls)
	}
}

func TestServeCoverReturnsLocalPath(t *testing.T) {
	t.Parallel()

	comic := &comics.Comic{ID: uuid.New()}
	comicsRepo := &fakeComicsRepository{getByIDResult: comic}
	coverStore := &fakeLocalCoverStore{
		diskPath: "/covers/abc.webp",
		mime:     "image/webp",
	}
	deps := validServiceDeps()
	deps.ComicsRepository = comicsRepo
	deps.LocalCoverStore = coverStore

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	path, contentType, err := svc.ServeCover(context.Background(), comics.GetByIDOpts{
		UserID: uuid.New(),
		ID:     comic.ID,
	})
	if err != nil {
		t.Fatalf("ServeCover: %v", err)
	}

	if path != coverStore.diskPath {
		t.Errorf("disk path = %q, want %q", path, coverStore.diskPath)
	}

	if contentType != coverStore.mime {
		t.Errorf("content type = %q, want %q", contentType, coverStore.mime)
	}

	if coverStore.serveCalls != 1 {
		t.Errorf("ServeLocal called %d times, want 1", coverStore.serveCalls)
	}
}
