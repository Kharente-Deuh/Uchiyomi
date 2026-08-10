// SPDX-License-Identifier: AGPL-3.0-or-later

package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgmodels"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ sessions.SessionsRepository = (*PGSessionsRepository)(nil)

type Deps struct {
	DB *gorm.DB
}

func (deps *Deps) Validate() error {
	if deps.DB == nil {
		return errors.New("db is required")
	}

	return nil
}

type PGSessionsRepository struct {
	deps Deps
}

func New(deps Deps) (*PGSessionsRepository, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	return &PGSessionsRepository{deps: deps}, nil
}

func (r *PGSessionsRepository) db(ctx context.Context) gorm.Interface[pgmodels.Session] {
	return gorm.G[pgmodels.Session](pgtx.From(ctx, r.deps.DB))
}

func (r *PGSessionsRepository) Insert(ctx context.Context, opts sessions.InsertSessionOpts) (*sessions.Session, error) {
	model := &pgmodels.Session{
		ID:          uuid.New(),
		UserID:      opts.UserID,
		TokenHash:   opts.TokenHash,
		AuthMethod:  string(opts.AuthMethod),
		ExpiresAt:   opts.ExpiresAt,
		CreatedAt:   time.Now(),
		ProviderID:  opts.ProviderID,
		ProviderSID: opts.ProviderSID,
	}

	if err := r.db(ctx).Create(ctx, model); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, domain.ErrAlreadyExists
		}

		return nil, fmt.Errorf("r.db(ctx).Create: %w", err)
	}

	session := r.modelToDomain(*model)

	return &session, nil
}

//nolint:lll
func (r *PGSessionsRepository) GetByTokenHash(ctx context.Context, hash []byte) (*sessions.Session, *users.User, error) {
	model, err := r.db(ctx).
		Joins(clause.JoinTarget{Association: "User"}, nil).
		Where("sessions.token_hash = ?", hash).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, domain.ErrNotFound
		}

		return nil, nil, fmt.Errorf("r.db(ctx).First: %w", err)
	}

	session := r.modelToDomain(model)
	user := users.User{
		ID:        model.User.ID,
		Name:      model.User.Name,
		IsAdmin:   model.User.IsAdmin,
		CreatedAt: model.User.CreatedAt,
		UpdatedAt: model.User.UpdatedAt,
	}

	return &session, &user, nil
}

func (r *PGSessionsRepository) UpdateExpiry(ctx context.Context, id uuid.UUID, expiresAt time.Time) error {
	if _, err := r.db(ctx).Where("id = ?", id).Update(ctx, "expires_at", expiresAt); err != nil {
		return fmt.Errorf("r.db(ctx).Update: %w", err)
	}

	return nil
}

func (r *PGSessionsRepository) DeleteByTokenHash(ctx context.Context, hash []byte) error {
	if _, err := r.db(ctx).Where("token_hash = ?", hash).Delete(ctx); err != nil {
		return fmt.Errorf("r.db(ctx).Delete: %w", err)
	}

	return nil
}

func (r *PGSessionsRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	if _, err := r.db(ctx).Where("user_id = ?", userID).Delete(ctx); err != nil {
		return fmt.Errorf("r.db(ctx).Delete: %w", err)
	}

	return nil
}

func (r *PGSessionsRepository) DeleteByUserAndProvider(
	ctx context.Context,
	userID, providerID uuid.UUID,
) error {
	if _, err := r.db(ctx).Where("user_id = ? AND provider_id = ?", userID, providerID).Delete(ctx); err != nil {
		return fmt.Errorf("r.db(ctx).Delete: %w", err)
	}

	return nil
}

func (r *PGSessionsRepository) DeleteByProviderAndSID(ctx context.Context, providerID uuid.UUID, sid string) error {
	if _, err := r.db(ctx).Where("provider_id = ? AND provider_sid = ?", providerID, sid).Delete(ctx); err != nil {
		return fmt.Errorf("r.db(ctx).Delete: %w", err)
	}

	return nil
}

func (r *PGSessionsRepository) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	deleted, err := r.db(ctx).Where("expires_at <= ?", now).Delete(ctx)
	if err != nil {
		return 0, fmt.Errorf("r.db(ctx).Delete: %w", err)
	}

	return int64(deleted), nil
}

func (r *PGSessionsRepository) modelToDomain(model pgmodels.Session) sessions.Session {
	return sessions.Session{
		ID:          model.ID,
		UserID:      model.UserID,
		AuthMethod:  sessions.AuthMethod(model.AuthMethod),
		CreatedAt:   model.CreatedAt,
		ExpiresAt:   model.ExpiresAt,
		ProviderID:  model.ProviderID,
		ProviderSID: model.ProviderSID,
	}
}
