// SPDX-License-Identifier: AGPL-3.0-or-later

package auth_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/federatedidentities"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidc"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
)

const testLogoutSubject = "sub1"

func fakeLogoutTokenJWT(iss string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`))
	payload, _ := json.Marshal(map[string]string{"iss": iss})
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)

	return header + "." + payloadB64 + ".signature"
}

func TestBackchannelLogoutRevokesBySIDAndClearsRefreshToken(t *testing.T) {
	t.Parallel()

	f := newFakes()
	fiID := uuid.New()
	f.fir.fi = &federatedidentities.FederatedIdentity{
		ID:         fiID,
		Subject:    testLogoutSubject,
		ProviderID: f.opr.provider.ID,
		UserID:     f.ur.user.ID,
	}
	f.fir.getErr = nil
	f.oc.logoutToken = &oidcproviders.LogoutToken{SID: "s1", Subject: testLogoutSubject}

	raw := fakeLogoutTokenJWT(f.opr.provider.IssuerURL)

	if err := f.svc(t).BackchannelLogout(context.Background(), raw); err != nil {
		t.Fatalf("BackchannelLogout: %v", err)
	}

	if f.ss.revokeBySIDCalls != 1 {
		t.Errorf("RevokeByProviderAndSID called %d times, want 1", f.ss.revokeBySIDCalls)
	}

	if f.ss.gotRevokeBySIDProvider != f.opr.provider.ID || f.ss.gotRevokeBySID != "s1" {
		t.Errorf("RevokeByProviderAndSID(%v, %q), want (%v, %q)",
			f.ss.gotRevokeBySIDProvider, f.ss.gotRevokeBySID, f.opr.provider.ID, "s1")
	}

	if f.fir.updateCalls != 1 {
		t.Errorf("FederatedIdentitiesRepository.Update called %d times, want 1", f.fir.updateCalls)
	}

	if !f.fir.gotUpdateOpts.ClearRefreshToken {
		t.Error("ClearRefreshToken = false, want true")
	}

	if f.ss.revokeForProviderCalls != 0 {
		t.Errorf("RevokeForProvider called %d times, want 0", f.ss.revokeForProviderCalls)
	}
}

func TestBackchannelLogoutRevokesBySubjectOnly(t *testing.T) {
	t.Parallel()

	f := newFakes()
	fiID := uuid.New()
	f.fir.fi = &federatedidentities.FederatedIdentity{
		ID:         fiID,
		Subject:    testLogoutSubject,
		ProviderID: f.opr.provider.ID,
		UserID:     f.ur.user.ID,
	}
	f.fir.getErr = nil
	f.oc.logoutToken = &oidcproviders.LogoutToken{Subject: testLogoutSubject}

	raw := fakeLogoutTokenJWT(f.opr.provider.IssuerURL)

	if err := f.svc(t).BackchannelLogout(context.Background(), raw); err != nil {
		t.Fatalf("BackchannelLogout: %v", err)
	}

	if f.ss.revokeForProviderCalls != 1 {
		t.Errorf("RevokeForProvider called %d times, want 1", f.ss.revokeForProviderCalls)
	}

	if f.ss.gotRevokeUserID != f.ur.user.ID || f.ss.gotRevokeProviderID != f.opr.provider.ID {
		t.Errorf("RevokeForProvider(%v, %v), want (%v, %v)",
			f.ss.gotRevokeUserID, f.ss.gotRevokeProviderID, f.ur.user.ID, f.opr.provider.ID)
	}

	if f.fir.updateCalls != 1 {
		t.Errorf("FederatedIdentitiesRepository.Update called %d times, want 1", f.fir.updateCalls)
	}

	if !f.fir.gotUpdateOpts.ClearRefreshToken {
		t.Error("ClearRefreshToken = false, want true")
	}

	if f.ss.revokeBySIDCalls != 0 {
		t.Errorf("RevokeByProviderAndSID called %d times, want 0", f.ss.revokeBySIDCalls)
	}
}

func TestBackchannelLogoutNoMatchStillSucceeds(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.fir.getErr = domain.ErrNotFound
	f.oc.logoutToken = &oidcproviders.LogoutToken{Subject: "unknown-sub"}

	raw := fakeLogoutTokenJWT(f.opr.provider.IssuerURL)

	if err := f.svc(t).BackchannelLogout(context.Background(), raw); err != nil {
		t.Fatalf("BackchannelLogout: %v", err)
	}

	if f.ss.revokeForProviderCalls != 0 {
		t.Errorf("RevokeForProvider called %d times, want 0", f.ss.revokeForProviderCalls)
	}

	if f.ss.revokeBySIDCalls != 0 {
		t.Errorf("RevokeByProviderAndSID called %d times, want 0", f.ss.revokeBySIDCalls)
	}

	if f.fir.updateCalls != 0 {
		t.Errorf("FederatedIdentitiesRepository.Update called %d times, want 0", f.fir.updateCalls)
	}
}

func TestBackchannelLogoutUnknownIssuer(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.opr.issuerErr = domain.ErrNotFound

	raw := fakeLogoutTokenJWT("https://unknown.example.com")

	err := f.svc(t).BackchannelLogout(context.Background(), raw)
	if !errors.Is(err, auth.ErrLogoutTokenInvalid) {
		t.Errorf("BackchannelLogout = %v, want ErrLogoutTokenInvalid", err)
	}

	if f.oc.verifyLogoutCalls != 0 {
		t.Errorf("VerifyLogoutToken called %d times, want 0", f.oc.verifyLogoutCalls)
	}
}

func TestBackchannelLogoutSIDOnlyRevokesSessionsNotRefresh(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.oc.logoutToken = &oidcproviders.LogoutToken{SID: "s1"}

	raw := fakeLogoutTokenJWT(f.opr.provider.IssuerURL)

	if err := f.svc(t).BackchannelLogout(context.Background(), raw); err != nil {
		t.Fatalf("BackchannelLogout: %v", err)
	}

	if f.ss.revokeBySIDCalls != 1 {
		t.Errorf("RevokeByProviderAndSID called %d times, want 1", f.ss.revokeBySIDCalls)
	}

	if f.ss.gotRevokeBySIDProvider != f.opr.provider.ID || f.ss.gotRevokeBySID != "s1" {
		t.Errorf("RevokeByProviderAndSID(%v, %q), want (%v, %q)",
			f.ss.gotRevokeBySIDProvider, f.ss.gotRevokeBySID, f.opr.provider.ID, "s1")
	}

	if f.fir.updateCalls != 0 {
		t.Errorf("FederatedIdentitiesRepository.Update called %d times, want 0", f.fir.updateCalls)
	}

	if f.fir.getCalls != 0 {
		t.Errorf("FederatedIdentitiesRepository.Get called %d times, want 0", f.fir.getCalls)
	}
}

func TestBackchannelLogoutNeitherSIDNorSub(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.oc.logoutToken = &oidcproviders.LogoutToken{}

	raw := fakeLogoutTokenJWT(f.opr.provider.IssuerURL)

	if err := f.svc(t).BackchannelLogout(context.Background(), raw); err != nil {
		t.Fatalf("BackchannelLogout: %v", err)
	}

	if f.ss.revokeBySIDCalls != 0 {
		t.Errorf("RevokeByProviderAndSID called %d times, want 0", f.ss.revokeBySIDCalls)
	}

	if f.ss.revokeForProviderCalls != 0 {
		t.Errorf("RevokeForProvider called %d times, want 0", f.ss.revokeForProviderCalls)
	}

	if f.fir.getCalls != 0 {
		t.Errorf("FederatedIdentitiesRepository.Get called %d times, want 0", f.fir.getCalls)
	}

	if f.fir.updateCalls != 0 {
		t.Errorf("FederatedIdentitiesRepository.Update called %d times, want 0", f.fir.updateCalls)
	}
}

func TestBackchannelLogoutRevokeBySIDFailureStillSucceeds(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.ss.revokeBySIDErr = fmt.Errorf("db down")
	f.oc.logoutToken = &oidcproviders.LogoutToken{SID: "s1"}

	raw := fakeLogoutTokenJWT(f.opr.provider.IssuerURL)

	if err := f.svc(t).BackchannelLogout(context.Background(), raw); err != nil {
		t.Fatalf("BackchannelLogout: %v, want nil", err)
	}

	if f.ss.revokeBySIDCalls != 1 {
		t.Errorf("RevokeByProviderAndSID called %d times, want 1", f.ss.revokeBySIDCalls)
	}
}

func TestBackchannelLogoutClientUnavailable(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.oc.verifyLogoutErr = oidc.ErrClientUnavailable

	raw := fakeLogoutTokenJWT(f.opr.provider.IssuerURL)

	err := f.svc(t).BackchannelLogout(context.Background(), raw)
	if !errors.Is(err, auth.ErrLogoutTokenInvalid) {
		t.Errorf("BackchannelLogout = %v, want ErrLogoutTokenInvalid", err)
	}
}

func TestBackchannelLogoutInvalidToken(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.oc.verifyLogoutErr = oidcproviders.ErrLogoutTokenInvalid

	raw := fakeLogoutTokenJWT(f.opr.provider.IssuerURL)

	err := f.svc(t).BackchannelLogout(context.Background(), raw)
	if !errors.Is(err, auth.ErrLogoutTokenInvalid) {
		t.Errorf("BackchannelLogout = %v, want ErrLogoutTokenInvalid", err)
	}
}

func TestBackchannelLogoutReplay(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.fir.fi = &federatedidentities.FederatedIdentity{
		ID:         uuid.New(),
		Subject:    testLogoutSubject,
		ProviderID: f.opr.provider.ID,
		UserID:     f.ur.user.ID,
	}
	f.fir.getErr = nil
	f.oc.logoutToken = &oidcproviders.LogoutToken{SID: "s1", Subject: testLogoutSubject}

	raw := fakeLogoutTokenJWT(f.opr.provider.IssuerURL)
	svc := f.svc(t)

	for range 2 {
		if err := svc.BackchannelLogout(context.Background(), raw); err != nil {
			t.Fatalf("BackchannelLogout: %v", err)
		}
	}

	if f.oc.verifyLogoutCalls != 2 {
		t.Errorf("VerifyLogoutToken called %d times, want 2", f.oc.verifyLogoutCalls)
	}

	if f.ss.revokeBySIDCalls != 2 {
		t.Errorf("RevokeByProviderAndSID called %d times, want 2", f.ss.revokeBySIDCalls)
	}
}
