// SPDX-License-Identifier: AGPL-3.0-or-later

package comics

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/logging"
)

func (s *Service) logger() *slog.Logger {
	if s.deps.Logger == nil {
		return slog.Default()
	}

	return s.deps.Logger
}

func (s *Service) RefreshChapterLists(ctx context.Context) error {
	names := make([]sources.SourceName, 0, len(s.deps.Sources))
	for name := range s.deps.Sources {
		names = append(names, name)
	}

	slices.Sort(names)

	for _, name := range names {
		if err := s.refreshSource(ctx, name); err != nil {
			s.logger().ErrorContext(ctx, "chapter list refresh failed for source",
				"source", name, logging.Err(err))
		}
	}

	return nil
}

func (s *Service) refreshSource(ctx context.Context, name sources.SourceName) error {
	src, ok := s.deps.Sources[name]
	if !ok || src == nil {
		return fmt.Errorf("source %s not found", name)
	}

	list, err := s.deps.ComicsRepository.ListByStatuses(ctx, ListByStatusesOpts{
		Source: name,
		Statuses: []sources.SeriesStatus{
			sources.SeriesStatusOngoing,
			sources.SeriesStatusHiatus,
		},
	})
	if err != nil {
		return fmt.Errorf("s.deps.ComicsRepository.ListByStatuses: %w", err)
	}

	for i := range list {
		comic := list[i]
		if err := s.refreshComic(ctx, src, comic); err != nil {
			s.logger().ErrorContext(ctx, "chapter list refresh failed for comic",
				"source", name, "slug", comic.Slug, logging.Err(err))
		}
	}

	return nil
}

func isPollable(status sources.SeriesStatus) bool {
	return status == sources.SeriesStatusOngoing || status == sources.SeriesStatusHiatus
}

func (s *Service) RefreshComic(ctx context.Context, opts RefreshComicOpts) (*Comic, error) {
	comic, err := s.deps.ComicsRepository.FindByID(ctx, opts.ID)
	if err != nil {
		return nil, fmt.Errorf("s.deps.ComicsRepository.FindByID: %w", err)
	}

	inLibrary, err := s.deps.LibraryRepository.ExistsByUserAndComic(ctx, opts.UserID, comic.ID)
	if err != nil {
		return nil, fmt.Errorf("s.deps.LibraryRepository.ExistsByUserAndComic: %w", err)
	}

	if !inLibrary {
		return nil, domain.ErrForbidden
	}

	if !isPollable(comic.Status) {
		return nil, domain.ErrConflict
	}

	return comic, nil
}

func (s *Service) refreshComic(ctx context.Context, src sources.Source, comic Comic) error {
	infos, err := src.GetInfosBySlug(ctx, sources.GetInfosBySlugOpts{
		Slug:  comic.Slug,
		Fresh: true,
	})
	if err != nil {
		return fmt.Errorf("src.GetInfosBySlug: %w", err)
	}

	srcChapters, err := src.GetChaptersBySlug(ctx, sources.GetChaptersBySlugOpts{
		Slug:  comic.Slug,
		Fresh: true,
	})
	if err != nil {
		return fmt.Errorf("src.GetChaptersBySlug: %w", err)
	}

	existing, err := s.deps.ChaptersService.ListByComicID(ctx, comic.ID)
	if err != nil {
		return fmt.Errorf("s.deps.ChaptersService.ListByComicID: %w", err)
	}

	missing := missingSourceChapters(existing, srcChapters)
	if len(missing) > 0 {
		created, err := s.deps.ChaptersService.CreateAll(ctx, comic.ID, missing)
		if err != nil {
			return fmt.Errorf("s.deps.ChaptersService.CreateAll: %w", err)
		}

		err = s.deps.ChaptersService.EnqueueDownloadable(ctx, created)
		if err != nil {
			return fmt.Errorf("s.deps.ChaptersService.EnqueueDownloadable: %w", err)
		}
	}

	err = s.deps.ComicsRepository.UpdateStatusAndChapterCount(ctx, UpdateStatusAndChapterCountOpts{
		ID:           comic.ID,
		Status:       infos.Status,
		ChapterCount: infos.ChapterCount,
	})
	if err != nil {
		return fmt.Errorf("s.deps.ComicsRepository.UpdateStatusAndChapterCount: %w", err)
	}

	return nil
}

func missingSourceChapters(
	existing []chapters.Chapter,
	srcChapters []sources.SourceChapter,
) []sources.SourceChapter {
	have := make(map[string]struct{}, len(existing))
	for _, chapter := range existing {
		have[chapter.SourceChapterSlug] = struct{}{}
	}

	var missing []sources.SourceChapter

	for _, srcChapter := range srcChapters {
		if _, ok := have[srcChapter.SourceChapterSlug]; ok {
			continue
		}

		missing = append(missing, srcChapter)
	}

	return missing
}
