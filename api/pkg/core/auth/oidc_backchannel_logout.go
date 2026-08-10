// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/federatedidentities"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidc"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/logging"
)

type verifiedLogoutToken struct {
	*oidcproviders.LogoutToken
	provider oidcproviders.OIDCProvider
}

func (s *Service) BackchannelLogout(ctx context.Context, rawToken string) error {
	logoutToken, err := s.verifyLogoutToken(ctx, rawToken)
	if err != nil {
		return err
	}

	provider := logoutToken.provider

	switch {
	case logoutToken.SID != "":
		if err := s.deps.SessionService.RevokeByProviderAndSID(ctx, provider.ID, logoutToken.SID); err != nil {
			s.deps.Logger.WarnContext(ctx, "failed to revoke sessions by sid during backchannel logout", logging.Err(err))
		}

		if logoutToken.Subject != "" {
			if err := s.clearFederatedRefreshToken(ctx, provider.ID, logoutToken.Subject); err != nil {
				s.deps.Logger.WarnContext(ctx, "failed to clear refresh token during backchannel logout", logging.Err(err))
			}
		}
	case logoutToken.Subject != "":
		fi, err := s.deps.FederatedIdentitiesRepository.Get(ctx, federatedidentities.GetFederatedIdentityOpts{
			ProviderID: provider.ID,
			Subject:    logoutToken.Subject,
		})
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil
			}

			s.deps.Logger.WarnContext(ctx, "failed to load federated identity during backchannel logout", logging.Err(err))

			return nil
		}

		if err := s.revokeFederatedAccess(ctx, *fi); err != nil {
			s.deps.Logger.WarnContext(ctx, "failed to revoke federated access during backchannel logout", logging.Err(err))
		}
	}

	return nil
}

func (s *Service) verifyLogoutToken(ctx context.Context, rawToken string) (*verifiedLogoutToken, error) {
	iss, err := parseUnverifiedIssuer(rawToken)
	if err != nil {
		return nil, ErrLogoutTokenInvalid
	}

	provider, err := s.deps.OIDCProvidersRepository.GetByIssuerURL(ctx, iss)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrLogoutTokenInvalid
		}

		return nil, fmt.Errorf("s.deps.OIDCProvidersRepository.GetByIssuerURL: %w", err)
	}

	token, err := s.deps.OIDCClient.VerifyLogoutToken(ctx, *provider, rawToken)
	if err != nil {
		if errors.Is(err, oidcproviders.ErrLogoutTokenInvalid) || errors.Is(err, oidc.ErrClientUnavailable) {
			return nil, ErrLogoutTokenInvalid
		}

		return nil, fmt.Errorf("s.deps.OIDCClient.VerifyLogoutToken: %w", err)
	}

	return &verifiedLogoutToken{provider: *provider, LogoutToken: token}, nil
}

func parseUnverifiedIssuer(raw string) (string, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("decode JWT payload: %w", err)
	}

	var claims struct {
		Iss string `json:"iss"`
	}

	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("unmarshal JWT payload: %w", err)
	}

	if claims.Iss == "" {
		return "", fmt.Errorf("iss claim is missing")
	}

	return claims.Iss, nil
}

func (s *Service) clearFederatedRefreshToken(ctx context.Context, providerID uuid.UUID, subject string) error {
	fi, err := s.deps.FederatedIdentitiesRepository.Get(ctx, federatedidentities.GetFederatedIdentityOpts{
		ProviderID: providerID,
		Subject:    subject,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}

		return fmt.Errorf("s.deps.FederatedIdentitiesRepository.Get: %w", err)
	}

	if err := s.deps.FederatedIdentitiesRepository.Update(ctx, federatedidentities.UpdateFederatedIdentityOpts{
		ID:                fi.ID,
		ClearRefreshToken: true,
	}); err != nil {
		return fmt.Errorf("s.deps.FederatedIdentitiesRepository.Update: %w", err)
	}

	return nil
}
