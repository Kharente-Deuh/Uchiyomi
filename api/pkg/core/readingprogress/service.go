// SPDX-License-Identifier: AGPL-3.0-or-later

package readingprogress

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
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

	chList, err := s.deps.Chapters.ListByComicID(ctx, opts.ComicID)
	if err != nil {
		return ListResult{}, fmt.Errorf("s.deps.Chapters.ListByComicID: %w", err)
	}

	if len(chList) == 0 {
		return ListResult{}, nil
	}

	progresses, err := s.deps.Repository.ListByUserAndComic(ctx, opts)
	if err != nil {
		return ListResult{}, fmt.Errorf("s.deps.Repository.ListByUserAndComic: %w", err)
	}

	if len(progresses) == 0 {
		return ListResult{}, nil
	}

	sort.SliceStable(chList, func(i, j int) bool {
		if chList[i].Number != chList[j].Number {
			return chList[i].Number < chList[j].Number
		}

		return bytes.Compare(chList[i].ID[:], chList[j].ID[:]) < 0
	})

	progMap := make(map[uuid.UUID]int, len(progresses))
	for _, p := range progresses {
		progMap[p.ChapterID] = p.Page
	}

	highestIdx := -1
	for i := range chList {
		page, ok := progMap[chList[i].ID]
		if !ok {
			continue
		}

		clamped := ClampPage(page, chList[i].PagesNb)
		if clamped > 1 || (chList[i].PagesNb > 0 && clamped >= chList[i].PagesNb) {
			highestIdx = i
		}
	}

	if highestIdx == -1 {
		return ListResult{}, nil
	}

	target := chList[highestIdx]
	targetPage := ClampPage(progMap[target.ID], target.PagesNb)
	if target.PagesNb > 0 && targetPage >= target.PagesNb && highestIdx+1 < len(chList) {
		next := chList[highestIdx+1]

		return ListResult{
			Continue: &Continue{
				ChapterID: next.ID,
				Page:      1,
			},
		}, nil
	}

	return ListResult{
		Continue: &Continue{
			ChapterID: target.ID,
			Page:      targetPage,
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

func (s *Service) SetRead(ctx context.Context, opts SetReadOpts) (ListResult, error) {
	ids := uniqueIDs(opts.ChapterIDs)
	if len(ids) == 0 {
		return ListResult{}, fmt.Errorf("%w: chapterIds is required", ErrInvalid)
	}

	if err := s.requireLibrary(ctx, opts.UserID, opts.ComicID); err != nil {
		return ListResult{}, err
	}

	found, err := s.deps.Chapters.GetByIds(ctx, ids)
	if err != nil {
		return ListResult{}, fmt.Errorf("s.deps.Chapters.GetByIds: %w", err)
	}

	if len(found) != len(ids) {
		return ListResult{}, domain.ErrNotFound
	}

	eligible := make([]chapters.Chapter, 0, len(found))
	for i := range found {
		if found[i].ComicID != opts.ComicID {
			return ListResult{}, domain.ErrNotFound
		}

		if found[i].PagesNb > 0 {
			eligible = append(eligible, found[i])
		}
	}

	if len(eligible) == 0 {
		return ListResult{}, fmt.Errorf("%w: no eligible chapters", ErrInvalid)
	}

	eligibleIDs := make([]uuid.UUID, len(eligible))
	for i, ch := range eligible {
		eligibleIDs[i] = ch.ID
	}

	if opts.Read {
		err = s.deps.Transactor.WithinTx(ctx, transaction.TxOpts{}, func(ctx context.Context) error {
			return s.markReadTx(ctx, opts.UserID, opts.ComicID, eligible)
		})
		if err != nil {
			return ListResult{}, fmt.Errorf("s.deps.Transactor.WithinTx: %w", err)
		}
	} else {
		err = s.deps.Repository.DeleteByUserAndChapterIDs(ctx, DeleteProgressOpts{
			UserID:     opts.UserID,
			ChapterIDs: eligibleIDs,
		})
		if err != nil {
			return ListResult{}, fmt.Errorf("s.deps.Repository.DeleteByUserAndChapterIDs: %w", err)
		}
	}

	return s.List(ctx, ListOpts{UserID: opts.UserID, ComicID: opts.ComicID})
}

func (s *Service) Delete(ctx context.Context, opts DeleteOpts) error {
	chapter, err := s.deps.Chapters.GetByID(ctx, opts.ChapterID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}

		return fmt.Errorf("s.deps.Chapters.GetByID: %w", err)
	}

	if err = s.requireLibrary(ctx, opts.UserID, chapter.ComicID); err != nil {
		return err
	}

	err = s.deps.Repository.DeleteByUserAndChapterIDs(ctx, DeleteProgressOpts{
		UserID:     opts.UserID,
		ChapterIDs: []uuid.UUID{opts.ChapterID},
	})
	if err != nil {
		return fmt.Errorf("s.deps.Repository.DeleteByUserAndChapterIDs: %w", err)
	}

	return nil
}

func (s *Service) markReadTx(
	ctx context.Context,
	userID, comicID uuid.UUID,
	eligible []chapters.Chapter,
) error {
	eligibleIDs := make([]uuid.UUID, len(eligible))
	for i, ch := range eligible {
		eligibleIDs[i] = ch.ID
	}

	existingProgress, err := s.deps.Repository.ListByUserAndChapterIDs(ctx, MapOpts{
		IDs:    eligibleIDs,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("s.deps.Repository.ListByUserAndChapterIDs: %w", err)
	}

	storedMap := make(map[uuid.UUID]int, len(existingProgress))
	for _, p := range existingProgress {
		storedMap[p.ChapterID] = p.Page
	}

	writtenAt := time.Now().UTC()
	for _, ch := range eligible {
		storedPage, ok := storedMap[ch.ID]
		if !ok || storedPage != ch.PagesNb {
			_, err := s.deps.Repository.Upsert(ctx, UpsertOpts{
				UpdatedAt: writtenAt,
				UserID:    userID,
				ComicID:   comicID,
				ChapterID: ch.ID,
				Page:      ch.PagesNb,
			})
			if err != nil {
				return fmt.Errorf("s.deps.Repository.Upsert: %w", err)
			}
		}
	}

	return nil
}

func uniqueIDs(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		out = append(out, id)
	}

	return out
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
