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
	RefreshTokenEnc []byte
	Claims          map[string]any
	Subject         string
	ID              uuid.UUID
	UserID          uuid.UUID
	ProviderID      uuid.UUID
}

type CreateFederatedIdentityOpts struct {
	LastValidatedAt time.Time
	RefreshTokenEnc []byte
	Claims          map[string]any
	Subject         string
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
	RefreshTokenEnc   []byte
	SetRefreshToken   bool
	ClearRefreshToken bool
	Claims            map[string]any
	ID                uuid.UUID
}
