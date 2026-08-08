// SPDX-License-Identifier: AGPL-3.0-or-later

package password

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PasswordCreds struct {
	UpdatedAt time.Time
	Hash      string
	UserID    uuid.UUID
}

type PasswordCredsRepository interface {
	Create(context.Context, UpsertPasswordCredsOpts) (*PasswordCreds, error)
	GetByUserID(context.Context, uuid.UUID) (*PasswordCreds, error)
	UpdateByUserID(context.Context, UpsertPasswordCredsOpts) error
}

type UpsertPasswordCredsOpts struct {
	Hash   string
	UserID uuid.UUID
}
