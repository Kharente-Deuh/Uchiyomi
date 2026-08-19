// SPDX-License-Identifier: AGPL-3.0-or-later

package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readersettings"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgmodels"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ readersettings.Repository = (*PGReaderSettingsRepository)(nil)

type Deps struct {
	DB *gorm.DB
}

func (deps *Deps) Validate() error {
	if deps.DB == nil {
		return errors.New("db is required")
	}

	return nil
}

type PGReaderSettingsRepository struct {
	deps Deps
}

func New(deps Deps) (*PGReaderSettingsRepository, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	r := &PGReaderSettingsRepository{deps: deps}

	return r, nil
}

func (r *PGReaderSettingsRepository) db(ctx context.Context) gorm.Interface[pgmodels.ReaderSettings] {
	return gorm.G[pgmodels.ReaderSettings](pgtx.From(ctx, r.deps.DB))
}

func (r *PGReaderSettingsRepository) ListByUser(
	ctx context.Context, userID uuid.UUID,
) ([]readersettings.Profile, error) {
	models, err := r.db(ctx).Where("user_id = ?", userID).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := make([]readersettings.Profile, 0, len(models))
	for i := range models {
		ret = append(ret, models[i].Domain())
	}

	return ret, nil
}

func (r *PGReaderSettingsRepository) Upsert(
	ctx context.Context, opts readersettings.UpsertOpts,
) (readersettings.Profile, error) {
	model := &pgmodels.ReaderSettings{
		ID:          uuid.New(),
		UserID:      opts.UserID,
		ComicType:   pgmodels.ComicTypeFromDomain(opts.Type),
		ReadingMode: string(opts.ReadingMode),
		PageScale:   string(opts.PageScale),
		DoublePage:  opts.DoublePage,
	}

	err := pgtx.From(ctx, r.deps.DB).WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "comic_type"}},
		DoUpdates: clause.AssignmentColumns([]string{"reading_mode", "page_scale", "double_page", "updated_at"}),
	}).Create(model).Error
	if err != nil {
		return readersettings.Profile{}, fmt.Errorf("pgtx.From(ctx, r.deps.DB).Create: %w", err)
	}

	return model.Domain(), nil
}
