// SPDX-License-Identifier: AGPL-3.0-or-later

package readingprogress

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
)

var _ ReadingProgressService = (*Service)(nil)

type Deps struct {
	Repository Repository
	Library    LibraryMembership
	Comics     ComicLookup
	Chapters   ChapterLookup
}

func (deps *Deps) Validate() error {
	if deps.Repository == nil {
		return errors.New("repository is required")
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

	rows, err := s.deps.Repository.ListByUserAndComic(ctx, opts)
	if err != nil {
		return ListResult{}, fmt.Errorf("s.deps.Repository.ListByUserAndComic: %w", err)
	}

	if len(rows) == 0 {
		return ListResult{Chapters: []Progress{}}, nil
	}

	ids := make([]uuid.UUID, len(rows))
	for i, row := range rows {
		ids[i] = row.ChapterID
	}

	chapterList, err := s.deps.Chapters.GetByIds(ctx, ids)
	if err != nil {
		return ListResult{}, fmt.Errorf("s.deps.Chapters.GetByIds: %w", err)
	}

	pagesNb := make(map[uuid.UUID]int, len(chapterList))
	for _, chapter := range chapterList {
		pagesNb[chapter.ID] = chapter.PagesNb
	}

	chapters := make([]Progress, len(rows))
	for i, row := range rows {
		chapters[i] = Progress{
			UpdatedAt: row.UpdatedAt,
			ChapterID: row.ChapterID,
			Page:      ClampPage(row.Page, pagesNb[row.ChapterID]),
		}
	}

	result := ListResult{Chapters: chapters}
	first := chapters[0]
	result.Continue = &Continue{
		ChapterID: first.ChapterID,
		Page:      first.Page,
	}

	return result, nil
}

func (s *Service) Save(ctx context.Context, opts SaveOpts) (Progress, error) {
	if err := s.requireLibrary(ctx, opts.UserID, opts.ComicID); err != nil {
		return Progress{}, err
	}

	chapter, err := s.deps.Chapters.GetByID(ctx, opts.ChapterID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return Progress{}, domain.ErrNotFound
		}

		return Progress{}, fmt.Errorf("s.deps.Chapters.GetByID: %w", err)
	}

	if chapter.ComicID != opts.ComicID {
		return Progress{}, domain.ErrNotFound
	}

	stored, err := s.deps.Repository.Get(ctx, GetOpts{
		UserID:    opts.UserID,
		ComicID:   opts.ComicID,
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
		ComicID:   opts.ComicID,
		ChapterID: opts.ChapterID,
		Page:      page,
	})
	if err != nil {
		return Progress{}, fmt.Errorf("s.deps.Repository.Upsert: %w", err)
	}

	return saved, nil
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
