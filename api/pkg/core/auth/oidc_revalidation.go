// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/federatedidentities"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
)

func (s *Service) ApplySuccessfulRevalidation(
	ctx context.Context,
	provider oidcproviders.OIDCProvider,
	fi federatedidentities.FederatedIdentity,
	tokenSet *oidcproviders.TokenSet,
) error {
	_, isAdmin, allowed := s.mapClaims(provider, tokenSet)
	if !allowed {
		return s.revokeFederatedAccess(ctx, fi)
	}

	user, err := s.deps.UsersRepository.GetByID(ctx, fi.UserID)
	if err != nil {
		return fmt.Errorf("s.deps.UsersRepository.GetByID: %w", err)
	}

	if err := s.syncIsAdmin(ctx, provider, user, isAdmin); err != nil {
		return err
	}

	refreshTokenEnc, lastValidatedAt, err := s.federatedIdentityFieldsFromTokenSet(tokenSet)
	if err != nil {
		return err
	}

	updateOpts := federatedidentities.UpdateFederatedIdentityOpts{
		ID:              fi.ID,
		Claims:          tokenSet.Claims,
		LastValidatedAt: lastValidatedAt,
	}

	if len(refreshTokenEnc) > 0 {
		updateOpts.RefreshTokenEnc = refreshTokenEnc
		updateOpts.SetRefreshToken = true
	}

	if err := s.deps.FederatedIdentitiesRepository.Update(ctx, updateOpts); err != nil {
		return fmt.Errorf("s.deps.FederatedIdentitiesRepository.Update: %w", err)
	}

	return nil
}

func (s *Service) ApplyInvalidGrant(ctx context.Context, fi federatedidentities.FederatedIdentity) error {
	return s.revokeFederatedAccess(ctx, fi)
}

func (s *Service) revokeFederatedAccess(ctx context.Context, fi federatedidentities.FederatedIdentity) error {
	if err := s.deps.SessionService.RevokeForProvider(ctx, fi.UserID, fi.ProviderID); err != nil {
		return fmt.Errorf("s.deps.SessionService.RevokeForProvider: %w", err)
	}

	if err := s.deps.FederatedIdentitiesRepository.Update(ctx, federatedidentities.UpdateFederatedIdentityOpts{
		ID:                fi.ID,
		ClearRefreshToken: true,
	}); err != nil {
		return fmt.Errorf("s.deps.FederatedIdentitiesRepository.Update: %w", err)
	}

	return nil
}

func (s *Service) RevalidateFederatedIdentity(
	ctx context.Context,
	provider oidcproviders.OIDCProvider,
	fi federatedidentities.FederatedIdentity,
) error {
	if len(fi.RefreshTokenEnc) == 0 {
		return nil
	}

	refreshToken, err := s.deps.StateCipher.Open(fi.RefreshTokenEnc)
	if err != nil {
		return fmt.Errorf("s.deps.StateCipher.Open: %w", err)
	}

	tokenSet, err := s.deps.OIDCClient.Refresh(ctx, provider, string(refreshToken))
	if err != nil {
		if errors.Is(err, oidcproviders.ErrInvalidGrant) {
			return s.ApplyInvalidGrant(ctx, fi)
		}

		return fmt.Errorf("s.deps.OIDCClient.Refresh: %w", err)
	}

	return s.ApplySuccessfulRevalidation(ctx, provider, fi, tokenSet)
}
