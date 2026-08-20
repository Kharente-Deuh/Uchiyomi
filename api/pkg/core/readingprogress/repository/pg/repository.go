// SPDX-License-Identifier: AGPL-3.0-or-later

package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readingprogress"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgmodels"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ readingprogress.Repository = (*PGReadingProgressRepository)(nil)

type Deps struct {
	DB *gorm.DB
}

func (deps *Deps) Validate() error {
	if deps.DB == nil {
		return errors.New("db is required")
	}

	return nil
}

type PGReadingProgressRepository struct {
	deps Deps
}

func New(deps Deps) (*PGReadingProgressRepository, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	r := &PGReadingProgressRepository{deps: deps}

	return r, nil
}

func (r *PGReadingProgressRepository) db(ctx context.Context) gorm.Interface[pgmodels.ReadingProgress] {
	return gorm.G[pgmodels.ReadingProgress](pgtx.From(ctx, r.deps.DB))
}

func (r *PGReadingProgressRepository) ListByUserAndComic(
	ctx context.Context, opts readingprogress.ListOpts,
) ([]readingprogress.Progress, error) {
	models, err := r.db(ctx).
		Joins(clause.JoinTarget{Association: "LibraryEntry"}, nil).
		Where("library_entries.user_id = ? AND library_entries.comic_id = ?", opts.UserID, opts.ComicID).
		Order("reading_progress.updated_at DESC").
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("r.db(ctx).Find: %w", err)
	}

	ret := make([]readingprogress.Progress, 0, len(models))
	for i := range models {
		ret = append(ret, models[i].Domain())
	}

	return ret, nil
}

func (r *PGReadingProgressRepository) Get(
	ctx context.Context, opts readingprogress.GetOpts,
) (*readingprogress.Progress, error) {
	model, err := r.db(ctx).
		Joins(clause.JoinTarget{Association: "LibraryEntry"}, nil).
		Where(
			"library_entries.user_id = ? AND library_entries.comic_id = ? AND reading_progress.chapter_id = ?",
			opts.UserID, opts.ComicID, opts.ChapterID,
		).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.db(ctx).First: %w", err)
	}

	ret := model.Domain()

	return &ret, nil
}

func (r *PGReadingProgressRepository) Upsert(
	ctx context.Context, opts readingprogress.UpsertOpts,
) (readingprogress.Progress, error) {
	entry, err := gorm.G[pgmodels.LibraryEntry](pgtx.From(ctx, r.deps.DB)).
		Where("user_id = ? AND comic_id = ?", opts.UserID, opts.ComicID).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return readingprogress.Progress{}, domain.ErrNotFound
		}

		return readingprogress.Progress{}, fmt.Errorf("gorm.G[pgmodels.LibraryEntry].First: %w", err)
	}

	model := &pgmodels.ReadingProgress{
		ID:             uuid.New(),
		LibraryEntryID: entry.ID,
		ChapterID:      opts.ChapterID,
		Page:           opts.Page,
		UpdatedAt:      opts.UpdatedAt,
	}

	err = pgtx.From(ctx, r.deps.DB).WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "library_entry_id"}, {Name: "chapter_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"page", "updated_at"}),
	}).Create(model).Error
	if err != nil {
		return readingprogress.Progress{}, fmt.Errorf("pgtx.From(ctx, r.deps.DB).Create: %w", err)
	}

	return model.Domain(), nil
}
