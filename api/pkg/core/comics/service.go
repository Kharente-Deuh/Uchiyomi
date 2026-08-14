// SPDX-License-Identifier: AGPL-3.0-or-later

package comics

import (
	"context"
	"errors"
	"fmt"

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
	Sources           sources.SourceMap
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
	var comic *Comic

	comic, err := s.deps.ComicsRepository.GetBySourceSlug(ctx, GetBySourceSlugOpts(opts))

	if err == nil {
		return comic, nil
	}

	if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("s.deps.ComicsRepository.GetBySourceSlug: %w", err)
	}

	var item *sources.GetInfosBySlugResponse

	src, ok := s.deps.Sources[opts.Source]
	if !ok {
		return nil, fmt.Errorf("source %s not found", opts.Source)
	}

	item, err = src.GetInfosBySlug(ctx, sources.GetInfosBySlugOpts{
		Slug:   opts.Slug,
		UserID: opts.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("src.GetInfosBySlug: %w", err)
	}

	err = s.deps.Transactor.WithinTx(ctx, transaction.TxOpts{}, func(ctx context.Context) error {
		created, err := s.deps.ComicsRepository.Create(ctx, CreateComicOpts{
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

		comic = created

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("s.deps.Transactor.WithinTx: %w", err)
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

func (s *Service) GetMany(ctx context.Context, opts GetManyOpts) ([]Comic, error) {
	comics, err := s.deps.ComicsRepository.GetMany(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("s.deps.ComicsRepository.GetMany: %w", err)
	}

	return comics, nil
}

func (s *Service) Delete(ctx context.Context, opts DeleteOpts) error {
	err := s.deps.Transactor.WithinTx(ctx, transaction.TxOpts{}, func(ctx context.Context) error {
		err := s.deps.LibraryRepository.Delete(ctx, library.DeleteOpts{
			UserID:  opts.UserID,
			ComicID: opts.ID,
		})
		if err != nil && errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("s.deps.LibraryRepository.Delete: %w", err)
		}

		err = s.deps.ComicsRepository.Delete(ctx, opts.ID)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("s.deps.ComicsRepository.Delete: %w", err)
		}

		return nil
	})

	//nolint:wrapcheck
	return err
}
