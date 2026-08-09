// SPDX-License-Identifier: AGPL-3.0-or-later

package oidcproviders

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type OIDCProvidersRepository interface {
	GetByID(context.Context, uuid.UUID) (*OIDCProvider, error)
	GetByIssuerURL(context.Context, string) (*OIDCProvider, error)
	Create(context.Context, CreateOIDCProviderOpts) (*OIDCProvider, error)
	Update(context.Context, uuid.UUID, UpdateOIDCProviderOpts) (*OIDCProvider, error)
	DeleteByID(context.Context, uuid.UUID) error
	GetAll(context.Context) ([]LightOIDCProvider, error)
}

type OIDCProvider struct {
	UpdatedAt       time.Time
	CreatedAt       time.Time
	AdminClaim      *string
	AllowedClaim    *string
	ClientID        string
	UsernameClaim   string
	IssuerURL       string
	DisplayName     string
	Scopes          []string
	ClientSecretEnc []byte
	AdminValues     []string
	AllowedValues   []string
	ID              uuid.UUID
	AutoProvision   bool
}

type UpdateOIDCProviderOpts struct {
	DisplayName string

	IssuerURL       string
	ClientID        string
	ClientSecretEnc []byte
	Scopes          []string

	UsernameClaim string

	AdminClaim  *string
	AdminValues []string

	AllowedClaim  *string
	AllowedValues []string
	AutoProvision bool
}

type CreateOIDCProviderOpts struct {
	DisplayName string

	IssuerURL       string
	ClientID        string
	ClientSecretEnc []byte
	Scopes          []string

	UsernameClaim string

	AdminClaim  *string
	AdminValues []string

	AllowedClaim  *string
	AllowedValues []string
	AutoProvision bool
}

type LightOIDCProvider struct {
	DisplayName string
	ID          uuid.UUID
}

var ErrIncompleteDiscovery = errors.New("discovery document is incomplete")

type Discoverer interface {
	Discover(ctx context.Context, issuerURL string) (*Discovery, error)
}

type Discovery struct {
	Issuer                string
	AuthorizationEndpoint string
	TokenEndpoint         string
	UserInfoEndpoint      string
	EndSessionEndpoint    string
}
