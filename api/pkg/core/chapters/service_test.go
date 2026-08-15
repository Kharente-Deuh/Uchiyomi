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
	getByIDResult            *chapters.Chapter
	lastCreateMany           []chapters.CreateOpts
	listResumableResult      []chapters.Chapter
	listEarlyAccessResult    []chapters.Chapter
	createManyCalls          int
	listResumableCalls       int
	listEarlyAccessCalls     int
	getByIDCalls             int
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

func (f *fakeChaptersRepository) ListByComicID(context.Context, uuid.UUID) ([]chapters.Chapter, error) {
	panic("ListByComicID must not be called")
}

func (f *fakeChaptersRepository) ListResumable(context.Context) ([]chapters.Chapter, error) {
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
	lastUserID           uuid.UUID
	lastComicID          uuid.UUID
	existsByUserAndComic bool
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
	f.lastUserID = userID
	f.lastComicID = comicID

	return f.existsByUserAndComic, f.existsErr
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

func TestServiceEnqueueDownloadableRunsWithoutError(t *testing.T) {
	t.Parallel()

	downloader := &fakeChapterDownloader{}
	repo := &fakeChaptersRepository{}
	svc := newTestService(repo, downloader, &fakeLibraryRepository{existsByUserAndComic: true})

	now := time.Now()

	err := svc.EnqueueDownloadable(context.Background(), []chapters.Chapter{
		{Download: 100},
		{EarlyAccessUntil: now.Add(time.Hour)},
		{Download: 0, EarlyAccessUntil: now.Add(-time.Hour)},
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
	downloader := &fakeChapterDownloader{}
	repo := &fakeChaptersRepository{
		listEarlyAccessResult: []chapters.Chapter{
			{Download: 100},
			{EarlyAccessUntil: now.Add(time.Hour)},
			{Download: 0, EarlyAccessUntil: now.Add(-time.Hour)},
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
