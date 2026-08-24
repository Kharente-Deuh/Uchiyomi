// SPDX-License-Identifier: AGPL-3.0-or-later

package readingprogress

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
)

var _ ReadingProgressService = (*Service)(nil)

type Deps struct {
	Repository Repository
	Transactor transaction.Transactor
	Library    LibraryMembership
	Comics     ComicLookup
	Chapters   ChapterLookup
}

func (deps *Deps) Validate() error {
	if deps.Repository == nil {
		return errors.New("repository is required")
	}

	if deps.Transactor == nil {
		return errors.New("transactor is required")
	}

	if deps.Library == nil {
		return errors.New("library is required")
	}

	if deps.Comics == nil {
		return errors.New("comics is required")
	}

	if deps.Chapters == nil {
		return errors.New("chapters is required")
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

func (s *Service) List(ctx context.Context, opts ListOpts) (ListResult, error) {
	if err := s.requireLibrary(ctx, opts.UserID, opts.ComicID); err != nil {
		return ListResult{}, err
	}

	row, err := s.deps.Repository.GetLatestByUserAndComic(ctx, opts)
	if err != nil {
		return ListResult{}, fmt.Errorf("s.deps.Repository.GetLatestByUserAndComic: %w", err)
	}

	if row == nil {
		return ListResult{}, nil
	}

	chapter, err := s.deps.Chapters.GetByID(ctx, row.ChapterID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ListResult{
				Continue: &Continue{ChapterID: row.ChapterID, Page: row.Page},
			}, nil
		}

		return ListResult{}, fmt.Errorf("s.deps.Chapters.GetByID: %w", err)
	}

	return ListResult{
		Continue: &Continue{
			ChapterID: row.ChapterID,
			Page:      ClampPage(row.Page, chapter.PagesNb),
		},
	}, nil
}

func (s *Service) MapByChapterIDs(ctx context.Context, opts MapOpts) (map[uuid.UUID]Progress, error) {
	if len(opts.IDs) == 0 {
		return map[uuid.UUID]Progress{}, nil
	}

	rows, err := s.deps.Repository.ListByUserAndChapterIDs(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("s.deps.Repository.ListByUserAndChapterIDs: %w", err)
	}

	out := make(map[uuid.UUID]Progress, len(rows))
	for _, row := range rows {
		out[row.ChapterID] = row
	}

	return out, nil
}

func (s *Service) Save(ctx context.Context, opts SaveOpts) (Progress, error) {
	chapter, err := s.deps.Chapters.GetByID(ctx, opts.ChapterID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return Progress{}, domain.ErrNotFound
		}

		return Progress{}, fmt.Errorf("s.deps.Chapters.GetByID: %w", err)
	}

	if err = s.requireLibrary(ctx, opts.UserID, chapter.ComicID); err != nil {
		return Progress{}, err
	}

	stored, err := s.deps.Repository.Get(ctx, GetOpts{
		UserID:    opts.UserID,
		ComicID:   chapter.ComicID,
		ChapterID: opts.ChapterID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			stored = nil
		} else {
			return Progress{}, fmt.Errorf("s.deps.Repository.Get: %w", err)
		}
	}

	var storedPage *int
	if stored != nil {
		page := stored.Page
		storedPage = &page
	}

	page, err := MergePage(storedPage, opts.Page, chapter.PagesNb)
	if err != nil {
		return Progress{}, err
	}

	saved, err := s.deps.Repository.Upsert(ctx, UpsertOpts{
		UpdatedAt: time.Now().UTC(),
		UserID:    opts.UserID,
		ComicID:   chapter.ComicID,
		ChapterID: opts.ChapterID,
		Page:      page,
	})
	if err != nil {
		return Progress{}, fmt.Errorf("s.deps.Repository.Upsert: %w", err)
	}

	return saved, nil
}

func (s *Service) MarkRead(ctx context.Context, opts MarkReadOpts) (ListResult, error) {
	_ = ctx
	_ = opts

	return ListResult{}, fmt.Errorf("%w: not implemented", ErrInvalid)
}

func (s *Service) requireLibrary(ctx context.Context, userID, comicID uuid.UUID) error {
	inLibrary, err := s.deps.Library.ExistsByUserAndComic(ctx, userID, comicID)
	if err != nil {
		return fmt.Errorf("s.deps.Library.ExistsByUserAndComic: %w", err)
	}

	if !inLibrary {
		exists, lookupErr := s.deps.Comics.Exists(ctx, comicID)
		if lookupErr != nil {
			return fmt.Errorf("s.deps.Comics.Exists: %w", lookupErr)
		}

		if !exists {
			return domain.ErrNotFound
		}

		return domain.ErrForbidden
	}

	return nil
}
