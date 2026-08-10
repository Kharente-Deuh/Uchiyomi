// SPDX-License-Identifier: AGPL-3.0-or-later

package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/federatedidentities"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
)

func (s *Service) sealRefreshToken(refreshToken string) ([]byte, error) {
	if refreshToken == "" {
		return nil, nil
	}

	sealed, err := s.deps.StateCipher.Seal([]byte(refreshToken))
	if err != nil {
		return nil, fmt.Errorf("s.deps.StateCipher.Seal: %w", err)
	}

	return sealed, nil
}

func (s *Service) federatedIdentityFieldsFromTokenSet(
	tokenSet *oidcproviders.TokenSet,
) (refreshTokenEnc []byte, lastValidatedAt time.Time, err error) {
	refreshTokenEnc, err = s.sealRefreshToken(tokenSet.RefreshToken)
	if err != nil {
		return nil, time.Time{}, err
	}

	if len(refreshTokenEnc) > 0 {
		lastValidatedAt = s.now()
	}

	return refreshTokenEnc, lastValidatedAt, nil
}

func (s *Service) createFederatedIdentity(
	ctx context.Context,
	opts federatedidentities.CreateFederatedIdentityOpts,
) error {
	_, err := s.deps.FederatedIdentitiesRepository.Create(ctx, opts)
	if err != nil {
		return fmt.Errorf("s.deps.FederatedIdentitiesRepository.Create: %w", err)
	}

	return nil
}
