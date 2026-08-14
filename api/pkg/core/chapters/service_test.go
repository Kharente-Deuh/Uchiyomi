// SPDX-License-Identifier: AGPL-3.0-or-later

package chapters_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

type fakeChaptersRepository struct {
	lastCreateMany  []chapters.CreateOpts
	createManyCalls int
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

func (f *fakeChaptersRepository) GetByID(context.Context, uuid.UUID) (*chapters.Chapter, error) {
	panic("GetByID must not be called")
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
	enqueueCalls         int
	cleanupComicCalls    int
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

func TestServiceCreateAll(t *testing.T) {
	t.Parallel()

	repo := &fakeChaptersRepository{}
	svc, err := chapters.NewService(chapters.Deps{
		Repository:        repo,
		ChapterDownloader: &fakeChapterDownloader{},
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

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
	svc, err := chapters.NewService(chapters.Deps{
		Repository:        repo,
		ChapterDownloader: downloader,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	now := time.Now()

	err = svc.EnqueueDownloadable(context.Background(), []chapters.Chapter{
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

func TestServiceCleanupComic(t *testing.T) {
	t.Parallel()

	downloader := &fakeChapterDownloader{}
	svc, err := chapters.NewService(chapters.Deps{
		Repository:        &fakeChaptersRepository{},
		ChapterDownloader: downloader,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	comicID := uuid.New()
	chapterList := []chapters.Chapter{
		{ID: uuid.New(), ComicID: comicID, Number: 1},
	}

	err = svc.CleanupComic(context.Background(), comicID, chapterList)
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
