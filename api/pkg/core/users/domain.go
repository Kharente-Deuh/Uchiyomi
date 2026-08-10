// SPDX-License-Identifier: AGPL-3.0-or-later

package users

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	ID        uuid.UUID
	IsAdmin   bool
}

type UsersRepository interface {
	CountAdmins(context.Context) (int, error)
	GetByID(context.Context, uuid.UUID) (*User, error)
	GetByUsername(context.Context, string) (*User, error)
	Create(context.Context, CreateUserOpts) (*User, error)
	Update(context.Context, UpdateUserOpts) (*User, error)
}

type CreateUserOpts struct {
	Name    string
	IsAdmin bool
}

type UpdateUserOpts struct {
	ID      uuid.UUID
	IsAdmin bool
}
