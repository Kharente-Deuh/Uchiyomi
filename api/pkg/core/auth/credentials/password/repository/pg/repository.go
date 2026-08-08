// SPDX-License-Identifier: AGPL-3.0-or-later

package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/password"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgmodels"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
	"gorm.io/gorm"
)

var _ password.PasswordCredsRepository = (*PGPasswordCredsRepository)(nil)

type Deps struct {
	DB *gorm.DB
}

func (deps *Deps) Validate() error {
	if deps.DB == nil {
		return errors.New("db is required")
	}

	return nil
}

type PGPasswordCredsRepository struct {
	deps Deps
}

func New(deps Deps) (*PGPasswordCredsRepository, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	r := &PGPasswordCredsRepository{deps: deps}

	return r, nil
}

func (r *PGPasswordCredsRepository) db(ctx context.Context) gorm.Interface[pgmodels.PasswordCreds] {
	return gorm.G[pgmodels.PasswordCreds](pgtx.From(ctx, r.deps.DB))
}

//nolint:lll
func (r *PGPasswordCredsRepository) Create(ctx context.Context, opts password.UpsertPasswordCredsOpts) (*password.PasswordCreds, error) {
	model := &pgmodels.PasswordCreds{
		UserID:    opts.UserID,
		Hash:      opts.Hash,
		UpdatedAt: time.Now(),
	}

	err := r.db(ctx).Create(ctx, model)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, domain.ErrAlreadyExists
		}

		return nil, fmt.Errorf("r.db(ctx).Create: %w", err)
	}

	pwdCred := r.modelToDomain(*model)

	return &pwdCred, nil
}

//nolint:lll
func (r *PGPasswordCredsRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*password.PasswordCreds, error) {
	model, err := r.db(ctx).Where("user_id = ?", userID).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	pwdCreds := r.modelToDomain(model)

	return &pwdCreds, nil
}

func (r *PGPasswordCredsRepository) UpdateByUserID(ctx context.Context, opts password.UpsertPasswordCredsOpts) error {
	affectedNb, err := r.db(ctx).Where("user_id = ?", opts.UserID).Update(ctx, "hash", opts.Hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}

		return fmt.Errorf("r.db(ctx).Where(%s).Update: %w", opts.UserID, err)
	}

	if affectedNb == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *PGPasswordCredsRepository) modelToDomain(model pgmodels.PasswordCreds) password.PasswordCreds {
	return password.PasswordCreds{
		UserID:    model.UserID,
		Hash:      model.Hash,
		UpdatedAt: model.UpdatedAt,
	}
}
