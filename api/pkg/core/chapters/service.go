// SPDX-License-Identifier: AGPL-3.0-or-later

package chapters

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/library"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
)

var _ ChaptersService = (*Service)(nil)

type Deps struct {
	Repository        ChaptersRepository
	ChapterDownloader ChapterDownloader
	LibraryRepository library.LibraryRepository
	ComicLookup       ComicLookup
}

func (deps *Deps) Validate() error {
	if deps.Repository == nil {
		return errors.New("repository is required")
	}

	if deps.ChapterDownloader == nil {
		return errors.New("chapterDownloader is required")
	}

	if deps.LibraryRepository == nil {
		return errors.New("libraryRepository is required")
	}

	if deps.ComicLookup == nil {
		return errors.New("comicLookup is required")
	}

	return nil
}

type Service struct {
	deps Deps
}

func NewService(deps Deps) (*Service, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	return &Service{deps: deps}, nil
}

func (s *Service) CreateAll(
	ctx context.Context,
	comicID uuid.UUID,
	srcChapters []sources.SourceChapter,
) ([]Chapter, error) {
	opts := make([]CreateOpts, len(srcChapters))
	for i, srcChapter := range srcChapters {
		opts[i] = CreateOpts{
			ComicID:           comicID,
			SourceChapterSlug: srcChapter.SourceChapterSlug,
			Number:            srcChapter.Number,
			Title:             srcChapter.Title,
			PagesNb:           srcChapter.PageCount,
			PublishedAt:       srcChapter.PublishedAt,
			EarlyAccessUntil:  utils.OptionalTime(srcChapter.EarlyAccessUntil),
		}
	}

	created, err := s.deps.Repository.CreateMany(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("s.deps.Repository.CreateMany: %w", err)
	}

	return created, nil
}

func (s *Service) ListByComicID(ctx context.Context, comicID uuid.UUID) ([]Chapter, error) {
	chapters, err := s.deps.Repository.ListByComicID(ctx, comicID)
	if err != nil {
		return nil, fmt.Errorf("s.deps.Repository.ListByComicID: %w", err)
	}

	return chapters, nil
}

func (s *Service) EnqueueDownloadable(ctx context.Context, chapters []Chapter) error {
	now := time.Now()
	downloadable := make([]Chapter, 0, len(chapters))

	for _, chapter := range chapters {
		if !isDownloadable(chapter, now) {
			continue
		}

		downloadable = append(downloadable, chapter)
	}

	if len(downloadable) == 0 {
		return nil
	}

	err := s.deps.ChapterDownloader.Enqueue(ctx, downloadable)
	if err != nil {
		return fmt.Errorf("s.deps.ChapterDownloader.Enqueue: %w", err)
	}

	return nil
}

func (s *Service) EnqueueResumable(ctx context.Context) error {
	chapterList, err := s.deps.Repository.ListResumable(ctx)
	if err != nil {
		return fmt.Errorf("s.deps.Repository.ListResumable: %w", err)
	}

	if len(chapterList) == 0 {
		return nil
	}

	err = s.deps.ChapterDownloader.Enqueue(ctx, chapterList)
	if err != nil {
		return fmt.Errorf("s.deps.ChapterDownloader.Enqueue: %w", err)
	}

	return nil
}

func (s *Service) ScanEarlyAccess(ctx context.Context) error {
	now := time.Now()

	chapterList, err := s.deps.Repository.ListEarlyAccessUnlocked(ctx, now)
	if err != nil {
		return fmt.Errorf("s.deps.Repository.ListEarlyAccessUnlocked: %w", err)
	}

	return s.EnqueueDownloadable(ctx, chapterList)
}

func (s *Service) CleanupComic(ctx context.Context, comicID uuid.UUID, chapterList []Chapter) error {
	err := s.deps.ChapterDownloader.CleanupComic(ctx, comicID, chapterList)
	if err != nil {
		return fmt.Errorf("s.deps.ChapterDownloader.CleanupComic: %w", err)
	}

	return nil
}

func (s *Service) RetryDownload(ctx context.Context, opts RetryDownloadOpts) error {
	chapter, err := s.deps.Repository.GetByID(ctx, opts.ChapterID)
	if err != nil {
		return fmt.Errorf("s.deps.Repository.GetByID: %w", err)
	}

	inLibrary, err := s.deps.LibraryRepository.ExistsByUserAndComic(ctx, opts.UserID, chapter.ComicID)
	if err != nil {
		return fmt.Errorf("s.deps.LibraryRepository.ExistsByUserAndComic: %w", err)
	}

	if !inLibrary {
		return domain.ErrForbidden
	}

	if chapter.Download >= 100 {
		return domain.ErrConflict
	}

	if chapter.Download == -1 {
		err = s.deps.ChapterDownloader.ResetAndEnqueue(ctx, opts.ChapterID)
		if err != nil {
			return fmt.Errorf("s.deps.ChapterDownloader.ResetAndEnqueue: %w", err)
		}

		return nil
	}

	err = s.deps.ChapterDownloader.Resume(ctx, opts.ChapterID)
	if err != nil {
		return fmt.Errorf("s.deps.ChapterDownloader.Resume: %w", err)
	}

	return nil
}

func isDownloadable(chapter Chapter, now time.Time) bool {
	if chapter.Download >= 100 {
		return false
	}

	if chapter.EarlyAccessUntil != nil && chapter.EarlyAccessUntil.After(now) {
		return false
	}

	return true
}

func (s *Service) ListForLibrary(ctx context.Context, opts ListForLibraryOpts) ([]Chapter, error) {
	inLibrary, err := s.deps.LibraryRepository.ExistsByUserAndComic(ctx, opts.UserID, opts.ComicID)
	if err != nil {
		return nil, fmt.Errorf("s.deps.LibraryRepository.ExistsByUserAndComic: %w", err)
	}

	if !inLibrary {
		exists, lookupErr := s.deps.ComicLookup.Exists(ctx, opts.ComicID)
		if lookupErr != nil {
			return nil, fmt.Errorf("s.deps.ComicLookup.Exists: %w", lookupErr)
		}

		if !exists {
			return nil, domain.ErrNotFound
		}

		return nil, domain.ErrForbidden
	}

	chapterList, err := s.deps.Repository.ListByComicID(ctx, opts.ComicID)
	if err != nil {
		return nil, fmt.Errorf("s.deps.Repository.ListByComicID: %w", err)
	}

	slices.SortFunc(chapterList, func(a, b Chapter) int {
		return cmp.Compare(b.Number, a.Number)
	})

	if chapterList == nil {
		chapterList = []Chapter{}
	}

	return chapterList, nil
}

func (s *Service) GetByIds(ctx context.Context, opts GetByIdsOpts) ([]Chapter, error) {
	chapterList, err := s.deps.Repository.GetByIds(ctx, opts.IDs)
	if err != nil {
		return nil, fmt.Errorf("s.deps.Repository.GetByIds: %w", err)
	}

	accessibleComics := make(map[uuid.UUID]bool)
	accessible := make([]Chapter, 0, len(chapterList))

	for _, chapter := range chapterList {
		inLibrary, known := accessibleComics[chapter.ComicID]
		if !known {
			inLibrary, err = s.deps.LibraryRepository.ExistsByUserAndComic(ctx, opts.UserID, chapter.ComicID)
			if err != nil {
				return nil, fmt.Errorf("s.deps.LibraryRepository.ExistsByUserAndComic: %w", err)
			}

			accessibleComics[chapter.ComicID] = inLibrary
		}

		if !inLibrary {
			continue
		}

		accessible = append(accessible, chapter)
	}

	return accessible, nil
}
