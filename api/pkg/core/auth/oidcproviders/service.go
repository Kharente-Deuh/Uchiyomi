// SPDX-License-Identifier: AGPL-3.0-or-later

package oidcproviders

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrUnreachableIssuer = errors.New("issuer is unreachable")
	ErrIncompleteIssuer  = errors.New("issuer discovery document is incomplete")
)

type Cipher interface {
	Seal(plaintext []byte) ([]byte, error)
}

type CacheEvictor interface {
	Evict(id uuid.UUID)
}

type ServiceConfig struct {
	RedirectURI string
}

func (cfg *ServiceConfig) Validate() error {
	if cfg.RedirectURI == "" {
		return errors.New("redirectURI is required")
	}

	return nil
}

type ServiceDeps struct {
	Repository OIDCProvidersRepository
	Cipher     Cipher
	Discoverer Discoverer
	Cache      CacheEvictor
}

func (deps *ServiceDeps) Validate() error {
	if deps.Repository == nil {
		return errors.New("repository is required")
	}

	if deps.Cipher == nil {
		return errors.New("cipher is required")
	}

	if deps.Discoverer == nil {
		return errors.New("discoverer is required")
	}

	if deps.Cache == nil {
		return errors.New("cache is required")
	}

	return nil
}

type CreateOpts struct {
	DisplayName   string
	Slug          string
	IssuerURL     string
	ClientID      string
	ClientSecret  string
	UsernameClaim string

	Scopes []string

	RoleClaim *string

	AdminValues   []string
	AllowedValues []string
	AutoProvision bool
}

type UpdateOpts struct {
	DisplayName   string
	Slug          string
	IssuerURL     string
	ClientID      string
	UsernameClaim string

	Scopes []string

	RoleClaim *string

	AdminValues   []string
	AllowedValues []string
	AutoProvision bool
}

type ProbeResult struct {
	Issuer                    string
	AuthorizationEndpoint     string
	TokenEndpoint             string
	UserInfoEndpoint          string
	EndSessionEndpoint        string
	RedirectURI               string
	SupportsRPInitiatedLogout bool
}

type Service struct {
	deps ServiceDeps
	cfg  ServiceConfig
}

func NewService(cfg ServiceConfig, deps ServiceDeps) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	return &Service{cfg: cfg, deps: deps}, nil
}

func (s *Service) List(ctx context.Context) ([]LightOIDCProvider, error) {
	providers, err := s.deps.Repository.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("s.deps.Repository.GetAll: %w", err)
	}

	return providers, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*OIDCProviderDetails, error) {
	provider, err := s.deps.Repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("s.deps.Repository.GetByID: %w", err)
	}

	users, err := s.deps.Repository.GetUsers(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("s.deps.Repository.GetUsers: %w", err)
	}

	return &OIDCProviderDetails{Provider: *withoutSecret(provider), Users: users}, nil
}

func (s *Service) Create(ctx context.Context, opts CreateOpts) (*OIDCProvider, error) {
	if _, err := s.discover(ctx, opts.IssuerURL); err != nil {
		return nil, err
	}

	secretEnc, err := s.deps.Cipher.Seal([]byte(opts.ClientSecret))
	if err != nil {
		return nil, fmt.Errorf("s.deps.Cipher.Seal: %w", err)
	}

	provider, err := s.deps.Repository.Create(ctx, CreateOIDCProviderOpts{
		DisplayName:     opts.DisplayName,
		Slug:            opts.Slug,
		IssuerURL:       opts.IssuerURL,
		ClientID:        opts.ClientID,
		ClientSecretEnc: secretEnc,
		Scopes:          opts.Scopes,
		UsernameClaim:   opts.UsernameClaim,
		RoleClaim:       opts.RoleClaim,
		AdminValues:     opts.AdminValues,
		AllowedValues:   opts.AllowedValues,
		AutoProvision:   opts.AutoProvision,
	})
	if err != nil {
		return nil, fmt.Errorf("s.deps.Repository.Create: %w", err)
	}

	return withoutSecret(provider), nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, opts UpdateOpts) (*OIDCProvider, error) {
	if _, err := s.discover(ctx, opts.IssuerURL); err != nil {
		return nil, err
	}

	provider, err := s.deps.Repository.Update(ctx, id, UpdateOIDCProviderOpts{
		DisplayName:   opts.DisplayName,
		Slug:          opts.Slug,
		IssuerURL:     opts.IssuerURL,
		ClientID:      opts.ClientID,
		Scopes:        opts.Scopes,
		UsernameClaim: opts.UsernameClaim,
		RoleClaim:     opts.RoleClaim,
		AdminValues:   opts.AdminValues,
		AllowedValues: opts.AllowedValues,
		AutoProvision: opts.AutoProvision,
	})
	if err != nil {
		return nil, fmt.Errorf("s.deps.Repository.Update: %w", err)
	}

	s.deps.Cache.Evict(id)

	return withoutSecret(provider), nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.deps.Repository.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("s.deps.Repository.DeleteByID: %w", err)
	}

	s.deps.Cache.Evict(id)

	return nil
}

func (s *Service) Probe(ctx context.Context, issuerURL string) (*ProbeResult, error) {
	discovery, err := s.discover(ctx, issuerURL)
	if err != nil {
		return nil, err
	}

	return &ProbeResult{
		Issuer:                    discovery.Issuer,
		AuthorizationEndpoint:     discovery.AuthorizationEndpoint,
		TokenEndpoint:             discovery.TokenEndpoint,
		UserInfoEndpoint:          discovery.UserInfoEndpoint,
		EndSessionEndpoint:        discovery.EndSessionEndpoint,
		RedirectURI:               s.cfg.RedirectURI,
		SupportsRPInitiatedLogout: discovery.EndSessionEndpoint != "",
	}, nil
}

func (s *Service) discover(ctx context.Context, issuerURL string) (*Discovery, error) {
	discovery, err := s.deps.Discoverer.Discover(ctx, issuerURL)
	if err != nil {
		if errors.Is(err, ErrIncompleteDiscovery) {
			return nil, fmt.Errorf("%w: %w", ErrIncompleteIssuer, err)
		}

		return nil, fmt.Errorf("%w: %w", ErrUnreachableIssuer, err)
	}

	return discovery, nil
}

func withoutSecret(provider *OIDCProvider) *OIDCProvider {
	if provider == nil {
		return nil
	}

	provider.ClientSecretEnc = nil

	return provider
}
