// SPDX-License-Identifier: AGPL-3.0-or-later

package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/library"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgmodels"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
	"gorm.io/gorm"
)

var _ library.LibraryRepository = (*PGLibraryRepository)(nil)

type Deps struct {
	DB *gorm.DB
}

func (deps *Deps) Validate() error {
	if deps.DB == nil {
		return errors.New("db is required")
	}

	return nil
}

type PGLibraryRepository struct {
	deps Deps
}

func New(deps Deps) (*PGLibraryRepository, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	r := &PGLibraryRepository{deps: deps}

	return r, nil
}

func (r *PGLibraryRepository) db(ctx context.Context) gorm.Interface[pgmodels.LibraryEntry] {
	return gorm.G[pgmodels.LibraryEntry](pgtx.From(ctx, r.deps.DB))
}

func (r *PGLibraryRepository) Create(ctx context.Context, opts library.CreateOpts) (*library.Entry, error) {
	model := &pgmodels.LibraryEntry{
		ID:      uuid.New(),
		UserID:  opts.UserID,
		ComicID: opts.ComicID,
		AddedAt: time.Now(),
	}

	err := r.db(ctx).Create(ctx, model)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, domain.ErrAlreadyExists
		}

		return nil, fmt.Errorf("r.db(ctx).Create: %w", err)
	}

	ret := model.Domain()

	return &ret, nil
}

func (r *PGLibraryRepository) Delete(ctx context.Context, opts library.DeleteOpts) error {
	affected, err := r.db(ctx).Where("user_id = ? AND comic_id = ?", opts.UserID, opts.ComicID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	if affected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *PGLibraryRepository) ExistsByComicID(ctx context.Context, comicID uuid.UUID) (bool, error) {
	count, err := r.db(ctx).Where("comic_id = ?", comicID).Count(ctx, "*")
	if err != nil {
		return false, fmt.Errorf("r.db(ctx).Count: %w", err)
	}

	return count > 0, nil
}

// func (r *PGLibraryRepository) GetByID(ctx context.Context, id uuid.UUID) (*library.Entry, error) {
// 	model, err := r.db(ctx).Where("id = ?", id).First(ctx)
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, domain.ErrNotFound
// 		}

// 		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
// 	}

// 	ret := model.Domain()

// 	return &ret, nil
// }

// //nolint:lll
// func (r *PGLibraryRepository) GetByIDByUser(ctx context.Context, opts library.GetByIDByUserOpts) (*library.Entry, error) {
// 	model, err := r.db(ctx).Where("user_id = ? AND id = ?", opts.UserID, opts.ID).First(ctx)
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, domain.ErrNotFound
// 		}

// 		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
// 	}

// 	ret := model.Domain()

// 	return &ret, nil
// }

// //nolint:lll
// func (r *PGLibraryRepository) GetBySlugByUserWithComic(ctx context.Context, opts library.GetBySlugByUserOpts) (*library.EntryWithComic, error) {
// 	model, err := r.db(ctx).Joins(clause.JoinTarget{Association: "Comic"}, nil).Where(
// 		"library_entries.user_id = ? AND comics.slug = ? AND comics.source = ?",
// 		opts.UserID, opts.Slug, opts.Source,
// 	).First(ctx)

// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, domain.ErrNotFound
// 		}

// 		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
// 	}

// 	return &library.EntryWithComic{
// 		Entry: model.Domain(),
// 		Comic: model.Comic.Domain(),
// 	}, nil
// }

// func (r *PGLibraryRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]library.EntryWithComic, error) {
// 	var models []pgmodels.LibraryEntry

// 	err := pgtx.From(ctx, r.deps.DB).
// 		WithContext(ctx).
// 		Preload("Comic").
// 		Where("user_id = ?", userID).
// 		Order("added_at DESC").
// 		Find(&models).Error
// 	if err != nil {
// 		return nil, fmt.Errorf("pgtx.From(ctx, r.deps.DB).Find: %w", err)
// 	}

// 	ret := make([]library.EntryWithComic, 0, len(models))
// 	for _, model := range models {
// 		ret = append(ret, library.EntryWithComic{
// 			Entry: model.Domain(),
// 			Comic: model.Comic.Domain(),
// 		})
// 	}

// 	return ret, nil
// }

// func (r *PGLibraryRepository) Create(ctx context.Context, opts library.CreateEntryOpts) (*library.Entry, error) {
// 	addedAt := opts.AddedAt
// 	if addedAt.IsZero() {
// 		addedAt = time.Now()
// 	}

// 	model := &pgmodels.LibraryEntry{
// 		ID:      uuid.New(),
// 		UserID:  opts.UserID,
// 		ComicID: opts.ComicID,
// 		AddedAt: addedAt,
// 	}

// 	err := r.db(ctx).Create(ctx, model)
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrDuplicatedKey) {
// 			return nil, domain.ErrAlreadyExists
// 		}

// 		return nil, fmt.Errorf("r.db(ctx).Create: %w", err)
// 	}

// 	ret := model.Domain()

// 	return &ret, nil
// }

// func (r *PGLibraryRepository) Delete(ctx context.Context, id uuid.UUID) error {
// 	affected, err := r.db(ctx).Where("id = ?", id).Delete(ctx)
// 	if err != nil {
// 		return fmt.Errorf("r.db(ctx).Where: %w", err)
// 	}

// 	if affected == 0 {
// 		return domain.ErrNotFound
// 	}

// 	return nil
// }

// func (r *PGLibraryRepository) GetByComicID(ctx context.Context, comicID uuid.UUID) ([]library.Entry, error) {
// 	models, err := r.db(ctx).Where("comic_id = ?", comicID).Find(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
// 	}

// 	ret := make([]library.Entry, 0, len(models))
// 	for _, model := range models {
// 		ret = append(ret, model.Domain())
// 	}

// 	return ret, nil
// }

// func (r *PGLibraryRepository) ListComics(ctx context.Context, opts library.ListEntriesOpts) ([]comics.Comic, error) {
// 	query := r.db(ctx).Joins(clause.JoinTarget{Association: "Comic"}, nil).Where("user_id = ?", opts.UserID)
// 	if opts.Source != nil {
// 		query = query.Where("comic.source = ?", opts.Source)
// 	}

// 	if opts.Type != nil {
// 		query = query.Where("comic.type = ?", opts.Type)
// 	}

// 	if opts.Status != nil {
// 		query = query.Where("comic.status = ?", opts.Status)
// 	}

// 	query = query.Offset(opts.Offset).Limit(opts.Limit).Order("added_at DESC")

// 	models, err := query.Find(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
// 	}

// 	ret := make([]comics.Comic, 0, len(models))
// 	for _, model := range models {
// 		ret = append(ret, model.Comic.Domain())
// 	}

// 	return ret, nil
// }
