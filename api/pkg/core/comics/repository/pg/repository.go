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
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *PGComicsRepository) GetByID(ctx context.Context, opts comics.GetByIDOpts) (*comics.Comic, error) {
	model, err := r.db(ctx).
		Joins(clause.JoinTarget{Association: pgmodels.ComicLibraryEntries}, nil).
		Where(`comics.id = ? AND "LibraryEntries"."user_id" = ?`, opts.ID, opts.UserID).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := model.Domain()

	return &ret, nil
}

//nolint:lll
func (r *PGComicsRepository) GetBySourceSlug(ctx context.Context, opts comics.GetBySourceSlugOpts) (*comics.Comic, error) {
	model, err := r.db(ctx).
		Joins(clause.JoinTarget{Association: pgmodels.ComicLibraryEntries}, nil).
		Where(`comics.slug = ? AND comics.source = ? AND "LibraryEntries"."user_id" = ?`, opts.Slug, opts.Source, opts.UserID).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := model.Domain()

	return &ret, nil
}

func (r *PGComicsRepository) Create(ctx context.Context, opts comics.CreateComicOpts) (*comics.Comic, error) {
	now := time.Now()

	model := &pgmodels.Comic{
		ID:           uuid.New(),
		Source:       opts.Source,
		Slug:         opts.Slug,
		Title:        opts.Title,
		Status:       pgmodels.ComicStatusFromDomain(opts.Status),
		ComicType:    pgmodels.ComicTypeFromDomain(opts.Type),
		Genres:       pq.StringArray(opts.Genres),
		ChapterCount: opts.ChapterCount,
		Author:       opts.Author,
		Artist:       opts.Artist,
		Description:  opts.Description,
		AltTitles:    pq.StringArray(opts.AltTitles),
		CreatedAt:    now,
		UpdatedAt:    now,
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

func (r *PGComicsRepository) GetMany(ctx context.Context, opts comics.GetManyOpts) ([]comics.Comic, error) {
	q := r.db(ctx).Joins(clause.JoinTarget{Association: pgmodels.ComicLibraryEntries}, nil).Order("comics.created_at DESC")
	if opts.UserID != nil {
		q = q.Where(`"LibraryEntries"."user_id" = ?`, opts.UserID)
	}

	if opts.Source != nil {
		q = q.Where("comics.source = ?", opts.Source)
	}

	if opts.Type != nil {
		q = q.Where("comics.type = ?", opts.Type)
	}

	if opts.Status != nil {
		q = q.Where("comics.status = ?", opts.Status)
	}

	models, err := q.Offset(opts.Offset).Limit(opts.Limit).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := make([]comics.Comic, len(models))
	for i, model := range models {
		ret[i] = model.Domain()
	}

	return ret, nil
}

// nolint:lll
func (r *PGComicsRepository) GetBySlugsAndSource(
	ctx context.Context,
	source sources.SourceName,
	slugs []string,
) ([]comics.Comic, error) {
	models, err := r.db(ctx).Where("source = ? AND slug IN (?)", source, slugs).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := make([]comics.Comic, len(models))
	for i, model := range models {
		ret[i] = model.Domain()
	}

	return ret, nil
}

func (r *PGComicsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	affected, err := r.db(ctx).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	if affected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
