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
	GetBySlug(context.Context, string) (*OIDCProvider, error)
	Create(context.Context, CreateOIDCProviderOpts) (*OIDCProvider, error)
	Update(context.Context, uuid.UUID, UpdateOIDCProviderOpts) (*OIDCProvider, error)
	DeleteByID(context.Context, uuid.UUID) error
	GetAll(context.Context) ([]LightOIDCProvider, error)
	GetUsers(context.Context, uuid.UUID) ([]OIDCProviderUser, error)
}

type OIDCProvider struct {
	UpdatedAt       time.Time
	CreatedAt       time.Time
	RoleClaim       *string
	ClientID        string
	UsernameClaim   string
	IssuerURL       string
	DisplayName     string
	Slug            string
	Scopes          []string
	ClientSecretEnc []byte
	AdminValues     []string
	AllowedValues   []string
	ID              uuid.UUID
	AutoProvision   bool
}

type UpdateOIDCProviderOpts struct {
	DisplayName string
	Slug        string

	IssuerURL string
	ClientID  string
	Scopes    []string

	UsernameClaim string
	RoleClaim     *string

	AdminValues   []string
	AllowedValues []string
	AutoProvision bool
}

type CreateOIDCProviderOpts struct {
	DisplayName string
	Slug        string

	IssuerURL       string
	ClientID        string
	ClientSecretEnc []byte
	Scopes          []string

	UsernameClaim string
	RoleClaim     *string

	AdminValues   []string
	AllowedValues []string
	AutoProvision bool
}

type OIDCProviderUser struct {
	LinkedAt time.Time
	Username string
	ID       uuid.UUID
	IsAdmin  bool
}

type OIDCProviderDetails struct {
	Users    []OIDCProviderUser
	Provider OIDCProvider
}

type LightOIDCProvider struct {
	CreatedAt   time.Time
	DisplayName string
	Slug        string
	ID          uuid.UUID
	UserCount   int64
}

var ErrSlugTaken = errors.New("oidc provider slug already exists")

var ErrIncompleteDiscovery = errors.New("discovery document is incomplete")

var ErrInvalidGrant = errors.New("oidc: refresh token is invalid or revoked")

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

type Client interface {
	AuthCodeURL(ctx context.Context, provider OIDCProvider, params AuthCodeParams) (string, error)
	Exchange(ctx context.Context, provider OIDCProvider, code, verifier, nonce, redirectURI string) (*TokenSet, error)
	Refresh(ctx context.Context, provider OIDCProvider, refreshToken string) (*TokenSet, error)
	EndSessionURL(ctx context.Context, provider OIDCProvider, postLogoutRedirectURI string) (string, bool, error)
	VerifyLogoutToken(ctx context.Context, provider OIDCProvider, raw string) (*LogoutToken, error)
}

type AuthCodeParams struct {
	RedirectURI string
	State       string
	Nonce       string
	Verifier    string
}

type TokenSet struct {
	Claims       map[string]any
	RefreshToken string
	Subject      string
	SID          string
}

type LogoutToken struct {
	Subject string
	SID     string
}

var ErrLogoutTokenInvalid = errors.New("oidc: logout token is invalid")
