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
	var rows []lightProviderRow

	err := pgtx.From(ctx, r.deps.DB).
		WithContext(ctx).
		Model(&pgmodels.OIDCProvider{}).
		Select("oidc_providers.id, oidc_providers.display_name, oidc_providers.created_at, " +
			"COUNT(DISTINCT federated_identities.user_id) AS user_count").
		Joins("LEFT JOIN federated_identities ON federated_identities.provider_id = oidc_providers.id").
		Group("oidc_providers.id").
		Order("oidc_providers.display_name, oidc_providers.id").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("pgtx.From(ctx, r.deps.DB).Scan: %w", err)
	}

	return utils.MapSlice(rows, r.lightRowToDomain), nil
}

//nolint:lll
func (r *PGOIDCProvidersRepository) GetUsers(ctx context.Context, id uuid.UUID) ([]oidcproviders.OIDCProviderUser, error) {
	var rows []providerUserRow

	err := pgtx.From(ctx, r.deps.DB).
		WithContext(ctx).
		Model(&pgmodels.FederatedIdentity{}).
		Select("users.id, users.name AS username, users.is_admin, federated_identities.created_at AS linked_at").
		Joins("JOIN users ON users.id = federated_identities.user_id").
		Where("federated_identities.provider_id = ?", id).
		Order("federated_identities.created_at, users.name").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("pgtx.From(ctx, r.deps.DB).Scan: %w", err)
	}

	return utils.MapSlice(rows, r.userRowToDomain), nil
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
		RoleClaim:       opts.RoleClaim,
		AdminValues:     opts.AdminValues,
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
		"role_claim":     opts.RoleClaim,
		"admin_values":   pq.StringArray(opts.AdminValues),
		"allowed_values": pq.StringArray(opts.AllowedValues),
		"auto_provision": opts.AutoProvision,
		"updated_at":     time.Now(),
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

type lightProviderRow struct {
	CreatedAt   time.Time
	DisplayName string
	ID          uuid.UUID
	UserCount   int64
}

type providerUserRow struct {
	LinkedAt time.Time
	Username string
	ID       uuid.UUID
	IsAdmin  bool
}

func (r *PGOIDCProvidersRepository) userRowToDomain(row providerUserRow) oidcproviders.OIDCProviderUser {
	return oidcproviders.OIDCProviderUser{
		ID:       row.ID,
		Username: row.Username,
		LinkedAt: row.LinkedAt,
		IsAdmin:  row.IsAdmin,
	}
}

func (r *PGOIDCProvidersRepository) lightRowToDomain(row lightProviderRow) oidcproviders.LightOIDCProvider {
	return oidcproviders.LightOIDCProvider{
		ID:          row.ID,
		DisplayName: row.DisplayName,
		CreatedAt:   row.CreatedAt,
		UserCount:   row.UserCount,
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
		RoleClaim:       model.RoleClaim,
		AdminValues:     model.AdminValues,
		AllowedValues:   model.AllowedValues,
		AutoProvision:   model.AutoProvision,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}
