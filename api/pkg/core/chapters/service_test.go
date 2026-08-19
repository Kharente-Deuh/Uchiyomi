// SPDX-License-Identifier: AGPL-3.0-or-later

//nolint:goconst,lll
package chapters_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/library"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

type fakeChaptersRepository struct {
	lastListEarlyAccessUntil time.Time
	getByIDErr               error
	getByIdsErr              error
	listByComicIDErr         error
	getByIDResult            *chapters.Chapter
	lastCreateMany           []chapters.CreateOpts
	listResumableResult      []chapters.Chapter
	listEarlyAccessResult    []chapters.Chapter
	listByComicIDResult      []chapters.Chapter
	getByIdsResult           []chapters.Chapter
	lastGetByIds             []uuid.UUID
	lastListByComicID        uuid.UUID
	createManyCalls          int
	listResumableCalls       int
	listEarlyAccessCalls     int
	listByComicIDCalls       int
	getByIDCalls             int
	getByIdsCalls            int
}

func (f *fakeChaptersRepository) Create(context.Context, chapters.CreateOpts) (*chapters.Chapter, error) {
	panic("Create must not be called")
}

func (f *fakeChaptersRepository) CreateMany(_ context.Context, opts []chapters.CreateOpts) ([]chapters.Chapter, error) {
	f.createManyCalls++
	f.lastCreateMany = opts

	created := make([]chapters.Chapter, len(opts))
	for i, opt := range opts {
		created[i] = chapters.Chapter{
			ID:                uuid.New(),
			ComicID:           opt.ComicID,
			SourceChapterSlug: opt.SourceChapterSlug,
			Number:            opt.Number,
			Title:             opt.Title,
			PagesNb:           opt.PagesNb,
			PublishedAt:       opt.PublishedAt,
			EarlyAccessUntil:  opt.EarlyAccessUntil,
		}
	}

	return created, nil
}

func (f *fakeChaptersRepository) ListByComicID(_ context.Context, comicID uuid.UUID) ([]chapters.Chapter, error) {
	f.listByComicIDCalls++
	f.lastListByComicID = comicID

	if f.listByComicIDErr != nil {
		return nil, f.listByComicIDErr
	}

	return f.listByComicIDResult, nil
}

func (f *fakeChaptersRepository) ListResumable(context.Context, time.Time) ([]chapters.Chapter, error) {
	f.listResumableCalls++

	return f.listResumableResult, nil
}

func (f *fakeChaptersRepository) ListEarlyAccessUnlocked(_ context.Context, now time.Time) ([]chapters.Chapter, error) {
	f.listEarlyAccessCalls++
	f.lastListEarlyAccessUntil = now

	return f.listEarlyAccessResult, nil
}

func (f *fakeChaptersRepository) GetByID(_ context.Context, id uuid.UUID) (*chapters.Chapter, error) {
	f.getByIDCalls++

	if f.getByIDErr != nil {
		return nil, f.getByIDErr
	}

	if f.getByIDResult != nil {
		return f.getByIDResult, nil
	}

	return &chapters.Chapter{ID: id}, nil
}

func (f *fakeChaptersRepository) UpdateDownload(context.Context, uuid.UUID, int) error {
	panic("UpdateDownload must not be called")
}

func (f *fakeChaptersRepository) UpdatePagesNb(context.Context, uuid.UUID, int) error {
	panic("UpdatePagesNb must not be called")
}

func (f *fakeChaptersRepository) GetByIds(_ context.Context, ids []uuid.UUID) ([]chapters.Chapter, error) {
	f.getByIdsCalls++
	f.lastGetByIds = ids

	if f.getByIdsErr != nil {
		return nil, f.getByIdsErr
	}

	return f.getByIdsResult, nil
}

type fakeChapterDownloader struct {
	lastEnqueueChapters  []chapters.Chapter
	lastCleanupChapters  []chapters.Chapter
	lastResetChapterID   uuid.UUID
	lastResumeChapterID  uuid.UUID
	enqueueCalls         int
	cleanupComicCalls    int
	resetAndEnqueueCalls int
	resumeCalls          int
	lastCleanupComicID   uuid.UUID
}

func (f *fakeChapterDownloader) Enqueue(_ context.Context, chapterList []chapters.Chapter) error {
	f.enqueueCalls++
	f.lastEnqueueChapters = chapterList

	return nil
}

func (f *fakeChapterDownloader) CleanupComic(_ context.Context, comicID uuid.UUID, chapterList []chapters.Chapter) error {
	f.cleanupComicCalls++
	f.lastCleanupComicID = comicID
	f.lastCleanupChapters = chapterList

	return nil
}

func (f *fakeChapterDownloader) ResetAndEnqueue(_ context.Context, chapterID uuid.UUID) error {
	f.resetAndEnqueueCalls++
	f.lastResetChapterID = chapterID

	return nil
}

func (f *fakeChapterDownloader) Resume(_ context.Context, chapterID uuid.UUID) error {
	f.resumeCalls++
	f.lastResumeChapterID = chapterID

	return nil
}

type fakeLibraryRepository struct {
	existsErr            error
	inLibraryComicIDs    map[uuid.UUID]bool
	lastUserID           uuid.UUID
	lastComicID          uuid.UUID
	existsByUserAndComic bool
	existsCalls          int
}

func (f *fakeLibraryRepository) Create(context.Context, library.CreateOpts) (*library.Entry, error) {
	panic("Create must not be called")
}

func (f *fakeLibraryRepository) Delete(context.Context, library.DeleteOpts) error {
	panic("Delete must not be called")
}

func (f *fakeLibraryRepository) ExistsByComicID(context.Context, uuid.UUID) (bool, error) {
	panic("ExistsByComicID must not be called")
}

func (f *fakeLibraryRepository) ExistsByUserAndComic(
	_ context.Context, userID, comicID uuid.UUID,
) (bool, error) {
	f.existsCalls++
	f.lastUserID = userID
	f.lastComicID = comicID

	if f.inLibraryComicIDs != nil {
		return f.inLibraryComicIDs[comicID], f.existsErr
	}

	return f.existsByUserAndComic, f.existsErr
}

type fakeComicLookup struct {
	existsErr error
	lastID    uuid.UUID
	exists    bool
	calls     int
}

func (f *fakeComicLookup) Exists(_ context.Context, id uuid.UUID) (bool, error) {
	f.calls++
	f.lastID = id

	return f.exists, f.existsErr
}

func newTestService(
	repo *fakeChaptersRepository,
	downloader *fakeChapterDownloader,
	libraryRepo *fakeLibraryRepository,
) *chapters.Service {
	svc, err := chapters.NewService(chapters.Deps{
		Repository:        repo,
		ChapterDownloader: downloader,
		LibraryRepository: libraryRepo,
		ComicLookup:       &fakeComicLookup{exists: true},
	})
	if err != nil {
		panic(err)
	}

	return svc
}

func TestServiceCreateAll(t *testing.T) {
	t.Parallel()

	repo := &fakeChaptersRepository{}
	svc := newTestService(repo, &fakeChapterDownloader{}, &fakeLibraryRepository{existsByUserAndComic: true})

	comicID := uuid.New()
	publishedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	got, err := svc.CreateAll(context.Background(), comicID, []sources.SourceChapter{
		{
			SourceChapterSlug: "chapter-1",
			Number:            1,
			Title:             "Chapter 1",
			PageCount:         42,
			PublishedAt:       publishedAt,
		},
	})
	if err != nil {
		t.Fatalf("CreateAll: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("len(CreateAll()) = %d, want 1", len(got))
	}

	if repo.createManyCalls != 1 {
		t.Errorf("CreateMany called %d times, want 1", repo.createManyCalls)
	}

	if len(repo.lastCreateMany) != 1 || repo.lastCreateMany[0].PagesNb != 42 || repo.lastCreateMany[0].ComicID != comicID {
		t.Errorf("CreateMany opts = %+v", repo.lastCreateMany)
	}
}

func TestServiceListByComicID(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	want := []chapters.Chapter{{ID: uuid.New(), ComicID: comicID, Title: "Chapter 1"}}
	repo := &fakeChaptersRepository{listByComicIDResult: want}
	svc := newTestService(repo, &fakeChapterDownloader{}, &fakeLibraryRepository{})

	got, err := svc.ListByComicID(context.Background(), comicID)
	if err != nil {
		t.Fatalf("ListByComicID: %v", err)
	}

	if repo.listByComicIDCalls != 1 || repo.lastListByComicID != comicID {
		t.Errorf("ListByComicID comic = %s, calls = %d", repo.lastListByComicID, repo.listByComicIDCalls)
	}

	if len(got) != 1 || got[0].ID != want[0].ID {
		t.Errorf("ListByComicID() = %+v, want %+v", got, want)
	}
}

func TestServiceListByComicIDError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("db down")
	repo := &fakeChaptersRepository{listByComicIDErr: sentinel}
	svc := newTestService(repo, &fakeChapterDownloader{}, &fakeLibraryRepository{})

	got, err := svc.ListByComicID(context.Background(), uuid.New())
	if !errors.Is(err, sentinel) {
		t.Fatalf("ListByComicID = %v, want wrapped sentinel", err)
	}

	if got != nil {
		t.Errorf("ListByComicID returned %+v in addition to the error", got)
	}
}

func TestServiceEnqueueDownloadableRunsWithoutError(t *testing.T) {
	t.Parallel()

	downloader := &fakeChapterDownloader{}
	repo := &fakeChaptersRepository{}
	svc := newTestService(repo, downloader, &fakeLibraryRepository{existsByUserAndComic: true})

	now := time.Now()
	lockedUntil := now.Add(time.Hour)
	unlockedUntil := now.Add(-time.Hour)

	err := svc.EnqueueDownloadable(context.Background(), []chapters.Chapter{
		{Download: 100},
		{EarlyAccessUntil: &lockedUntil},
		{Download: 0, EarlyAccessUntil: &unlockedUntil},
	})
	if err != nil {
		t.Fatalf("EnqueueDownloadable: %v", err)
	}

	if downloader.enqueueCalls != 1 {
		t.Errorf("Enqueue called %d times, want 1", downloader.enqueueCalls)
	}

	if len(downloader.lastEnqueueChapters) != 1 {
		t.Fatalf("enqueued chapters = %+v, want one downloadable chapter", downloader.lastEnqueueChapters)
	}
}

func TestServiceEnqueueResumable(t *testing.T) {
	t.Parallel()

	downloader := &fakeChapterDownloader{}
	repo := &fakeChaptersRepository{
		listResumableResult: []chapters.Chapter{
			{ID: uuid.New(), Download: 42},
			{ID: uuid.New(), Download: -1},
		},
	}
	svc := newTestService(repo, downloader, &fakeLibraryRepository{existsByUserAndComic: true})

	err := svc.EnqueueResumable(context.Background())
	if err != nil {
		t.Fatalf("EnqueueResumable: %v", err)
	}

	if repo.listResumableCalls != 1 {
		t.Errorf("ListResumable called %d times, want 1", repo.listResumableCalls)
	}

	if downloader.enqueueCalls != 1 {
		t.Errorf("Enqueue called %d times, want 1", downloader.enqueueCalls)
	}

	if len(downloader.lastEnqueueChapters) != 2 {
		t.Fatalf("enqueued chapters = %+v, want 2 resumable chapters", downloader.lastEnqueueChapters)
	}
}

func TestServiceScanEarlyAccess(t *testing.T) {
	t.Parallel()

	now := time.Now()
	lockedUntil := now.Add(time.Hour)
	unlockedUntil := now.Add(-time.Hour)
	downloader := &fakeChapterDownloader{}
	repo := &fakeChaptersRepository{
		listEarlyAccessResult: []chapters.Chapter{
			{Download: 100},
			{EarlyAccessUntil: &lockedUntil},
			{Download: 0, EarlyAccessUntil: &unlockedUntil},
		},
	}
	svc := newTestService(repo, downloader, &fakeLibraryRepository{existsByUserAndComic: true})

	err := svc.ScanEarlyAccess(context.Background())
	if err != nil {
		t.Fatalf("ScanEarlyAccess: %v", err)
	}

	if repo.listEarlyAccessCalls != 1 {
		t.Errorf("ListEarlyAccessUnlocked called %d times, want 1", repo.listEarlyAccessCalls)
	}

	if downloader.enqueueCalls != 1 {
		t.Errorf("Enqueue called %d times, want 1", downloader.enqueueCalls)
	}

	if len(downloader.lastEnqueueChapters) != 1 {
		t.Fatalf("enqueued chapters = %+v, want one unlocked chapter", downloader.lastEnqueueChapters)
	}
}

func TestServiceCleanupComic(t *testing.T) {
	t.Parallel()

	downloader := &fakeChapterDownloader{}
	svc := newTestService(&fakeChaptersRepository{}, downloader, &fakeLibraryRepository{existsByUserAndComic: true})

	comicID := uuid.New()
	chapterList := []chapters.Chapter{
		{ID: uuid.New(), ComicID: comicID, Number: 1},
	}

	err := svc.CleanupComic(context.Background(), comicID, chapterList)
	if err != nil {
		t.Fatalf("CleanupComic: %v", err)
	}

	if downloader.cleanupComicCalls != 1 || downloader.lastCleanupComicID != comicID {
		t.Errorf("CleanupComic comic ID = %s, want %s", downloader.lastCleanupComicID, comicID)
	}

	if len(downloader.lastCleanupChapters) != 1 {
		t.Errorf("CleanupComic chapters = %+v", downloader.lastCleanupChapters)
	}
}

func TestServiceRetryDownloadNotFound(t *testing.T) {
	t.Parallel()

	chapterID := uuid.New()
	repo := &fakeChaptersRepository{getByIDErr: domain.ErrNotFound}
	svc := newTestService(repo, &fakeChapterDownloader{}, &fakeLibraryRepository{})

	err := svc.RetryDownload(context.Background(), chapters.RetryDownloadOpts{
		UserID:    uuid.New(),
		ChapterID: chapterID,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("RetryDownload = %v, want domain.ErrNotFound", err)
	}
}

func TestServiceRetryDownloadForbiddenWhenNotInLibrary(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicID := uuid.New()
	chapterID := uuid.New()
	repo := &fakeChaptersRepository{
		getByIDResult: &chapters.Chapter{ID: chapterID, ComicID: comicID, Download: 42},
	}
	libraryRepo := &fakeLibraryRepository{existsByUserAndComic: false}
	svc := newTestService(repo, &fakeChapterDownloader{}, libraryRepo)

	err := svc.RetryDownload(context.Background(), chapters.RetryDownloadOpts{
		UserID:    userID,
		ChapterID: chapterID,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("RetryDownload = %v, want domain.ErrForbidden", err)
	}

	if libraryRepo.lastUserID != userID || libraryRepo.lastComicID != comicID {
		t.Errorf("library check user/comic = %s/%s, want %s/%s",
			libraryRepo.lastUserID, libraryRepo.lastComicID, userID, comicID)
	}
}

func TestServiceRetryDownloadConflictWhenComplete(t *testing.T) {
	t.Parallel()

	chapterID := uuid.New()
	repo := &fakeChaptersRepository{
		getByIDResult: &chapters.Chapter{ID: chapterID, ComicID: uuid.New(), Download: 100},
	}
	svc := newTestService(repo, &fakeChapterDownloader{}, &fakeLibraryRepository{existsByUserAndComic: true})

	err := svc.RetryDownload(context.Background(), chapters.RetryDownloadOpts{
		UserID:    uuid.New(),
		ChapterID: chapterID,
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("RetryDownload = %v, want domain.ErrConflict", err)
	}
}

func TestServiceRetryDownloadResetsFailedChapter(t *testing.T) {
	t.Parallel()

	chapterID := uuid.New()
	downloader := &fakeChapterDownloader{}
	repo := &fakeChaptersRepository{
		getByIDResult: &chapters.Chapter{ID: chapterID, ComicID: uuid.New(), Download: -1},
	}
	svc := newTestService(repo, downloader, &fakeLibraryRepository{existsByUserAndComic: true})

	err := svc.RetryDownload(context.Background(), chapters.RetryDownloadOpts{
		UserID:    uuid.New(),
		ChapterID: chapterID,
	})
	if err != nil {
		t.Fatalf("RetryDownload: %v", err)
	}

	if downloader.resetAndEnqueueCalls != 1 || downloader.lastResetChapterID != chapterID {
		t.Errorf("ResetAndEnqueue chapter = %s, calls = %d", downloader.lastResetChapterID, downloader.resetAndEnqueueCalls)
	}

	if downloader.resumeCalls != 0 {
		t.Errorf("Resume called %d times, want 0", downloader.resumeCalls)
	}
}

func TestServiceRetryDownloadResumesInterruptedChapter(t *testing.T) {
	t.Parallel()

	chapterID := uuid.New()
	downloader := &fakeChapterDownloader{}
	repo := &fakeChaptersRepository{
		getByIDResult: &chapters.Chapter{ID: chapterID, ComicID: uuid.New(), Download: 42},
	}
	svc := newTestService(repo, downloader, &fakeLibraryRepository{existsByUserAndComic: true})

	err := svc.RetryDownload(context.Background(), chapters.RetryDownloadOpts{
		UserID:    uuid.New(),
		ChapterID: chapterID,
	})
	if err != nil {
		t.Fatalf("RetryDownload: %v", err)
	}

	if downloader.resumeCalls != 1 || downloader.lastResumeChapterID != chapterID {
		t.Errorf("Resume chapter = %s, calls = %d", downloader.lastResumeChapterID, downloader.resumeCalls)
	}

	if downloader.resetAndEnqueueCalls != 0 {
		t.Errorf("ResetAndEnqueue called %d times, want 0", downloader.resetAndEnqueueCalls)
	}
}

func TestServiceGetByIds(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	want := []chapters.Chapter{
		{ID: ids[0], Title: "Chapter 1", Number: 1, ComicID: uuid.New()},
		{ID: ids[1], Title: "Chapter 2", Number: 2, ComicID: uuid.New()},
	}
	repo := &fakeChaptersRepository{getByIdsResult: want}
	libraryRepo := &fakeLibraryRepository{existsByUserAndComic: true}
	svc := newTestService(repo, &fakeChapterDownloader{}, libraryRepo)

	got, err := svc.GetByIds(context.Background(), chapters.GetByIdsOpts{UserID: userID, IDs: ids})
	if err != nil {
		t.Fatalf("GetByIds: %v", err)
	}

	if repo.getByIdsCalls != 1 {
		t.Errorf("GetByIds called %d times, want 1", repo.getByIdsCalls)
	}

	if len(repo.lastGetByIds) != 2 || repo.lastGetByIds[0] != ids[0] || repo.lastGetByIds[1] != ids[1] {
		t.Errorf("GetByIds ids = %v, want %v", repo.lastGetByIds, ids)
	}

	if libraryRepo.lastUserID != userID {
		t.Errorf("ExistsByUserAndComic userID = %s, want %s", libraryRepo.lastUserID, userID)
	}

	if len(got) != 2 || got[0].ID != want[0].ID || got[1].Title != want[1].Title {
		t.Errorf("GetByIds() = %+v, want %+v", got, want)
	}
}

func TestServiceGetByIdsOmitsChaptersNotInLibrary(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	inLibraryComic := uuid.New()
	otherComic := uuid.New()
	keptID := uuid.New()
	omittedID := uuid.New()
	repo := &fakeChaptersRepository{
		getByIdsResult: []chapters.Chapter{
			{ID: keptID, ComicID: inLibraryComic, Title: "Kept"},
			{ID: omittedID, ComicID: otherComic, Title: "Omitted"},
			{ID: uuid.New(), ComicID: inLibraryComic, Title: "Also kept"},
		},
	}
	libraryRepo := &fakeLibraryRepository{
		inLibraryComicIDs: map[uuid.UUID]bool{inLibraryComic: true},
	}
	svc := newTestService(repo, &fakeChapterDownloader{}, libraryRepo)

	got, err := svc.GetByIds(context.Background(), chapters.GetByIdsOpts{
		UserID: userID,
		IDs:    []uuid.UUID{keptID, omittedID},
	})
	if err != nil {
		t.Fatalf("GetByIds: %v", err)
	}

	if len(got) != 2 || got[0].ID != keptID || got[1].Title != "Also kept" {
		t.Errorf("GetByIds() = %+v, want kept chapters only", got)
	}

	if libraryRepo.existsCalls != 2 {
		t.Errorf("ExistsByUserAndComic called %d times, want 2 (once per comic)", libraryRepo.existsCalls)
	}
}

func TestServiceGetByIdsError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("db down")
	repo := &fakeChaptersRepository{getByIdsErr: sentinel}
	svc := newTestService(repo, &fakeChapterDownloader{}, &fakeLibraryRepository{})

	got, err := svc.GetByIds(context.Background(), chapters.GetByIdsOpts{
		UserID: uuid.New(),
		IDs:    []uuid.UUID{uuid.New()},
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("GetByIds = %v, want wrapped sentinel", err)
	}

	if got != nil {
		t.Errorf("GetByIds returned %+v in addition to the error", got)
	}
}

func TestServiceGetByIdsLibraryError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("library down")
	repo := &fakeChaptersRepository{
		getByIdsResult: []chapters.Chapter{{ID: uuid.New(), ComicID: uuid.New()}},
	}
	svc := newTestService(repo, &fakeChapterDownloader{}, &fakeLibraryRepository{existsErr: sentinel})

	got, err := svc.GetByIds(context.Background(), chapters.GetByIdsOpts{
		UserID: uuid.New(),
		IDs:    []uuid.UUID{uuid.New()},
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("GetByIds = %v, want wrapped library sentinel", err)
	}

	if got != nil {
		t.Errorf("GetByIds returned %+v in addition to the error", got)
	}
}

func TestServiceListForLibraryReturnsDescendingIncludingLocked(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicID := uuid.New()
	lockedUntil := time.Date(2026, time.December, 1, 0, 0, 0, 0, time.UTC)
	low := chapters.Chapter{ID: uuid.New(), ComicID: comicID, Number: 1, Title: "One", Download: 100}
	locked := chapters.Chapter{
		ID: uuid.New(), ComicID: comicID, Number: 3, Title: "Locked", Download: 0, EarlyAccessUntil: &lockedUntil,
	}
	mid := chapters.Chapter{ID: uuid.New(), ComicID: comicID, Number: 2, Title: "Two", Download: 40}
	repo := &fakeChaptersRepository{listByComicIDResult: []chapters.Chapter{low, locked, mid}}
	libraryRepo := &fakeLibraryRepository{existsByUserAndComic: true}
	svc := newTestService(repo, &fakeChapterDownloader{}, libraryRepo)

	got, err := svc.ListForLibrary(context.Background(), chapters.ListForLibraryOpts{
		UserID:  userID,
		ComicID: comicID,
	})
	if err != nil {
		t.Fatalf("ListForLibrary: %v", err)
	}

	if libraryRepo.lastUserID != userID || libraryRepo.lastComicID != comicID {
		t.Errorf("library check user/comic = %s/%s, want %s/%s",
			libraryRepo.lastUserID, libraryRepo.lastComicID, userID, comicID)
	}

	if repo.listByComicIDCalls != 1 || repo.lastListByComicID != comicID {
		t.Errorf("ListByComicID comic = %s, calls = %d", repo.lastListByComicID, repo.listByComicIDCalls)
	}

	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}

	if got[0].Number != 3 || got[1].Number != 2 || got[2].Number != 1 {
		t.Errorf("order = %v, %v, %v, want 3, 2, 1", got[0].Number, got[1].Number, got[2].Number)
	}

	if got[0].EarlyAccessUntil == nil || !got[0].EarlyAccessUntil.Equal(lockedUntil) {
		t.Errorf("locked chapter missing earlyAccessUntil: %+v", got[0])
	}
}

func TestServiceListForLibraryEmpty(t *testing.T) {
	t.Parallel()

	svc := newTestService(
		&fakeChaptersRepository{},
		&fakeChapterDownloader{},
		&fakeLibraryRepository{existsByUserAndComic: true},
	)

	got, err := svc.ListForLibrary(context.Background(), chapters.ListForLibraryOpts{
		UserID:  uuid.New(),
		ComicID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("ListForLibrary: %v", err)
	}

	if got == nil || len(got) != 0 {
		t.Errorf("ListForLibrary() = %+v, want empty slice", got)
	}
}

func TestServiceListForLibraryForbiddenWhenComicExistsButNotInLibrary(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	lookup := &fakeComicLookup{exists: true}
	svc, err := chapters.NewService(chapters.Deps{
		Repository:        &fakeChaptersRepository{listByComicIDResult: []chapters.Chapter{{Number: 1}}},
		ChapterDownloader: &fakeChapterDownloader{},
		LibraryRepository: &fakeLibraryRepository{existsByUserAndComic: false},
		ComicLookup:       lookup,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	got, err := svc.ListForLibrary(context.Background(), chapters.ListForLibraryOpts{
		UserID:  uuid.New(),
		ComicID: comicID,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("ListForLibrary = %v, want domain.ErrForbidden", err)
	}

	if got != nil {
		t.Errorf("ListForLibrary returned %+v in addition to the error", got)
	}

	if lookup.calls != 1 || lookup.lastID != comicID {
		t.Errorf("ComicLookup.Exists comic = %s, calls = %d", lookup.lastID, lookup.calls)
	}
}

func TestServiceListForLibraryNotFoundWhenComicMissing(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	lookup := &fakeComicLookup{exists: false}
	svc, err := chapters.NewService(chapters.Deps{
		Repository:        &fakeChaptersRepository{},
		ChapterDownloader: &fakeChapterDownloader{},
		LibraryRepository: &fakeLibraryRepository{existsByUserAndComic: false},
		ComicLookup:       lookup,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	got, err := svc.ListForLibrary(context.Background(), chapters.ListForLibraryOpts{
		UserID:  uuid.New(),
		ComicID: comicID,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ListForLibrary = %v, want domain.ErrNotFound", err)
	}

	if got != nil {
		t.Errorf("ListForLibrary returned %+v in addition to the error", got)
	}

	if lookup.calls != 1 || lookup.lastID != comicID {
		t.Errorf("ComicLookup.Exists comic = %s, calls = %d", lookup.lastID, lookup.calls)
	}
}

func TestServiceListForLibraryDoesNotLookupComicWhenInLibrary(t *testing.T) {
	t.Parallel()

	lookup := &fakeComicLookup{exists: false}
	svc, err := chapters.NewService(chapters.Deps{
		Repository:        &fakeChaptersRepository{},
		ChapterDownloader: &fakeChapterDownloader{},
		LibraryRepository: &fakeLibraryRepository{existsByUserAndComic: true},
		ComicLookup:       lookup,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	_, err = svc.ListForLibrary(context.Background(), chapters.ListForLibraryOpts{
		UserID:  uuid.New(),
		ComicID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("ListForLibrary: %v", err)
	}

	if lookup.calls != 0 {
		t.Errorf("ComicLookup.Exists called %d times, want 0", lookup.calls)
	}
}

func TestServiceListForLibraryLibraryError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("library down")
	svc := newTestService(
		&fakeChaptersRepository{},
		&fakeChapterDownloader{},
		&fakeLibraryRepository{existsErr: sentinel},
	)

	got, err := svc.ListForLibrary(context.Background(), chapters.ListForLibraryOpts{
		UserID:  uuid.New(),
		ComicID: uuid.New(),
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("ListForLibrary = %v, want wrapped library sentinel", err)
	}

	if got != nil {
		t.Errorf("ListForLibrary returned %+v in addition to the error", got)
	}
}

func TestServiceListForLibraryListError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("db down")
	svc := newTestService(
		&fakeChaptersRepository{listByComicIDErr: sentinel},
		&fakeChapterDownloader{},
		&fakeLibraryRepository{existsByUserAndComic: true},
	)

	got, err := svc.ListForLibrary(context.Background(), chapters.ListForLibraryOpts{
		UserID:  uuid.New(),
		ComicID: uuid.New(),
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("ListForLibrary = %v, want wrapped list sentinel", err)
	}

	if got != nil {
		t.Errorf("ListForLibrary returned %+v in addition to the error", got)
	}
}

func TestServiceListForLibraryLookupError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("comics down")
	svc, err := chapters.NewService(chapters.Deps{
		Repository:        &fakeChaptersRepository{},
		ChapterDownloader: &fakeChapterDownloader{},
		LibraryRepository: &fakeLibraryRepository{existsByUserAndComic: false},
		ComicLookup:       &fakeComicLookup{existsErr: sentinel},
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	got, err := svc.ListForLibrary(context.Background(), chapters.ListForLibraryOpts{
		UserID:  uuid.New(),
		ComicID: uuid.New(),
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("ListForLibrary = %v, want wrapped lookup sentinel", err)
	}

	if got != nil {
		t.Errorf("ListForLibrary returned %+v in addition to the error", got)
	}
}
