// SPDX-License-Identifier: AGPL-3.0-or-later

package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
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

func (r *PGLibraryRepository) GetByID(ctx context.Context, id uuid.UUID) (*library.Entry, error) {
	model, err := r.db(ctx).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := r.modelToDomain(model)

	return &ret, nil
}

func (r *PGLibraryRepository) GetByUserAndComic(
	ctx context.Context,
	userID uuid.UUID,
	comicID uuid.UUID,
) (*library.Entry, error) {
	model, err := r.db(ctx).Where("user_id = ? AND comic_id = ?", userID, comicID).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := r.modelToDomain(model)

	return &ret, nil
}

func (r *PGLibraryRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]library.EntryWithComic, error) {
	var models []pgmodels.LibraryEntry

	err := pgtx.From(ctx, r.deps.DB).
		WithContext(ctx).
		Preload("Comic").
		Where("user_id = ?", userID).
		Order("added_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("pgtx.From(ctx, r.deps.DB).Find: %w", err)
	}

	ret := make([]library.EntryWithComic, 0, len(models))
	for _, model := range models {
		ret = append(ret, library.EntryWithComic{
			Entry: r.modelToDomain(model),
			Comic: r.comicModelToDomain(model.Comic),
		})
	}

	return ret, nil
}

func (r *PGLibraryRepository) Create(ctx context.Context, opts library.CreateEntryOpts) (*library.Entry, error) {
	addedAt := opts.AddedAt
	if addedAt.IsZero() {
		addedAt = time.Now()
	}

	model := &pgmodels.LibraryEntry{
		ID:      uuid.New(),
		UserID:  opts.UserID,
		ComicID: opts.ComicID,
		AddedAt: addedAt,
	}

	err := r.db(ctx).Create(ctx, model)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, domain.ErrAlreadyExists
		}

		return nil, fmt.Errorf("r.db(ctx).Create: %w", err)
	}

	ret := r.modelToDomain(*model)

	return &ret, nil
}

func (r *PGLibraryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	affected, err := r.db(ctx).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	if affected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *PGLibraryRepository) modelToDomain(model pgmodels.LibraryEntry) library.Entry {
	return library.Entry{
		ID:      model.ID,
		UserID:  model.UserID,
		ComicID: model.ComicID,
		AddedAt: model.AddedAt,
	}
}

func (r *PGLibraryRepository) comicModelToDomain(model pgmodels.Comic) comics.Comic {
	return comics.Comic{
		ID:               model.ID,
		Source:           model.Source,
		Slug:             model.Slug,
		Title:            model.Title,
		Status:           model.Status,
		Type:             model.ComicType,
		Genres:           model.Genres,
		ChapterCount:     model.ChapterCount,
		Author:           model.Author,
		Artist:           model.Artist,
		Description:      model.Description,
		AltTitles:        model.AltTitles,
		Rating:           model.Rating,
		ReleaseYear:      model.ReleaseYear,
		SourceURL:        model.SourceURL,
		ExternalCoverURL: model.ExternalCoverURL,
		LocalCoverPath:   model.LocalCoverPath,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
	}
}
