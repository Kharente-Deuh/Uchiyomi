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
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgmodels"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

var _ comics.ComicsRepository = (*PGComicsRepository)(nil)

type Deps struct {
	DB *gorm.DB
}

func (deps *Deps) Validate() error {
	if deps.DB == nil {
		return errors.New("db is required")
	}

	return nil
}

type PGComicsRepository struct {
	deps Deps
}

func New(deps Deps) (*PGComicsRepository, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	r := &PGComicsRepository{deps: deps}

	return r, nil
}

func (r *PGComicsRepository) db(ctx context.Context) gorm.Interface[pgmodels.Comic] {
	return gorm.G[pgmodels.Comic](pgtx.From(ctx, r.deps.DB))
}

func (r *PGComicsRepository) GetByID(ctx context.Context, id uuid.UUID) (*comics.Comic, error) {
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

func (r *PGComicsRepository) GetBySourceSlug(ctx context.Context, key comics.SourceSlugKey) (*comics.Comic, error) {
	model, err := r.db(ctx).Where("source = ? AND slug = ?", key.Source, key.Slug).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := r.modelToDomain(model)

	return &ret, nil
}

func (r *PGComicsRepository) Create(ctx context.Context, opts comics.CreateComicOpts) (*comics.Comic, error) {
	now := time.Now()

	model := &pgmodels.Comic{
		ID:               uuid.New(),
		Source:           opts.Source,
		Slug:             opts.Slug,
		Title:            opts.Title,
		Status:           opts.Status,
		ComicType:        opts.Type,
		Genres:           pq.StringArray(opts.Genres),
		ChapterCount:     opts.ChapterCount,
		Author:           opts.Author,
		Artist:           opts.Artist,
		Description:      opts.Description,
		AltTitles:        pq.StringArray(opts.AltTitles),
		Rating:           opts.Rating,
		ReleaseYear:      opts.ReleaseYear,
		SourceURL:        opts.SourceURL,
		ExternalCoverURL: opts.ExternalCoverURL,
		LocalCoverPath:   opts.LocalCoverPath,
		CreatedAt:        now,
		UpdatedAt:        now,
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

func (r *PGComicsRepository) modelToDomain(model pgmodels.Comic) comics.Comic {
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
