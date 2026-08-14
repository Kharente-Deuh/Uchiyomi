// SPDX-License-Identifier: AGPL-3.0-or-later

package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgmodels"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
	"gorm.io/gorm"
)

var _ chapters.ChaptersRepository = (*PGChaptersRepository)(nil)

type Deps struct {
	DB *gorm.DB
}

func (deps *Deps) Validate() error {
	if deps.DB == nil {
		return errors.New("db is required")
	}

	return nil
}

type PGChaptersRepository struct {
	deps Deps
}

func New(deps Deps) (*PGChaptersRepository, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	r := &PGChaptersRepository{deps: deps}

	return r, nil
}

func (r *PGChaptersRepository) db(ctx context.Context) gorm.Interface[pgmodels.Chapter] {
	return gorm.G[pgmodels.Chapter](pgtx.From(ctx, r.deps.DB))
}

func (r *PGChaptersRepository) Create(ctx context.Context, opts chapters.CreateOpts) (*chapters.Chapter, error) {
	model := &pgmodels.Chapter{
		ID:                uuid.New(),
		ComicID:           opts.ComicID,
		SourceChapterSlug: opts.SourceChapterSlug,
		Number:            opts.Number,
		Title:             opts.Title,
		PagesNb:           opts.PagesNb,
		PublishedAt:       opts.PublishedAt,
		EarlyAccessUntil:  opts.EarlyAccessUntil,
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

func (r *PGChaptersRepository) ListByComicID(ctx context.Context, comicID uuid.UUID) ([]chapters.Chapter, error) {
	models, err := r.db(ctx).Where("comic_id = ?", comicID).Order("number ASC").Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := make([]chapters.Chapter, 0, len(models))
	for _, model := range models {
		ret = append(ret, model.Domain())
	}

	return ret, nil
}
