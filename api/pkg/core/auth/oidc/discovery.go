// SPDX-License-Identifier: AGPL-3.0-or-later

package oidc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
)

var _ oidcproviders.Discoverer = (*Discoverer)(nil)

var (
	ErrDiscoveryFailed     = errors.New("issuer discovery failed")
	ErrDiscoveryIncomplete = errors.New("discovery document is incomplete")
)

const defaultTimeout = 10 * time.Second

type Config struct {
	Timeout time.Duration
}

func (cfg *Config) Validate() error {
	if cfg.Timeout < 0 {
		return errors.New("timeout must not be negative")
	}

	return nil
}

type Deps struct {
	HTTPClient *http.Client
}

func (deps *Deps) Validate() error {
	if deps.HTTPClient == nil {
		return errors.New("httpClient is required")
	}

	return nil
}

type Discoverer struct {
	deps Deps
	cfg  Config
}

func New(cfg Config, deps Deps) (*Discoverer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}

	return &Discoverer{cfg: cfg, deps: deps}, nil
}

func (d *Discoverer) Discover(ctx context.Context, issuerURL string) (*oidcproviders.Discovery, error) {
	ctx, cancel := context.WithTimeout(gooidc.ClientContext(ctx, d.deps.HTTPClient), d.cfg.Timeout)
	defer cancel()

	provider, err := gooidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDiscoveryFailed, err)
	}

	var doc struct {
		AuthorizationEndpoint string `json:"authorization_endpoint"`
		TokenEndpoint         string `json:"token_endpoint"`
		UserInfoEndpoint      string `json:"userinfo_endpoint"`
		EndSessionEndpoint    string `json:"end_session_endpoint"`
	}

	if err := provider.Claims(&doc); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDiscoveryFailed, err)
	}

	if doc.AuthorizationEndpoint == "" || doc.TokenEndpoint == "" {
		return nil, fmt.Errorf("%w: %w", oidcproviders.ErrIncompleteDiscovery, ErrDiscoveryIncomplete)
	}

	return &oidcproviders.Discovery{
		Issuer:                issuerURL,
		AuthorizationEndpoint: doc.AuthorizationEndpoint,
		TokenEndpoint:         doc.TokenEndpoint,
		UserInfoEndpoint:      doc.UserInfoEndpoint,
		EndSessionEndpoint:    doc.EndSessionEndpoint,
	}, nil
}
