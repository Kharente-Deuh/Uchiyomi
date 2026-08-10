// SPDX-License-Identifier: AGPL-3.0-or-later

package oidc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	"golang.org/x/oauth2"
)

var _ oidcproviders.Client = (*Client)(nil)

var (
	ErrClientUnavailable = errors.New("oidc client: provider is unavailable")
	ErrExchangeFailed    = errors.New("oidc client: token exchange failed")
	ErrRefreshFailed     = errors.New("oidc client: token refresh failed")
	ErrIDTokenInvalid    = errors.New("oidc client: id token verification failed")
	ErrNonceMismatch     = errors.New("oidc client: id token nonce mismatch")
)

const (
	defaultClientTimeout = 10 * time.Second
	defaultCacheTTL      = 15 * time.Minute

	openIDScope        = "openid"
	offlineAccessScope = "offline_access"
	sidClaim           = "sid"

	logoutTokenEvent   = "http://schemas.openid.net/event/backchannel-logout"
	logoutTokenIATSkew = 5 * time.Minute
)

type Decrypter interface {
	Open(ciphertext []byte) ([]byte, error)
}

type ClientConfig struct {
	Timeout  time.Duration
	CacheTTL time.Duration
}

func (cfg *ClientConfig) Validate() error {
	if cfg.Timeout < 0 {
		return errors.New("timeout must not be negative")
	}

	if cfg.CacheTTL < 0 {
		return errors.New("cacheTTL must not be negative")
	}

	return nil
}

type ClientDeps struct {
	HTTPClient *http.Client
	Cipher     Decrypter
}

func (deps *ClientDeps) Validate() error {
	if deps.HTTPClient == nil {
		return errors.New("httpClient is required")
	}

	if deps.Cipher == nil {
		return errors.New("cipher is required")
	}

	return nil
}

type cacheEntry struct {
	provider  *gooidc.Provider
	fetchedAt time.Time
}

type Client struct {
	cache map[uuid.UUID]cacheEntry
	deps  ClientDeps
	cfg   ClientConfig
	mu    sync.RWMutex
}

func NewClient(cfg ClientConfig, deps ClientDeps) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = defaultClientTimeout
	}

	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = defaultCacheTTL
	}

	return &Client{cfg: cfg, deps: deps, cache: make(map[uuid.UUID]cacheEntry)}, nil
}

func (c *Client) AuthCodeURL(
	ctx context.Context,
	provider oidcproviders.OIDCProvider,
	params oidcproviders.AuthCodeParams,
) (string, error) {
	p, err := c.providerFor(ctx, provider)
	if err != nil {
		return "", fmt.Errorf("client.providerFor: %w", err)
	}

	cfg := oauth2.Config{
		ClientID: provider.ClientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:  p.Endpoint().AuthURL,
			TokenURL: p.Endpoint().TokenURL,
		},
		RedirectURL: params.RedirectURI,
		Scopes:      dedup(provider.Scopes, openIDScope, offlineAccessScope),
	}

	return cfg.AuthCodeURL(
		params.State,
		oauth2.SetAuthURLParam("nonce", params.Nonce),
		oauth2.S256ChallengeOption(params.Verifier),
	), nil
}

func (c *Client) Exchange(
	ctx context.Context,
	provider oidcproviders.OIDCProvider,
	code, verifier, nonce, redirectURI string,
) (*oidcproviders.TokenSet, error) {
	if nonce == "" {
		return nil, fmt.Errorf("%w: nonce must not be empty", ErrNonceMismatch)
	}

	if verifier == "" {
		return nil, fmt.Errorf("%w: verifier must not be empty", ErrExchangeFailed)
	}

	p, err := c.providerFor(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("client.providerFor: %w", err)
	}

	secret, err := c.deps.Cipher.Open(provider.ClientSecretEnc)
	if err != nil {
		return nil, fmt.Errorf("cipher.Open: %w", err)
	}

	cfg := oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: string(secret),
		Endpoint: oauth2.Endpoint{
			AuthURL:  p.Endpoint().AuthURL,
			TokenURL: p.Endpoint().TokenURL,
		},
		RedirectURL: redirectURI,
	}

	exchangeCtx := gooidc.ClientContext(ctx, c.deps.HTTPClient)

	token, err := cfg.Exchange(exchangeCtx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrExchangeFailed, err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, fmt.Errorf("%w: token response is missing the id_token", ErrExchangeFailed)
	}

	idTokenVerifier := p.Verifier(&gooidc.Config{ClientID: provider.ClientID})

	idToken, err := idTokenVerifier.Verify(exchangeCtx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrIDTokenInvalid, err)
	}

	if idToken.Nonce != nonce {
		return nil, ErrNonceMismatch
	}

	var claims map[string]any
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrIDTokenInvalid, err)
	}

	sid, _ := claims[sidClaim].(string)

	return &oidcproviders.TokenSet{
		Subject:      idToken.Subject,
		SID:          sid,
		Claims:       claims,
		RefreshToken: token.RefreshToken,
	}, nil
}

func (c *Client) Refresh(
	ctx context.Context,
	provider oidcproviders.OIDCProvider,
	refreshToken string,
) (*oidcproviders.TokenSet, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("%w: refresh token must not be empty", ErrRefreshFailed)
	}

	p, err := c.providerFor(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("client.providerFor: %w", err)
	}

	secret, err := c.deps.Cipher.Open(provider.ClientSecretEnc)
	if err != nil {
		return nil, fmt.Errorf("cipher.Open: %w", err)
	}

	cfg := oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: string(secret),
		Endpoint: oauth2.Endpoint{
			AuthURL:  p.Endpoint().AuthURL,
			TokenURL: p.Endpoint().TokenURL,
		},
		Scopes: dedup(provider.Scopes, openIDScope, offlineAccessScope),
	}

	refreshCtx := gooidc.ClientContext(ctx, c.deps.HTTPClient)

	token, err := cfg.TokenSource(refreshCtx, &oauth2.Token{RefreshToken: refreshToken}).Token()
	if err != nil {
		if retrieveErr, ok := err.(*oauth2.RetrieveError); ok && retrieveErr.ErrorCode == "invalid_grant" {
			return nil, oidcproviders.ErrInvalidGrant
		}

		return nil, fmt.Errorf("%w: %w", ErrRefreshFailed, err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, fmt.Errorf("%w: token response is missing the id_token", ErrRefreshFailed)
	}

	idTokenVerifier := p.Verifier(&gooidc.Config{ClientID: provider.ClientID})

	idToken, err := idTokenVerifier.Verify(refreshCtx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrIDTokenInvalid, err)
	}

	var claims map[string]any
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrIDTokenInvalid, err)
	}

	sid, _ := claims[sidClaim].(string)

	rotated := token.RefreshToken
	if rotated == "" {
		rotated = refreshToken
	}

	return &oidcproviders.TokenSet{
		Subject:      idToken.Subject,
		SID:          sid,
		Claims:       claims,
		RefreshToken: rotated,
	}, nil
}

func (c *Client) EndSessionURL(
	ctx context.Context,
	provider oidcproviders.OIDCProvider,
	postLogoutRedirectURI string,
) (string, bool, error) {
	p, err := c.providerFor(ctx, provider)
	if err != nil {
		return "", false, fmt.Errorf("client.providerFor: %w", err)
	}

	var claims struct {
		EndSessionEndpoint string `json:"end_session_endpoint"`
	}

	if err := p.Claims(&claims); err != nil {
		return "", false, fmt.Errorf("provider.Claims: %w", err)
	}

	if claims.EndSessionEndpoint == "" {
		return "", false, nil
	}

	u, err := url.Parse(claims.EndSessionEndpoint)
	if err != nil {
		return "", false, fmt.Errorf("url.Parse: %w", err)
	}

	q := u.Query()
	q.Set("client_id", provider.ClientID)
	q.Set("post_logout_redirect_uri", postLogoutRedirectURI)
	u.RawQuery = q.Encode()

	return u.String(), true, nil
}

func (c *Client) VerifyLogoutToken(
	ctx context.Context,
	provider oidcproviders.OIDCProvider,
	raw string,
) (*oidcproviders.LogoutToken, error) {
	p, err := c.providerFor(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("client.providerFor: %w", err)
	}

	verifier := p.Verifier(&gooidc.Config{
		ClientID:        provider.ClientID,
		SkipExpiryCheck: true,
	})

	token, err := verifier.Verify(ctx, raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", oidcproviders.ErrLogoutTokenInvalid, err)
	}

	var claims map[string]any
	if err := token.Claims(&claims); err != nil {
		return nil, fmt.Errorf("%w: %w", oidcproviders.ErrLogoutTokenInvalid, err)
	}

	if _, hasNonce := claims["nonce"]; hasNonce {
		return nil, fmt.Errorf("%w: nonce must not be present", oidcproviders.ErrLogoutTokenInvalid)
	}

	events, _ := claims["events"].(map[string]any)
	if events == nil {
		return nil, fmt.Errorf("%w: events claim is missing", oidcproviders.ErrLogoutTokenInvalid)
	}

	if _, ok := events[logoutTokenEvent]; !ok {
		return nil, fmt.Errorf("%w: backchannel-logout event is missing", oidcproviders.ErrLogoutTokenInvalid)
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, fmt.Errorf("%w: iat is missing or invalid", oidcproviders.ErrLogoutTokenInvalid)
	}

	iatTime := time.Unix(int64(iat), 0)
	now := time.Now()
	if iatTime.Before(now.Add(-logoutTokenIATSkew)) || iatTime.After(now.Add(logoutTokenIATSkew)) {
		return nil, fmt.Errorf("%w: iat is outside acceptable skew", oidcproviders.ErrLogoutTokenInvalid)
	}

	sid, _ := claims["sid"].(string)
	sub, _ := claims["sub"].(string)

	return &oidcproviders.LogoutToken{Subject: sub, SID: sid}, nil
}

func (c *Client) Evict(id uuid.UUID) {
	c.mu.Lock()
	delete(c.cache, id)
	c.mu.Unlock()
}

func (c *Client) providerFor(ctx context.Context, provider oidcproviders.OIDCProvider) (*gooidc.Provider, error) {
	c.mu.RLock()
	entry, ok := c.cache[provider.ID]
	c.mu.RUnlock()

	if ok && time.Since(entry.fetchedAt) < c.cfg.CacheTTL {
		return entry.provider, nil
	}

	fetchCtx, cancel := context.WithTimeout(gooidc.ClientContext(ctx, c.deps.HTTPClient), c.cfg.Timeout)
	defer cancel()

	fetched, err := gooidc.NewProvider(fetchCtx, provider.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientUnavailable, err)
	}

	c.mu.Lock()
	c.cache[provider.ID] = cacheEntry{provider: fetched, fetchedAt: time.Now()}
	c.mu.Unlock()

	return fetched, nil
}

func dedup(scopes []string, extra ...string) []string {
	seen := make(map[string]bool, len(scopes)+len(extra))
	out := make([]string, 0, len(scopes)+len(extra))

	for _, s := range append(append([]string{}, scopes...), extra...) {
		if seen[s] {
			continue
		}

		seen[s] = true

		out = append(out, s)
	}

	return out
}
