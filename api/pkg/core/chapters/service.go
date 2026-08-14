// SPDX-License-Identifier: AGPL-3.0-or-later

package chapters

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

var _ ChaptersService = (*Service)(nil)

type Deps struct {
	Repository        ChaptersRepository
	ChapterDownloader ChapterDownloader
}

func (deps *Deps) Validate() error {
	if deps.Repository == nil {
		return errors.New("repository is required")
	}

	if deps.ChapterDownloader == nil {
		return errors.New("chapterDownloader is required")
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
			EarlyAccessUntil:  srcChapter.EarlyAccessUntil,
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

func isDownloadable(chapter Chapter, now time.Time) bool {
	if chapter.Download >= 100 {
		return false
	}

	if chapter.EarlyAccessUntil.After(now) {
		return false
	}

	return true
}
