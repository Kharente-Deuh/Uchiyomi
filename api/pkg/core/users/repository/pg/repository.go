// SPDX-License-Identifier: AGPL-3.0-or-later

package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgmodels"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
	"gorm.io/gorm"
)

var _ users.UsersRepository = (*PGUsersRepository)(nil)

type Deps struct {
	DB *gorm.DB
}

func (deps *Deps) Validate() error {
	if deps.DB == nil {
		return errors.New("db is required")
	}

	return nil
}

type PGUsersRepository struct {
	deps Deps
}

func New(deps Deps) (*PGUsersRepository, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	r := &PGUsersRepository{deps: deps}

	return r, nil
}

func (r *PGUsersRepository) db(ctx context.Context) gorm.Interface[pgmodels.User] {
	return gorm.G[pgmodels.User](pgtx.From(ctx, r.deps.DB))
}

func (r *PGUsersRepository) CountAdmins(ctx context.Context) (int, error) {
	count, err := r.db(ctx).Where("is_admin = ?", true).Count(ctx, "*")
	if err != nil {
		return 0, fmt.Errorf("r.db(ctx).Count: %w", err)
	}

	return int(count), nil
}

func (r *PGUsersRepository) GetByID(ctx context.Context, id uuid.UUID) (*users.User, error) {
	user, err := r.db(ctx).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := r.modelToDomain(user)

	return &ret, nil
}

func (r *PGUsersRepository) Create(ctx context.Context, opts users.CreateUserOpts) (*users.User, error) {
	model := &pgmodels.User{
		ID:        uuid.New(),
		Name:      opts.Name,
		IsAdmin:   opts.IsAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := r.db(ctx).Create(ctx, model)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, domain.ErrAlreadyExists
		}

		return nil, fmt.Errorf("r.db(ctx).Create: %w", err)
	}

	user := r.modelToDomain(*model)

	return &user, nil
}

func (r *PGUsersRepository) GetByUsername(ctx context.Context, name string) (*users.User, error) {
	user, err := r.db(ctx).Where("name = ?", name).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := r.modelToDomain(user)

	return &ret, nil
}

func (r *PGUsersRepository) Update(ctx context.Context, opts users.UpdateUserOpts) (*users.User, error) {
	affected, err := r.db(ctx).Where("id = ?", opts.ID).Update(ctx, "is_admin", opts.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	if affected == 0 {
		return nil, domain.ErrNotFound
	}

	user, err := r.db(ctx).Where("id = ?", opts.ID).First(ctx)
	if err != nil {
		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := r.modelToDomain(user)

	return &ret, nil
}

func (r *PGUsersRepository) modelToDomain(model pgmodels.User) users.User {
	return users.User{
		ID:        model.ID,
		Name:      model.Name,
		IsAdmin:   model.IsAdmin,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
