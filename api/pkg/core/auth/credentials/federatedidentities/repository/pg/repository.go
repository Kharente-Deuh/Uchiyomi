// SPDX-License-Identifier: AGPL-3.0-or-later

package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/federatedidentities"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgmodels"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
	"gorm.io/gorm"
)

var _ federatedidentities.FederatedIdentitiesRepository = (*PGFederatedIdentitiesRepository)(nil)

type Deps struct {
	DB *gorm.DB
}

func (deps *Deps) Validate() error {
	if deps.DB == nil {
		return errors.New("db is required")
	}

	return nil
}

type PGFederatedIdentitiesRepository struct {
	deps Deps
}

func New(deps Deps) (*PGFederatedIdentitiesRepository, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	r := &PGFederatedIdentitiesRepository{deps: deps}

	return r, nil
}

func (r *PGFederatedIdentitiesRepository) db(ctx context.Context) gorm.Interface[pgmodels.FederatedIdentity] {
	return gorm.G[pgmodels.FederatedIdentity](pgtx.From(ctx, r.deps.DB))
}

//nolint:lll
func (r *PGFederatedIdentitiesRepository) Create(ctx context.Context, opts federatedidentities.CreateFederatedIdentityOpts) (*federatedidentities.FederatedIdentity, error) {
	now := time.Now()
	model := &pgmodels.FederatedIdentity{
		ID:              uuid.New(),
		UserID:          opts.UserID,
		ProviderID:      opts.ProviderID,
		Subject:         opts.Subject,
		Claims:          opts.Claims,
		RefreshTokenEnc: opts.RefreshTokenEnc,
		LastValidatedAt: opts.LastValidatedAt,
		LastLoginAt:     now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	err := r.db(ctx).Create(ctx, model)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, domain.ErrAlreadyExists
		}

		return nil, fmt.Errorf("r.db(ctx).Create: %w", err)
	}

	fi := r.modelToDomain(*model)

	return &fi, nil
}

//nolint:lll
func (r *PGFederatedIdentitiesRepository) Get(ctx context.Context, opts federatedidentities.GetFederatedIdentityOpts) (*federatedidentities.FederatedIdentity, error) {
	model, err := r.db(ctx).Where(
		"provider_id = ? AND subject = ?",
		opts.ProviderID, opts.Subject,
	).First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	fi := r.modelToDomain(model)

	return &fi, nil
}

//nolint:lll
func (r *PGFederatedIdentitiesRepository) Update(ctx context.Context, opts federatedidentities.UpdateFederatedIdentityOpts) error {
	updates := pgmodels.FederatedIdentity{
		Claims:          opts.Claims,
		LastLoginAt:     opts.LastLoginAt,
		LastValidatedAt: opts.LastValidatedAt,
	}

	if opts.SetRefreshToken {
		updates.RefreshTokenEnc = opts.RefreshTokenEnc
	}

	if opts.ClearRefreshToken {
		updates.RefreshTokenEnc = nil
	}

	affectedNb, err := r.db(ctx).Where("id = ?", opts.ID).Updates(ctx, updates)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}

		return fmt.Errorf("r.db(ctx).Where(%s).Update: %w", opts.ID, err)
	}

	if affectedNb == 0 {
		return domain.ErrNotFound
	}

	return nil
}

//nolint:lll
func (r *PGFederatedIdentitiesRepository) ListDueForRevalidation(
	ctx context.Context,
	before time.Time,
) ([]federatedidentities.FederatedIdentity, error) {
	now := time.Now()

	var models []pgmodels.FederatedIdentity

	err := pgtx.From(ctx, r.deps.DB).
		Model(&pgmodels.FederatedIdentity{}).
		Distinct().
		Joins(`INNER JOIN sessions ON sessions.user_id = federated_identities.user_id
			AND sessions.provider_id = federated_identities.provider_id`).
		Where("federated_identities.refresh_token_enc IS NOT NULL").
		Where("federated_identities.last_validated_at < ?", before).
		Where("sessions.expires_at > ?", now).
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("db query: %w", err)
	}

	out := make([]federatedidentities.FederatedIdentity, len(models))
	for i, model := range models {
		out[i] = r.modelToDomain(model)
	}

	return out, nil
}

//nolint:lll
func (r *PGFederatedIdentitiesRepository) modelToDomain(model pgmodels.FederatedIdentity) federatedidentities.FederatedIdentity {
	return federatedidentities.FederatedIdentity{
		ID:              model.ID,
		UserID:          model.UserID,
		ProviderID:      model.ProviderID,
		Subject:         model.Subject,
		Claims:          model.Claims,
		RefreshTokenEnc: model.RefreshTokenEnc,
		LastValidatedAt: model.LastValidatedAt,
		LastLoginAt:     model.LastLoginAt,
		CreatedAt:       model.CreatedAt,
	}
}
