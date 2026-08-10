// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidc"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"golang.org/x/oauth2"
)

const stateTokenBytes = 32

type oidcStatePayload struct {
	IssuedAt   time.Time
	State      string
	Nonce      string
	Verifier   string
	Redirect   string
	ProviderID uuid.UUID
}

func (s *Service) StartOIDCLogin(ctx context.Context, opts StartOIDCLoginOpts) (*OIDCStart, error) {
	provider, err := s.deps.OIDCProvidersRepository.GetByID(ctx, opts.ProviderID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrOIDCUnavailable
		}

		return nil, fmt.Errorf("%w: %w", ErrOIDCUnavailable, err)
	}

	state, err := randomToken()
	if err != nil {
		return nil, fmt.Errorf("randomToken: %w", err)
	}

	nonce, err := randomToken()
	if err != nil {
		return nil, fmt.Errorf("randomToken: %w", err)
	}

	verifier := oauth2.GenerateVerifier()

	authCodeURL, err := s.deps.OIDCClient.AuthCodeURL(ctx, *provider, oidcproviders.AuthCodeParams{
		RedirectURI: s.cfg.RedirectURI,
		State:       state,
		Nonce:       nonce,
		Verifier:    verifier,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOIDCUnavailable, err)
	}

	cookieValue, err := s.encodeState(oidcStatePayload{
		ProviderID: provider.ID,
		State:      state,
		Nonce:      nonce,
		Verifier:   verifier,
		Redirect:   opts.Redirect,
		IssuedAt:   s.now(),
	})
	if err != nil {
		return nil, fmt.Errorf("s.encodeState: %w", err)
	}

	return &OIDCStart{
		AuthCodeURL:      authCodeURL,
		StateCookieValue: cookieValue,
		ExpiresAt:        s.now().Add(s.cfg.StateCookieTTL),
	}, nil
}

func (s *Service) FinishOIDCLogin(ctx context.Context, opts FinishOIDCLoginOpts) (*OIDCLoginResult, error) {
	payload, err := s.decodeState(opts.StateCookieValue)
	if err != nil {
		return nil, err
	}

	if opts.State == "" || payload.State != opts.State {
		return nil, ErrOIDCState
	}

	if s.now().Sub(payload.IssuedAt) > s.cfg.StateCookieTTL {
		return nil, ErrOIDCState
	}

	if opts.ErrorParam == "access_denied" {
		return nil, ErrOIDCDenied
	}

	if opts.ErrorParam != "" {
		return nil, ErrOIDCUnavailable
	}

	provider, err := s.deps.OIDCProvidersRepository.GetByID(ctx, payload.ProviderID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrOIDCUnavailable
		}

		return nil, fmt.Errorf("%w: %w", ErrOIDCUnavailable, err)
	}

	tokenSet, err := s.deps.OIDCClient.Exchange(
		ctx, *provider, opts.Code, payload.Verifier, payload.Nonce, s.cfg.RedirectURI,
	)
	if err != nil {
		if errors.Is(err, oidc.ErrNonceMismatch) {
			return nil, ErrOIDCState
		}

		return nil, fmt.Errorf("%w: %w", ErrOIDCUnavailable, err)
	}

	username, isAdmin, allowed := s.mapClaims(*provider, tokenSet)
	if !allowed {
		return nil, ErrOIDCNotAllowed
	}

	user, err := s.resolveIdentity(ctx, *provider, tokenSet, username)
	if err != nil {
		return nil, err
	}

	if err := s.syncIsAdmin(ctx, *provider, user, isAdmin); err != nil {
		return nil, err
	}

	session, err := s.deps.SessionService.Create(ctx, sessions.CreateSessionOpts{
		UserID:      user.ID,
		AuthMethod:  sessions.AuthMethodOIDC,
		ProviderID:  &provider.ID,
		ProviderSID: nilIfEmpty(tokenSet.SID),
	})
	if err != nil {
		return nil, fmt.Errorf("s.deps.SessionService.Create: %w", err)
	}

	return &OIDCLoginResult{Session: session, Redirect: safeRedirectPath(payload.Redirect)}, nil
}

func (s *Service) syncIsAdmin(
	ctx context.Context,
	provider oidcproviders.OIDCProvider,
	user *users.User,
	isAdmin bool,
) error {
	if !mapsAdminRole(provider) {
		return nil
	}

	if user.IsAdmin && !isAdmin {
		admins, err := s.deps.UsersRepository.CountAdmins(ctx)
		if err != nil {
			return fmt.Errorf("s.deps.UsersRepository.CountAdmins: %w", err)
		}

		if admins <= 1 {
			s.deps.Logger.WarnContext(ctx,
				"kept the last admin: the oidc provider maps them to a non-admin role",
				"userID", user.ID, "providerID", provider.ID,
			)

			return nil
		}
	}

	if _, err := s.deps.UsersRepository.Update(ctx, users.UpdateUserOpts{ID: user.ID, IsAdmin: isAdmin}); err != nil {
		return fmt.Errorf("s.deps.UsersRepository.Update: %w", err)
	}

	return nil
}

func (s *Service) now() time.Time {
	return s.deps.Now()
}

func (s *Service) encodeState(payload oidcStatePayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("json.Marshal: %w", err)
	}

	sealed, err := s.deps.StateCipher.Seal(raw)
	if err != nil {
		return "", fmt.Errorf("s.deps.StateCipher.Seal: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

func (s *Service) decodeState(cookieValue string) (*oidcStatePayload, error) {
	if cookieValue == "" {
		return nil, ErrOIDCState
	}

	sealed, err := base64.RawURLEncoding.DecodeString(cookieValue)
	if err != nil {
		return nil, ErrOIDCState
	}

	raw, err := s.deps.StateCipher.Open(sealed)
	if err != nil {
		return nil, ErrOIDCState
	}

	var payload oidcStatePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, ErrOIDCState
	}

	return &payload, nil
}

func randomToken() (string, error) {
	buf := make([]byte, stateTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("rand.Read: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

func safeRedirectPath(p string) string {
	u, err := url.Parse(p)
	if err != nil || u.Scheme != "" || u.Host != "" ||
		!strings.HasPrefix(u.Path, "/") || strings.HasPrefix(u.Path, "//") || strings.Contains(u.Path, "\\") {
		return "/"
	}

	return p
}
