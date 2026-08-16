// SPDX-License-Identifier: AGPL-3.0-or-later

package comics

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/library"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
)

var _ ComicsService = (*Service)(nil)

type Deps struct {
	ComicsRepository  ComicsRepository
	Transactor        transaction.Transactor
	LibraryRepository library.LibraryRepository
	ChaptersService   chapters.ChaptersService
	LocalCoverStore   LocalCoverStore
	Sources           sources.SourceMap
	Logger            *slog.Logger
}

func (deps *Deps) Validate() error {
	if deps.ComicsRepository == nil {
		return errors.New("comics repository is required")
	}

	if deps.Transactor == nil {
		return errors.New("transactor is required")
	}

	if deps.LibraryRepository == nil {
		return errors.New("library repository is required")
	}

	if deps.ChaptersService == nil {
		return errors.New("chapters service is required")
	}

	if deps.LocalCoverStore == nil {
		return errors.New("local cover store is required")
	}

	if deps.Sources == nil {
		return errors.New("sources is required")
	}

	for _, source := range deps.Sources {
		if source == nil {
			return errors.New("source is required")
		}
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

func (s *Service) Create(ctx context.Context, opts CreateOpts) (*Comic, error) {
	comic, err := s.deps.ComicsRepository.FindBySourceSlug(ctx, FindBySourceSlugOpts{
		Source: opts.Source,
		Slug:   opts.Slug,
	})
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("s.deps.ComicsRepository.FindBySourceSlug: %w", err)
	}

	if comic != nil {
		return s.addExistingComicToLibrary(ctx, opts, comic)
	}

	return s.createNewComic(ctx, opts)
}

func (s *Service) addExistingComicToLibrary(ctx context.Context, opts CreateOpts, comic *Comic) (*Comic, error) {
	err := s.deps.Transactor.WithinTx(ctx, transaction.TxOpts{}, func(ctx context.Context) error {
		_, err := s.deps.LibraryRepository.Create(ctx, library.CreateOpts{
			UserID:  opts.UserID,
			ComicID: comic.ID,
		})
		if err != nil && !errors.Is(err, domain.ErrAlreadyExists) {
			return fmt.Errorf("s.deps.LibraryRepository.Create: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("s.deps.Transactor.WithinTx: %w", err)
	}

	chapterList, err := s.deps.ChaptersService.ListByComicID(ctx, comic.ID)
	if err != nil {
		return nil, fmt.Errorf("s.deps.ChaptersService.ListByComicID: %w", err)
	}

	err = s.deps.ChaptersService.EnqueueDownloadable(ctx, chapterList)
	if err != nil {
		return nil, fmt.Errorf("s.deps.ChaptersService.EnqueueDownloadable: %w", err)
	}

	return comic, nil
}

func (s *Service) createNewComic(ctx context.Context, opts CreateOpts) (*Comic, error) {
	src, ok := s.deps.Sources[opts.Source]
	if !ok {
		return nil, fmt.Errorf("source %s not found", opts.Source)
	}

	item, err := src.GetInfosBySlug(ctx, sources.GetInfosBySlugOpts{
		Slug:   opts.Slug,
		UserID: opts.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("src.GetInfosBySlug: %w", err)
	}

	srcChapters, err := src.GetChaptersBySlug(ctx, sources.GetChaptersBySlugOpts{Slug: opts.Slug})
	if err != nil {
		return nil, fmt.Errorf("src.GetChaptersBySlug: %w", err)
	}

	id := uuid.New()

	if err = s.deps.LocalCoverStore.ObtainLocal(ctx, id, string(opts.Source), opts.Slug); err != nil {
		return nil, fmt.Errorf("s.deps.LocalCoverStore.ObtainLocal: %w", err)
	}

	var comic *Comic

	err = s.deps.Transactor.WithinTx(ctx, transaction.TxOpts{}, func(ctx context.Context) error {
		created, err := s.deps.ComicsRepository.Create(ctx, CreateComicOpts{
			ID:           id,
			Status:       item.Status,
			Type:         item.Type,
			Description:  item.Description,
			Source:       opts.Source,
			Artist:       item.Artist,
			Slug:         item.Slug,
			Author:       item.Author,
			Title:        item.Title,
			AltTitles:    item.AltTitles,
			Genres:       item.Genres,
			ChapterCount: item.ChapterCount,
		})
		if err != nil {
			return fmt.Errorf("s.deps.ComicsRepository.Create: %w", err)
		}

		_, err = s.deps.LibraryRepository.Create(ctx, library.CreateOpts{
			UserID:  opts.UserID,
			ComicID: created.ID,
		})
		if err != nil {
			return fmt.Errorf("s.deps.LibraryRepository.Create: %w", err)
		}

		_, err = s.deps.ChaptersService.CreateAll(ctx, created.ID, srcChapters)
		if err != nil {
			return fmt.Errorf("s.deps.ChaptersService.CreateAll: %w", err)
		}

		comic = created

		return nil
	})
	if err != nil {
		if remErr := s.deps.LocalCoverStore.RemoveLocal(id); remErr != nil {
			if s.deps.Logger != nil {
				s.deps.Logger.ErrorContext(ctx, "failed to remove orphan cover", "comic_id", id, "error", remErr)
			}
		}

		if errors.Is(err, domain.ErrAlreadyExists) {
			winner, findErr := s.deps.ComicsRepository.FindBySourceSlug(ctx, FindBySourceSlugOpts{
				Source: opts.Source,
				Slug:   opts.Slug,
			})
			if findErr != nil {
				return nil, fmt.Errorf("s.deps.ComicsRepository.FindBySourceSlug: %w", findErr)
			}

			return s.addExistingComicToLibrary(ctx, opts, winner)
		}

		return nil, fmt.Errorf("s.deps.Transactor.WithinTx: %w", err)
	}

	chapterList, err := s.deps.ChaptersService.ListByComicID(ctx, comic.ID)
	if err != nil {
		return nil, fmt.Errorf("s.deps.ChaptersService.ListByComicID: %w", err)
	}

	err = s.deps.ChaptersService.EnqueueDownloadable(ctx, chapterList)
	if err != nil {
		return nil, fmt.Errorf("s.deps.ChaptersService.EnqueueDownloadable: %w", err)
	}

	return comic, nil
}

func (s *Service) GetByID(ctx context.Context, opts GetByIDOpts) (*Comic, error) {
	comic, err := s.deps.ComicsRepository.GetByID(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("s.deps.ComicsRepository.GetByID: %w", err)
	}

	return comic, nil
}

func (s *Service) ServeCover(ctx context.Context, opts GetByIDOpts) (string, string, error) {
	comic, err := s.GetByID(ctx, opts)
	if err != nil {
		return "", "", fmt.Errorf("s.GetByID: %w", err)
	}

	path, contentType, err := s.deps.LocalCoverStore.ServeLocal(ctx, comic.ID)
	if err != nil {
		return "", "", fmt.Errorf("s.deps.LocalCoverStore.ServeLocal: %w", err)
	}

	return path, contentType, nil
}

func (s *Service) GetMany(ctx context.Context, opts GetManyOpts) ([]Comic, error) {
	comics, err := s.deps.ComicsRepository.GetMany(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("s.deps.ComicsRepository.GetMany: %w", err)
	}

	return comics, nil
}

func (s *Service) Delete(ctx context.Context, opts DeleteOpts) error {
	var comicDeleted bool
	var chaptersToCleanup []chapters.Chapter

	err := s.deps.Transactor.WithinTx(ctx, transaction.TxOpts{}, func(ctx context.Context) error {
		err := s.deps.LibraryRepository.Delete(ctx, library.DeleteOpts{
			UserID:  opts.UserID,
			ComicID: opts.ID,
		})
		if err != nil {
			return fmt.Errorf("s.deps.LibraryRepository.Delete: %w", err)
		}

		hasEntries, err := s.deps.LibraryRepository.ExistsByComicID(ctx, opts.ID)
		if err != nil {
			return fmt.Errorf("s.deps.LibraryRepository.ExistsByComicID: %w", err)
		}

		if hasEntries {
			return nil
		}

		chapterList, err := s.deps.ChaptersService.ListByComicID(ctx, opts.ID)
		if err != nil {
			return fmt.Errorf("s.deps.ChaptersService.ListByComicID: %w", err)
		}

		err = s.deps.ComicsRepository.Delete(ctx, opts.ID)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("s.deps.ComicsRepository.Delete: %w", err)
		}

		chaptersToCleanup = chapterList
		comicDeleted = true

		return nil
	})
	if err != nil {
		//nolint:wrapcheck
		return err
	}

	if comicDeleted {
		err = s.deps.ChaptersService.CleanupComic(ctx, opts.ID, chaptersToCleanup)
		if err != nil {
			return fmt.Errorf("s.deps.ChaptersService.CleanupComic: %w", err)
		}
	}

	return nil
}
