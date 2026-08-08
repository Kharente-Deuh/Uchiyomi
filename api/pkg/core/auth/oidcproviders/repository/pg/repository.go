// SPDX-License-Identifier: AGPL-3.0-or-later

package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgmodels"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ oidcproviders.OIDCProvidersRepository = (*PGOIDCProvidersRepository)(nil)

type Deps struct {
	DB *gorm.DB
}

func (deps *Deps) Validate() error {
	if deps.DB == nil {
		return errors.New("db is required")
	}

	return nil
}

type PGOIDCProvidersRepository struct {
	deps Deps
}

func New(deps Deps) (*PGOIDCProvidersRepository, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	r := &PGOIDCProvidersRepository{deps: deps}

	return r, nil
}

func (r *PGOIDCProvidersRepository) db(ctx context.Context) gorm.Interface[pgmodels.OIDCProvider] {
	return gorm.G[pgmodels.OIDCProvider](pgtx.From(ctx, r.deps.DB))
}

func (r *PGOIDCProvidersRepository) GetByID(ctx context.Context, id uuid.UUID) (*oidcproviders.OIDCProvider, error) {
	model, err := r.db(ctx).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	p := r.modelToDomain(model)

	return &p, nil
}

//nolint:lll
func (r *PGOIDCProvidersRepository) GetByIssuerURL(ctx context.Context, issuerURL string) (*oidcproviders.OIDCProvider, error) {
	model, err := r.db(ctx).Where("issuer_url = ?", issuerURL).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	p := r.modelToDomain(model)

	return &p, nil
}

func (r *PGOIDCProvidersRepository) GetAll(ctx context.Context) ([]oidcproviders.LightOIDCProvider, error) {
	models, err := r.db(ctx).Order("display_name").Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("r.db(ctx).Find: %w", err)
	}

	return utils.MapSlice(models, r.lightModelToDomain), nil
}

//nolint:lll
func (r *PGOIDCProvidersRepository) Create(ctx context.Context, opts oidcproviders.CreateOIDCProviderOpts) (*oidcproviders.OIDCProvider, error) {
	now := time.Now()
	model := &pgmodels.OIDCProvider{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,

		DisplayName:     opts.DisplayName,
		IssuerURL:       opts.IssuerURL,
		ClientID:        opts.ClientID,
		ClientSecretEnc: opts.ClientSecretEnc,
		Scopes:          opts.Scopes,
		UsernameClaim:   opts.UsernameClaim,
		AdminClaim:      opts.AdminClaim,
		AdminValues:     opts.AdminValues,
		AllowedClaim:    opts.AllowedClaim,
		AllowedValues:   opts.AllowedValues,
		AutoProvision:   opts.AutoProvision,
	}

	err := r.db(ctx).Create(ctx, model)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, domain.ErrAlreadyExists
		}

		return nil, fmt.Errorf("r.db(ctx).Create: %w", err)
	}

	p := r.modelToDomain(*model)

	return &p, nil
}

//nolint:lll
func (r *PGOIDCProvidersRepository) Update(ctx context.Context, id uuid.UUID, opts oidcproviders.UpdateOIDCProviderOpts) (*oidcproviders.OIDCProvider, error) {
	values := map[string]any{
		"display_name":   opts.DisplayName,
		"issuer_url":     opts.IssuerURL,
		"client_id":      opts.ClientID,
		"scopes":         pq.StringArray(opts.Scopes),
		"username_claim": opts.UsernameClaim,
		"admin_claim":    opts.AdminClaim,
		"admin_values":   pq.StringArray(opts.AdminValues),
		"allowed_claim":  opts.AllowedClaim,
		"allowed_values": pq.StringArray(opts.AllowedValues),
		"auto_provision": opts.AutoProvision,
		"updated_at":     time.Now(),
	}

	if opts.ClientSecretEnc != nil {
		values["client_secret_enc"] = opts.ClientSecretEnc
	}

	rows, err := r.db(ctx).Where("id = ?", id).Set(clause.Assignments(values)).Update(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, domain.ErrAlreadyExists
		}

		return nil, fmt.Errorf("r.db(ctx).Updates: %w", err)
	}

	if rows == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *PGOIDCProvidersRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	rows, err := r.db(ctx).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("r.db(ctx).Delete: %w", err)
	}

	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *PGOIDCProvidersRepository) lightModelToDomain(model pgmodels.OIDCProvider) oidcproviders.LightOIDCProvider {
	return oidcproviders.LightOIDCProvider{
		ID:          model.ID,
		DisplayName: model.DisplayName,
	}
}

func (r *PGOIDCProvidersRepository) modelToDomain(model pgmodels.OIDCProvider) oidcproviders.OIDCProvider {
	return oidcproviders.OIDCProvider{
		ID:              model.ID,
		DisplayName:     model.DisplayName,
		IssuerURL:       model.IssuerURL,
		ClientID:        model.ClientID,
		ClientSecretEnc: model.ClientSecretEnc,
		Scopes:          model.Scopes,
		UsernameClaim:   model.UsernameClaim,
		AdminClaim:      model.AdminClaim,
		AdminValues:     model.AdminValues,
		AllowedClaim:    model.AllowedClaim,
		AllowedValues:   model.AllowedValues,
		AutoProvision:   model.AutoProvision,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}
