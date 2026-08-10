// SPDX-License-Identifier: AGPL-3.0-or-later

package federatedidentities

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type FederatedIdentitiesRepository interface {
	Create(context.Context, CreateFederatedIdentityOpts) (*FederatedIdentity, error)
	Get(context.Context, GetFederatedIdentityOpts) (*FederatedIdentity, error)
	Update(context.Context, UpdateFederatedIdentityOpts) error
	ListDueForRevalidation(context.Context, time.Time) ([]FederatedIdentity, error)
}

type FederatedIdentity struct {
	LastValidatedAt time.Time
	LastLoginAt     time.Time
	CreatedAt       time.Time
	Claims          map[string]any
	Subject         string
	RefreshTokenEnc []byte
	ID              uuid.UUID
	UserID          uuid.UUID
	ProviderID      uuid.UUID
}

type CreateFederatedIdentityOpts struct {
	LastValidatedAt time.Time
	Claims          map[string]any
	Subject         string
	RefreshTokenEnc []byte
	UserID          uuid.UUID
	ProviderID      uuid.UUID
}

type GetFederatedIdentityOpts struct {
	Subject    string
	ProviderID uuid.UUID
}

type UpdateFederatedIdentityOpts struct {
	LastValidatedAt   time.Time
	LastLoginAt       time.Time
	Claims            map[string]any
	RefreshTokenEnc   []byte
	ID                uuid.UUID
	SetRefreshToken   bool
	ClearRefreshToken bool
}
