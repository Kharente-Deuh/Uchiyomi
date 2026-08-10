// SPDX-License-Identifier: AGPL-3.0-or-later

package auth_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/federatedidentities"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
)

type revalidationSessionService struct {
	gotUserID     uuid.UUID
	gotProviderID uuid.UUID
	calls         int
	err           error
}

func (s *revalidationSessionService) Create(
	context.Context,
	sessions.CreateSessionOpts,
) (*sessions.IssuedSession, error) {
	panic("Create is not used by revalidation")
}

func (s *revalidationSessionService) Authenticate(context.Context, string) (*sessions.AuthenticatedSession, error) {
	panic("Authenticate is not used by revalidation")
}

func (s *revalidationSessionService) Revoke(context.Context, string) error {
	panic("Revoke is not used by revalidation")
}

func (s *revalidationSessionService) RevokeAllForUser(context.Context, uuid.UUID) error {
	panic("RevokeAllForUser is not used by revalidation")
}

func (s *revalidationSessionService) RevokeForProvider(_ context.Context, userID, providerID uuid.UUID) error {
	s.calls++
	s.gotUserID = userID
	s.gotProviderID = providerID

	return s.err
}

type revalidationOIDCClient struct {
	tokenSet *oidcproviders.TokenSet
	err      error
}

func (c *revalidationOIDCClient) AuthCodeURL(
	context.Context,
	oidcproviders.OIDCProvider,
	oidcproviders.AuthCodeParams,
) (string, error) {
	panic("AuthCodeURL is not used by revalidation")
}

func (c *revalidationOIDCClient) Exchange(
	context.Context,
	oidcproviders.OIDCProvider,
	string, string, string, string,
) (*oidcproviders.TokenSet, error) {
	panic("Exchange is not used by revalidation")
}

func (c *revalidationOIDCClient) Refresh(
	context.Context,
	oidcproviders.OIDCProvider,
	string,
) (*oidcproviders.TokenSet, error) {
	return c.tokenSet, c.err
}

func (c *revalidationOIDCClient) EndSessionURL(
	context.Context,
	oidcproviders.OIDCProvider,
	string,
) (string, bool, error) {
	panic("EndSessionURL is not used by revalidation")
}

func newRevalidationAuth(t *testing.T, f *fakes, ss sessions.SessionService, oc oidcproviders.Client) *auth.Service {
	t.Helper()

	s, err := auth.New(f.cfg, auth.Deps{
		HashService:                   f.hs,
		UsersRepository:               f.ur,
		PwdRepository:                 f.pr,
		Transactor:                    f.tr,
		SessionService:                ss,
		OIDCProvidersRepository:       f.opr,
		FederatedIdentitiesRepository: f.fir,
		OIDCClient:                    oc,
		StateCipher:                   f.sc,
		Logger:                        slog.New(slog.DiscardHandler),
		Now:                           func() time.Time { return f.now },
	})
	if err != nil {
		t.Fatalf("auth.New: %v", err)
	}

	return s
}

func TestRevalidateFederatedIdentityInvalidGrantRevokesProviderSessions(t *testing.T) {
	t.Parallel()

	f := newFakes()
	ss := &revalidationSessionService{}

	sealed, err := f.sc.Seal([]byte("refresh-token"))
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	fi := federatedidentities.FederatedIdentity{
		ID:              uuid.New(),
		UserID:          f.ur.user.ID,
		ProviderID:      f.opr.provider.ID,
		RefreshTokenEnc: sealed,
	}

	s := newRevalidationAuth(t, f, ss, &revalidationOIDCClient{err: oidcproviders.ErrInvalidGrant})

	if err := s.RevalidateFederatedIdentity(context.Background(), *f.opr.provider, fi); err != nil {
		t.Fatalf("RevalidateFederatedIdentity: %v", err)
	}

	if ss.calls != 1 {
		t.Errorf("RevokeForProvider called %d times, want 1", ss.calls)
	}

	if ss.gotUserID != fi.UserID || ss.gotProviderID != fi.ProviderID {
		t.Errorf("RevokeForProvider(%v, %v), want (%v, %v)",
			ss.gotUserID, ss.gotProviderID, fi.UserID, fi.ProviderID)
	}

	if f.fir.updateCalls != 1 {
		t.Errorf("Update called %d times, want 1", f.fir.updateCalls)
	}

	if !f.fir.gotUpdateOpts.ClearRefreshToken {
		t.Error("ClearRefreshToken = false, want true")
	}
}

func TestRevalidateFederatedIdentitySuccessUpdatesClaimsAndAdmin(t *testing.T) {
	t.Parallel()

	f := newFakes()
	role := adminRole
	f.opr.provider.RoleClaim = &role
	f.opr.provider.AdminValues = []string{adminRole}

	sealed, err := f.sc.Seal([]byte("refresh-token"))
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	fi := federatedidentities.FederatedIdentity{
		ID:              uuid.New(),
		UserID:          f.ur.user.ID,
		ProviderID:      f.opr.provider.ID,
		RefreshTokenEnc: sealed,
	}

	tokenSet := &oidcproviders.TokenSet{
		Subject:      "sub-1",
		RefreshToken: "rotated",
		Claims: map[string]any{
			usernameClaimKey: userName,
			roleClaimKey:     adminRole,
		},
	}

	s := newRevalidationAuth(t, f, &revalidationSessionService{}, &revalidationOIDCClient{tokenSet: tokenSet})

	if err := s.RevalidateFederatedIdentity(context.Background(), *f.opr.provider, fi); err != nil {
		t.Fatalf("RevalidateFederatedIdentity: %v", err)
	}

	if f.fir.updateCalls != 1 {
		t.Fatalf("Update called %d times, want 1", f.fir.updateCalls)
	}

	if !f.fir.gotUpdateOpts.SetRefreshToken {
		t.Error("SetRefreshToken = false, want true")
	}

	if f.fir.gotUpdateOpts.Claims[roleClaimKey] != adminRole {
		t.Errorf("Claims = %v", f.fir.gotUpdateOpts.Claims)
	}
}

func TestRevalidateFederatedIdentityTransientFailureDoesNotRevoke(t *testing.T) {
	t.Parallel()

	f := newFakes()
	ss := &revalidationSessionService{}

	sealed, err := f.sc.Seal([]byte("refresh-token"))
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	fi := federatedidentities.FederatedIdentity{
		ID:              uuid.New(),
		UserID:          f.ur.user.ID,
		ProviderID:      f.opr.provider.ID,
		RefreshTokenEnc: sealed,
	}

	sentinel := errors.New("gateway timeout")
	s := newRevalidationAuth(t, f, ss, &revalidationOIDCClient{err: sentinel})

	err = s.RevalidateFederatedIdentity(context.Background(), *f.opr.provider, fi)
	if !errors.Is(err, sentinel) {
		t.Fatalf("RevalidateFederatedIdentity = %v, want %v", err, sentinel)
	}

	if ss.calls != 0 {
		t.Errorf("RevokeForProvider called %d times, want 0", ss.calls)
	}

	if f.fir.updateCalls != 0 {
		t.Errorf("Update called %d times, want 0", f.fir.updateCalls)
	}
}

func TestRevalidateFederatedIdentitySkipsMissingRefreshToken(t *testing.T) {
	t.Parallel()

	f := newFakes()
	ss := &revalidationSessionService{}

	fi := federatedidentities.FederatedIdentity{
		ID:         uuid.New(),
		UserID:     f.ur.user.ID,
		ProviderID: f.opr.provider.ID,
	}

	s := newRevalidationAuth(t, f, ss, &revalidationOIDCClient{})

	if err := s.RevalidateFederatedIdentity(context.Background(), *f.opr.provider, fi); err != nil {
		t.Fatalf("RevalidateFederatedIdentity: %v", err)
	}

	if ss.calls != 0 {
		t.Errorf("RevokeForProvider called %d times, want 0", ss.calls)
	}
}
