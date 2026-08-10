// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
)

func (s *Service) Logout(ctx context.Context, opts LogoutOpts) (*LogoutResult, error) {
	if err := s.deps.SessionService.Revoke(ctx, opts.Token); err != nil {
		return nil, fmt.Errorf("s.deps.SessionService.Revoke: %w", err)
	}

	return &LogoutResult{EndSessionURL: s.buildEndSessionURL(ctx, opts.Session)}, nil
}

func (s *Service) buildEndSessionURL(ctx context.Context, session sessions.Session) string {
	if session.AuthMethod != sessions.AuthMethodOIDC || session.ProviderID == nil {
		return ""
	}

	provider, err := s.deps.OIDCProvidersRepository.GetByID(ctx, *session.ProviderID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ""
		}

		s.deps.Logger.WarnContext(ctx, "failed to load oidc provider for logout", "err", err)

		return ""
	}

	endSessionURL, supported, err := s.deps.OIDCClient.EndSessionURL(
		ctx,
		*provider,
		s.cfg.PublicURL+"/login",
	)
	if err != nil {
		s.deps.Logger.WarnContext(ctx, "failed to build oidc end session url", "err", err)

		return ""
	}

	if !supported {
		return ""
	}

	return endSessionURL
}
