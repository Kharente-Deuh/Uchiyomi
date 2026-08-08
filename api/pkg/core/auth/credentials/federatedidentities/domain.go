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
}

type FederatedIdentity struct {
	CreatedAt   time.Time
	LastLoginAt time.Time
	Claims      map[string]any
	Subject     string
	ID          uuid.UUID
	UserID      uuid.UUID
	ProviderID  uuid.UUID
}

type CreateFederatedIdentityOpts struct {
	Claims     map[string]any
	Subject    string
	UserID     uuid.UUID
	ProviderID uuid.UUID
}

type GetFederatedIdentityOpts struct {
	Subject    string
	ProviderID uuid.UUID
}

type UpdateFederatedIdentityOpts struct {
	LastLoginAt time.Time
	Claims      map[string]any
	ID          uuid.UUID
}
