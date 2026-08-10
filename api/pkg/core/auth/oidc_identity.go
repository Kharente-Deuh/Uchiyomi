// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/federatedidentities"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
)

func (s *Service) resolveIdentity(
	ctx context.Context,
	provider oidcproviders.OIDCProvider,
	tokenSet *oidcproviders.TokenSet,
	username string,
) (*users.User, error) {
	fi, err := s.deps.FederatedIdentitiesRepository.Get(ctx, federatedidentities.GetFederatedIdentityOpts{
		Subject:    tokenSet.Subject,
		ProviderID: provider.ID,
	})
	if err == nil {
		return s.reuseIdentity(ctx, *fi, tokenSet)
	}

	if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("s.deps.FederatedIdentitiesRepository.Get: %w", err)
	}

	user, err := s.deps.UsersRepository.GetByUsername(ctx, username)
	if err == nil {
		return s.linkIdentity(ctx, provider, tokenSet, user)
	}

	if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("s.deps.UsersRepository.GetByUsername: %w", err)
	}

	if !provider.AutoProvision {
		return nil, ErrOIDCNoAccount
	}

	return s.provisionIdentity(ctx, provider, tokenSet, username)
}

func (s *Service) reuseIdentity(
	ctx context.Context,
	fi federatedidentities.FederatedIdentity,
	tokenSet *oidcproviders.TokenSet,
) (*users.User, error) {
	user, err := s.deps.UsersRepository.GetByID(ctx, fi.UserID)
	if err != nil {
		return nil, fmt.Errorf("s.deps.UsersRepository.GetByID: %w", err)
	}

	refreshTokenEnc, lastValidatedAt, err := s.federatedIdentityFieldsFromTokenSet(tokenSet)
	if err != nil {
		return nil, err
	}

	updateOpts := federatedidentities.UpdateFederatedIdentityOpts{
		ID:          fi.ID,
		Claims:      tokenSet.Claims,
		LastLoginAt: s.now(),
	}

	if len(refreshTokenEnc) > 0 {
		updateOpts.RefreshTokenEnc = refreshTokenEnc
		updateOpts.SetRefreshToken = true
		updateOpts.LastValidatedAt = lastValidatedAt
	}

	err = s.deps.FederatedIdentitiesRepository.Update(ctx, updateOpts)
	if err != nil {
		return nil, fmt.Errorf("s.deps.FederatedIdentitiesRepository.Update: %w", err)
	}

	return user, nil
}

func (s *Service) linkIdentity(
	ctx context.Context,
	provider oidcproviders.OIDCProvider,
	tokenSet *oidcproviders.TokenSet,
	user *users.User,
) (*users.User, error) {
	refreshTokenEnc, lastValidatedAt, err := s.federatedIdentityFieldsFromTokenSet(tokenSet)
	if err != nil {
		return nil, err
	}

	err = s.createFederatedIdentity(ctx, federatedidentities.CreateFederatedIdentityOpts{
		Subject:         tokenSet.Subject,
		ProviderID:      provider.ID,
		UserID:          user.ID,
		Claims:          tokenSet.Claims,
		RefreshTokenEnc: refreshTokenEnc,
		LastValidatedAt: lastValidatedAt,
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) provisionIdentity(
	ctx context.Context,
	provider oidcproviders.OIDCProvider,
	tokenSet *oidcproviders.TokenSet,
	username string,
) (*users.User, error) {
	var user *users.User

	err := s.deps.Transactor.WithinTx(ctx, transaction.TxOpts{}, func(ctx context.Context) error {
		var err error

		user, err = s.deps.UsersRepository.Create(ctx, users.CreateUserOpts{Name: username, IsAdmin: false})
		if err != nil {
			return fmt.Errorf("s.deps.UsersRepository.Create: %w", err)
		}

		refreshTokenEnc, lastValidatedAt, err := s.federatedIdentityFieldsFromTokenSet(tokenSet)
		if err != nil {
			return err
		}

		_, err = s.deps.FederatedIdentitiesRepository.Create(ctx, federatedidentities.CreateFederatedIdentityOpts{
			Subject:         tokenSet.Subject,
			ProviderID:      provider.ID,
			UserID:          user.ID,
			Claims:          tokenSet.Claims,
			RefreshTokenEnc: refreshTokenEnc,
			LastValidatedAt: lastValidatedAt,
		})
		if err != nil {
			return fmt.Errorf("s.deps.FederatedIdentitiesRepository.Create: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("s.deps.Transactor.WithinTx: %w", err)
	}

	return user, nil
}
