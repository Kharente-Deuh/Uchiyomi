// SPDX-License-Identifier: AGPL-3.0-or-later

package auth_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/federatedidentities"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidc"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
)

const (
	testOIDCCode     = "the-code"
	testRedirectPath = "/library"
	roleClaimKey     = "role"
	staffRole        = "staff"
	adminRole        = "admin"
	guestRole        = "guest"
)

type fakeOIDCProvidersRepo struct {
	err      error
	provider *oidcproviders.OIDCProvider
	gotID    uuid.UUID
	calls    int
}

func (f *fakeOIDCProvidersRepo) GetByID(_ context.Context, id uuid.UUID) (*oidcproviders.OIDCProvider, error) {
	f.calls++
	f.gotID = id

	if f.err != nil {
		return nil, f.err
	}

	return f.provider, nil
}

func (f *fakeOIDCProvidersRepo) GetByIssuerURL(context.Context, string) (*oidcproviders.OIDCProvider, error) {
	panic("GetByIssuerURL is not used by the auth service")
}

func (f *fakeOIDCProvidersRepo) Create(
	context.Context,
	oidcproviders.CreateOIDCProviderOpts,
) (*oidcproviders.OIDCProvider, error) {
	panic("Create is not used by the auth service")
}

func (f *fakeOIDCProvidersRepo) Update(
	context.Context,
	uuid.UUID,
	oidcproviders.UpdateOIDCProviderOpts,
) (*oidcproviders.OIDCProvider, error) {
	panic("Update is not used by the auth service")
}

func (f *fakeOIDCProvidersRepo) DeleteByID(context.Context, uuid.UUID) error {
	panic("DeleteByID is not used by the auth service")
}

func (f *fakeOIDCProvidersRepo) GetAll(context.Context) ([]oidcproviders.LightOIDCProvider, error) {
	panic("GetAll is not used by the auth service")
}

func (f *fakeOIDCProvidersRepo) GetUsers(context.Context, uuid.UUID) ([]oidcproviders.OIDCProviderUser, error) {
	panic("GetUsers is not used by the auth service")
}

type fakeFederatedIdentitiesRepo struct {
	getErr        error
	createErr     error
	updateErr     error
	fi            *federatedidentities.FederatedIdentity
	created       *federatedidentities.FederatedIdentity
	gotUpdateOpts federatedidentities.UpdateFederatedIdentityOpts
	gotGetOpts    federatedidentities.GetFederatedIdentityOpts
	gotCreateOpts federatedidentities.CreateFederatedIdentityOpts
	getCalls      int
	createCalls   int
	updateCalls   int
}

func (f *fakeFederatedIdentitiesRepo) Get(
	_ context.Context,
	opts federatedidentities.GetFederatedIdentityOpts,
) (*federatedidentities.FederatedIdentity, error) {
	f.getCalls++
	f.gotGetOpts = opts

	if f.getErr != nil {
		return nil, f.getErr
	}

	return f.fi, nil
}

func (f *fakeFederatedIdentitiesRepo) Create(
	_ context.Context,
	opts federatedidentities.CreateFederatedIdentityOpts,
) (*federatedidentities.FederatedIdentity, error) {
	f.createCalls++
	f.gotCreateOpts = opts

	if f.createErr != nil {
		return nil, f.createErr
	}

	if f.created != nil {
		return f.created, nil
	}

	return &federatedidentities.FederatedIdentity{
		ID:         uuid.New(),
		Subject:    opts.Subject,
		ProviderID: opts.ProviderID,
		UserID:     opts.UserID,
		Claims:     opts.Claims,
	}, nil
}

func (f *fakeFederatedIdentitiesRepo) Update(
	_ context.Context,
	opts federatedidentities.UpdateFederatedIdentityOpts,
) error {
	f.updateCalls++
	f.gotUpdateOpts = opts

	return f.updateErr
}

type fakeOIDCClient struct {
	authCodeErr           error
	exchangeErr           error
	endSessionErr         error
	tokenSet              *oidcproviders.TokenSet
	authCodeURL           string
	endSessionURL         string
	gotCode               string
	gotVerifier           string
	gotNonce              string
	gotRedirectURI        string
	gotPostLogoutRedirect string
	gotAuthParams         oidcproviders.AuthCodeParams
	gotProvider           oidcproviders.OIDCProvider
	authCodeCalls         int
	exchangeCalls         int
	endSessionCalls       int
	endSessionSupported   bool
}

func (f *fakeOIDCClient) AuthCodeURL(
	_ context.Context,
	provider oidcproviders.OIDCProvider,
	params oidcproviders.AuthCodeParams,
) (string, error) {
	f.authCodeCalls++
	f.gotProvider = provider
	f.gotAuthParams = params

	if f.authCodeErr != nil {
		return "", f.authCodeErr
	}

	return f.authCodeURL, nil
}

func (f *fakeOIDCClient) Exchange(
	_ context.Context,
	provider oidcproviders.OIDCProvider,
	code, verifier, nonce, redirectURI string,
) (*oidcproviders.TokenSet, error) {
	f.exchangeCalls++
	f.gotProvider = provider
	f.gotCode = code
	f.gotVerifier = verifier
	f.gotNonce = nonce
	f.gotRedirectURI = redirectURI

	if f.exchangeErr != nil {
		return nil, f.exchangeErr
	}

	return f.tokenSet, nil
}

func (f *fakeOIDCClient) EndSessionURL(
	_ context.Context,
	provider oidcproviders.OIDCProvider,
	postLogoutRedirectURI string,
) (string, bool, error) {
	f.endSessionCalls++
	f.gotProvider = provider
	f.gotPostLogoutRedirect = postLogoutRedirectURI

	if f.endSessionErr != nil {
		return "", false, f.endSessionErr
	}

	return f.endSessionURL, f.endSessionSupported, nil
}

func TestStartOIDCLoginBuildsAuthCodeURLAndCookie(t *testing.T) {
	t.Parallel()

	f := newFakes()

	got, err := f.svc(t).StartOIDCLogin(context.Background(), auth.StartOIDCLoginOpts{
		ProviderID: f.opr.provider.ID,
		Redirect:   testRedirectPath,
	})
	if err != nil {
		t.Fatalf("StartOIDCLogin: %v", err)
	}

	if got.AuthCodeURL != f.oc.authCodeURL {
		t.Errorf("AuthCodeURL = %q, want %q", got.AuthCodeURL, f.oc.authCodeURL)
	}

	if got.StateCookieValue == "" {
		t.Fatal("StateCookieValue empty")
	}

	if !got.ExpiresAt.Equal(f.now.Add(f.cfg.StateCookieTTL)) {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, f.now.Add(f.cfg.StateCookieTTL))
	}

	if f.oc.gotAuthParams.RedirectURI != oidcRedirectURI {
		t.Errorf("RedirectURI = %q, want %q", f.oc.gotAuthParams.RedirectURI, oidcRedirectURI)
	}

	if f.oc.gotAuthParams.State == "" || f.oc.gotAuthParams.Nonce == "" || f.oc.gotAuthParams.Verifier == "" {
		t.Errorf("AuthCodeParams incomplets: %+v", f.oc.gotAuthParams)
	}
}

func TestStartOIDCLoginCookieRoundTripsThroughFinish(t *testing.T) {
	t.Parallel()

	f := newFakes()
	svc := f.svc(t)

	start, err := svc.StartOIDCLogin(context.Background(), auth.StartOIDCLoginOpts{
		ProviderID: f.opr.provider.ID,
		Redirect:   testRedirectPath,
	})
	if err != nil {
		t.Fatalf("StartOIDCLogin: %v", err)
	}

	got, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		Code:             testOIDCCode,
		State:            f.oc.gotAuthParams.State,
		StateCookieValue: start.StateCookieValue,
	})
	if err != nil {
		t.Fatalf("FinishOIDCLogin: %v", err)
	}

	if got.Redirect != testRedirectPath {
		t.Errorf("Redirect = %q, want /library", got.Redirect)
	}

	if f.oc.gotVerifier != f.oc.gotAuthParams.Verifier {
		t.Errorf("Exchange received verifier %q, want %q", f.oc.gotVerifier, f.oc.gotAuthParams.Verifier)
	}

	if f.oc.gotNonce != f.oc.gotAuthParams.Nonce {
		t.Errorf("Exchange received nonce %q, want %q", f.oc.gotNonce, f.oc.gotAuthParams.Nonce)
	}

	if f.oc.gotRedirectURI != oidcRedirectURI {
		t.Errorf("Exchange received redirectURI %q, want %q", f.oc.gotRedirectURI, oidcRedirectURI)
	}
}

func TestStartOIDCLoginUnknownProviderIsUnavailable(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.opr.err = domain.ErrNotFound

	_, err := f.svc(t).StartOIDCLogin(context.Background(), auth.StartOIDCLoginOpts{ProviderID: uuid.New()})
	if !errors.Is(err, auth.ErrOIDCUnavailable) {
		t.Errorf("err = %v, want ErrOIDCUnavailable", err)
	}
}

func TestStartOIDCLoginRepositoryFailureIsUnavailable(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.opr.err = errors.New("connection refused")

	_, err := f.svc(t).StartOIDCLogin(context.Background(), auth.StartOIDCLoginOpts{ProviderID: uuid.New()})
	if !errors.Is(err, auth.ErrOIDCUnavailable) {
		t.Errorf("err = %v, want ErrOIDCUnavailable", err)
	}
}

func TestStartOIDCLoginAuthCodeURLFailureIsUnavailable(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.oc.authCodeErr = errors.New("discovery unreachable")

	_, err := f.svc(t).StartOIDCLogin(context.Background(), auth.StartOIDCLoginOpts{ProviderID: f.opr.provider.ID})
	if !errors.Is(err, auth.ErrOIDCUnavailable) {
		t.Errorf("err = %v, want ErrOIDCUnavailable", err)
	}
}

func startCookie(t *testing.T, f *fakes, svc *auth.Service) (string, string) {
	t.Helper()

	start, err := svc.StartOIDCLogin(context.Background(), auth.StartOIDCLoginOpts{
		ProviderID: f.opr.provider.ID,
		Redirect:   testRedirectPath,
	})
	if err != nil {
		t.Fatalf("StartOIDCLogin: %v", err)
	}

	return start.StateCookieValue, f.oc.gotAuthParams.State
}

func TestFinishOIDCLoginRejectsMissingCookie(t *testing.T) {
	t.Parallel()

	f := newFakes()
	svc := f.svc(t)

	_, state := startCookie(t, f, svc)

	_, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		Code:  testOIDCCode,
		State: state,
	})
	if !errors.Is(err, auth.ErrOIDCState) {
		t.Errorf("err = %v, want ErrOIDCState", err)
	}
}

func TestFinishOIDCLoginRejectsCorruptCookie(t *testing.T) {
	t.Parallel()

	f := newFakes()
	svc := f.svc(t)

	_, state := startCookie(t, f, svc)

	_, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		Code:             testOIDCCode,
		State:            state,
		StateCookieValue: "not-a-valid-cookie",
	})
	if !errors.Is(err, auth.ErrOIDCState) {
		t.Errorf("err = %v, want ErrOIDCState", err)
	}
}

func TestFinishOIDCLoginRejectsStateMismatch(t *testing.T) {
	t.Parallel()

	f := newFakes()
	svc := f.svc(t)

	cookie, _ := startCookie(t, f, svc)

	_, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		Code:             testOIDCCode,
		State:            "not-the-real-state",
		StateCookieValue: cookie,
	})
	if !errors.Is(err, auth.ErrOIDCState) {
		t.Errorf("err = %v, want ErrOIDCState", err)
	}
}

func TestFinishOIDCLoginRejectsExpiredCookie(t *testing.T) {
	t.Parallel()

	f := newFakes()
	svc := f.svc(t)

	cookie, state := startCookie(t, f, svc)

	f.now = f.now.Add(f.cfg.StateCookieTTL + time.Minute)

	_, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		Code:             testOIDCCode,
		State:            state,
		StateCookieValue: cookie,
	})
	if !errors.Is(err, auth.ErrOIDCState) {
		t.Errorf("err = %v, want ErrOIDCState", err)
	}
}

func TestFinishOIDCLoginDenied(t *testing.T) {
	t.Parallel()

	f := newFakes()
	svc := f.svc(t)

	cookie, state := startCookie(t, f, svc)

	_, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		State:            state,
		StateCookieValue: cookie,
		ErrorParam:       "access_denied",
	})
	if !errors.Is(err, auth.ErrOIDCDenied) {
		t.Errorf("err = %v, want ErrOIDCDenied", err)
	}
}

func TestFinishOIDCLoginOtherErrorParamIsUnavailable(t *testing.T) {
	t.Parallel()

	f := newFakes()
	svc := f.svc(t)

	cookie, state := startCookie(t, f, svc)

	_, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		State:            state,
		StateCookieValue: cookie,
		ErrorParam:       "server_error",
	})
	if !errors.Is(err, auth.ErrOIDCUnavailable) {
		t.Errorf("err = %v, want ErrOIDCUnavailable", err)
	}
}

func TestFinishOIDCLoginProviderDeletedMidFlow(t *testing.T) {
	t.Parallel()

	f := newFakes()
	svc := f.svc(t)

	cookie, state := startCookie(t, f, svc)

	f.opr.err = domain.ErrNotFound

	_, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		Code:             testOIDCCode,
		State:            state,
		StateCookieValue: cookie,
	})
	if !errors.Is(err, auth.ErrOIDCUnavailable) {
		t.Errorf("err = %v, want ErrOIDCUnavailable", err)
	}
}

func TestFinishOIDCLoginNonceMismatchIsState(t *testing.T) {
	t.Parallel()

	f := newFakes()
	svc := f.svc(t)

	cookie, state := startCookie(t, f, svc)

	f.oc.exchangeErr = oidc.ErrNonceMismatch

	_, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		Code:             testOIDCCode,
		State:            state,
		StateCookieValue: cookie,
	})
	if !errors.Is(err, auth.ErrOIDCState) {
		t.Errorf("err = %v, want ErrOIDCState", err)
	}
}

func TestFinishOIDCLoginOtherExchangeErrorIsUnavailable(t *testing.T) {
	t.Parallel()

	f := newFakes()
	svc := f.svc(t)

	cookie, state := startCookie(t, f, svc)

	f.oc.exchangeErr = errors.New("token endpoint down")

	_, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		Code:             testOIDCCode,
		State:            state,
		StateCookieValue: cookie,
	})
	if !errors.Is(err, auth.ErrOIDCUnavailable) {
		t.Errorf("err = %v, want ErrOIDCUnavailable", err)
	}
}

func TestFinishOIDCLoginHappyPathIssuesOIDCSession(t *testing.T) {
	t.Parallel()

	f := newFakes()
	svc := f.svc(t)

	cookie, state := startCookie(t, f, svc)

	got, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		Code:             testOIDCCode,
		State:            state,
		StateCookieValue: cookie,
	})
	if err != nil {
		t.Fatalf("FinishOIDCLogin: %v", err)
	}

	if got.Session == nil {
		t.Fatal("Session nil")
	}

	if f.ss.gotOpts.AuthMethod != sessions.AuthMethodOIDC {
		t.Errorf("AuthMethod = %v, want AuthMethodOIDC", f.ss.gotOpts.AuthMethod)
	}

	if f.ss.gotOpts.ProviderID == nil || *f.ss.gotOpts.ProviderID != f.opr.provider.ID {
		t.Errorf("ProviderID = %v, want %v", f.ss.gotOpts.ProviderID, f.opr.provider.ID)
	}

	if f.ss.gotOpts.ProviderSID == nil || *f.ss.gotOpts.ProviderSID != f.oc.tokenSet.SID {
		t.Errorf("ProviderSID = %v, want %v", f.ss.gotOpts.ProviderSID, f.oc.tokenSet.SID)
	}

	if got.Redirect != testRedirectPath {
		t.Errorf("Redirect = %q, want /library", got.Redirect)
	}

	if f.ur.updateCalls != 0 {
		t.Errorf("UsersRepository.Update called %d times without RoleClaim configured, want 0", f.ur.updateCalls)
	}
}

func TestFinishOIDCLoginUnsafeRedirectFallsBackToRoot(t *testing.T) {
	t.Parallel()

	f := newFakes()
	svc := f.svc(t)

	start, err := svc.StartOIDCLogin(context.Background(), auth.StartOIDCLoginOpts{
		ProviderID: f.opr.provider.ID,
		Redirect:   "//evil.example.com",
	})
	if err != nil {
		t.Fatalf("StartOIDCLogin: %v", err)
	}

	got, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		Code:             testOIDCCode,
		State:            f.oc.gotAuthParams.State,
		StateCookieValue: start.StateCookieValue,
	})
	if err != nil {
		t.Fatalf("FinishOIDCLogin: %v", err)
	}

	if got.Redirect != "/" {
		t.Errorf("Redirect = %q, want /", got.Redirect)
	}
}

func TestFinishOIDCLoginNotAllowedByClaims(t *testing.T) {
	t.Parallel()

	f := newFakes()
	role := roleClaimKey
	f.opr.provider.RoleClaim = &role
	f.opr.provider.AllowedValues = []string{staffRole}
	f.oc.tokenSet.Claims[roleClaimKey] = guestRole

	svc := f.svc(t)

	cookie, state := startCookie(t, f, svc)

	_, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		Code:             testOIDCCode,
		State:            state,
		StateCookieValue: cookie,
	})
	if !errors.Is(err, auth.ErrOIDCNotAllowed) {
		t.Errorf("err = %v, want ErrOIDCNotAllowed", err)
	}
}

func TestMapClaims(t *testing.T) {
	t.Parallel()

	role := roleClaimKey

	tests := map[string]struct {
		claims       map[string]any
		wantUsername string
		provider     oidcproviders.OIDCProvider
		wantIsAdmin  bool
		wantAllowed  bool
	}{
		"RoleClaim nil => never admin, always allowed": {
			provider:     oidcproviders.OIDCProvider{UsernameClaim: usernameClaimKey},
			claims:       map[string]any{usernameClaimKey: userName},
			wantUsername: userName,
			wantIsAdmin:  false,
			wantAllowed:  true,
		},
		"AllowedValues empty => allowed without restriction": {
			provider:     oidcproviders.OIDCProvider{UsernameClaim: usernameClaimKey, RoleClaim: &role},
			claims:       map[string]any{usernameClaimKey: userName, roleClaimKey: guestRole},
			wantUsername: userName,
			wantIsAdmin:  false,
			wantAllowed:  true,
		},
		"AllowedValues non-empty and intersects => allowed": {
			provider: oidcproviders.OIDCProvider{
				UsernameClaim: usernameClaimKey,
				RoleClaim:     &role,
				AllowedValues: []string{staffRole, adminRole},
			},
			claims:       map[string]any{usernameClaimKey: userName, roleClaimKey: staffRole},
			wantUsername: userName,
			wantIsAdmin:  false,
			wantAllowed:  true,
		},
		"AllowedValues non-empty and disjoint => denied": {
			provider: oidcproviders.OIDCProvider{
				UsernameClaim: usernameClaimKey,
				RoleClaim:     &role,
				AllowedValues: []string{staffRole, adminRole},
			},
			claims:       map[string]any{usernameClaimKey: userName, roleClaimKey: guestRole},
			wantUsername: userName,
			wantIsAdmin:  false,
			wantAllowed:  false,
		},
		"AdminValues empty => never admin": {
			provider: oidcproviders.OIDCProvider{
				UsernameClaim: usernameClaimKey,
				RoleClaim:     &role,
			},
			claims:       map[string]any{usernameClaimKey: userName, roleClaimKey: adminRole},
			wantUsername: userName,
			wantIsAdmin:  false,
			wantAllowed:  true,
		},
		"AdminValues intersecte => admin": {
			provider: oidcproviders.OIDCProvider{
				UsernameClaim: usernameClaimKey,
				RoleClaim:     &role,
				AdminValues:   []string{adminRole},
			},
			claims:       map[string]any{usernameClaimKey: userName, roleClaimKey: adminRole},
			wantUsername: userName,
			wantIsAdmin:  true,
			wantAllowed:  true,
		},
		"AdminValues disjoint => not admin": {
			provider: oidcproviders.OIDCProvider{
				UsernameClaim: usernameClaimKey,
				RoleClaim:     &role,
				AdminValues:   []string{adminRole},
			},
			claims:       map[string]any{usernameClaimKey: userName, roleClaimKey: staffRole},
			wantUsername: userName,
			wantIsAdmin:  false,
			wantAllowed:  true,
		},
		"role en tableau de string": {
			provider: oidcproviders.OIDCProvider{
				UsernameClaim: usernameClaimKey,
				RoleClaim:     &role,
				AdminValues:   []string{adminRole},
			},
			claims:       map[string]any{usernameClaimKey: userName, roleClaimKey: []string{staffRole, adminRole}},
			wantUsername: userName,
			wantIsAdmin:  true,
			wantAllowed:  true,
		},
		"role en tableau any": {
			provider: oidcproviders.OIDCProvider{
				UsernameClaim: usernameClaimKey,
				RoleClaim:     &role,
				AdminValues:   []string{adminRole},
			},
			claims:       map[string]any{usernameClaimKey: userName, roleClaimKey: []any{staffRole, adminRole}},
			wantUsername: userName,
			wantIsAdmin:  true,
			wantAllowed:  true,
		},
		"UsernameClaim missing => not allowed": {
			provider:     oidcproviders.OIDCProvider{UsernameClaim: usernameClaimKey},
			claims:       map[string]any{},
			wantUsername: "",
			wantIsAdmin:  false,
			wantAllowed:  false,
		},
		"UsernameClaim wrong type => not allowed": {
			provider:     oidcproviders.OIDCProvider{UsernameClaim: usernameClaimKey},
			claims:       map[string]any{usernameClaimKey: 42},
			wantUsername: "",
			wantIsAdmin:  false,
			wantAllowed:  false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newFakes()
			*f.opr.provider = tc.provider
			f.oc.tokenSet.Claims = tc.claims

			svc := f.svc(t)

			cookie, state := startCookie(t, f, svc)

			_, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
				Code:             testOIDCCode,
				State:            state,
				StateCookieValue: cookie,
			})

			if !tc.wantAllowed {
				if !errors.Is(err, auth.ErrOIDCNotAllowed) {
					t.Fatalf("err = %v, want ErrOIDCNotAllowed", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("FinishOIDCLogin: %v", err)
			}

			if f.ur.gotUpdate.IsAdmin != tc.wantIsAdmin {
				t.Errorf("IsAdmin = %v, want %v", f.ur.gotUpdate.IsAdmin, tc.wantIsAdmin)
			}

			if f.ur.gotName != tc.wantUsername {
				t.Errorf("username used = %q, want %q", f.ur.gotName, tc.wantUsername)
			}
		})
	}
}

func TestResolveIdentityReusesExistingFederatedIdentity(t *testing.T) {
	t.Parallel()

	f := newFakes()

	linkedUser := &users.User{ID: uuid.New(), Name: "bob"}
	f.ur.byID[linkedUser.ID] = linkedUser

	f.fir.getErr = nil
	f.fir.fi = &federatedidentities.FederatedIdentity{
		ID:         uuid.New(),
		Subject:    testSubject,
		ProviderID: f.opr.provider.ID,
		UserID:     linkedUser.ID,
	}

	svc := f.svc(t)

	cookie, state := startCookie(t, f, svc)

	if _, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		Code:             testOIDCCode,
		State:            state,
		StateCookieValue: cookie,
	}); err != nil {
		t.Fatalf("FinishOIDCLogin: %v", err)
	}

	wantGetOpts := federatedidentities.GetFederatedIdentityOpts{
		Subject:    testSubject,
		ProviderID: f.opr.provider.ID,
	}
	if f.fir.gotGetOpts != wantGetOpts {
		t.Errorf("GetFederatedIdentityOpts = %+v, want %+v", f.fir.gotGetOpts, wantGetOpts)
	}

	if f.ur.gotID != linkedUser.ID {
		t.Errorf("UsersRepository.GetByID received %v, want %v (fi.UserID)", f.ur.gotID, linkedUser.ID)
	}

	if f.ss.gotOpts.UserID != linkedUser.ID {
		t.Errorf("UserID de la session = %v, want %v", f.ss.gotOpts.UserID, linkedUser.ID)
	}

	if f.ur.byIDCalls != 1 {
		t.Errorf("UsersRepository.GetByID called %d times, want 1", f.ur.byIDCalls)
	}

	if f.fir.updateCalls != 1 {
		t.Errorf("FederatedIdentitiesRepository.Update called %d times, want 1", f.fir.updateCalls)
	}

	if f.fir.createCalls != 0 {
		t.Errorf("FederatedIdentitiesRepository.Create called %d times, want 0", f.fir.createCalls)
	}

	if f.tr.calls != 0 {
		t.Errorf("WithinTx called %d times, want 0 (no provisioning)", f.tr.calls)
	}
}

func TestResolveIdentityLinksLocalUsernameMatch(t *testing.T) {
	t.Parallel()

	f := newFakes()

	svc := f.svc(t)

	cookie, state := startCookie(t, f, svc)

	if _, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		Code:             testOIDCCode,
		State:            state,
		StateCookieValue: cookie,
	}); err != nil {
		t.Fatalf("FinishOIDCLogin: %v", err)
	}

	if f.ss.gotOpts.UserID != f.ur.user.ID {
		t.Errorf("UserID = %v, want %v", f.ss.gotOpts.UserID, f.ur.user.ID)
	}

	if f.fir.createCalls != 1 {
		t.Errorf("FederatedIdentitiesRepository.Create called %d times, want 1", f.fir.createCalls)
	}

	if f.fir.gotCreateOpts.UserID != f.ur.user.ID || f.fir.gotCreateOpts.Subject != f.oc.tokenSet.Subject {
		t.Errorf("CreateFederatedIdentityOpts = %+v", f.fir.gotCreateOpts)
	}

	if f.tr.calls != 0 {
		t.Errorf("WithinTx called %d times, want 0 (no provisioning)", f.tr.calls)
	}
}

func TestResolveIdentityAutoProvisionsNewUser(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.opr.provider.AutoProvision = true
	f.ur.byNameErr = domain.ErrNotFound

	svc := f.svc(t)

	cookie, state := startCookie(t, f, svc)

	got, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		Code:             testOIDCCode,
		State:            state,
		StateCookieValue: cookie,
	})
	if err != nil {
		t.Fatalf("FinishOIDCLogin: %v", err)
	}

	if got.Session == nil {
		t.Fatal("Session nil")
	}

	if f.tr.calls != 1 {
		t.Errorf("WithinTx called %d times, want 1", f.tr.calls)
	}

	if !f.ur.inTx {
		t.Error("UsersRepository.Create did not receive transactional ctx")
	}

	if f.fir.createCalls != 1 {
		t.Errorf("FederatedIdentitiesRepository.Create called %d times, want 1", f.fir.createCalls)
	}
}

func TestResolveIdentityNoMatchWithoutAutoProvisionRefuses(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.opr.provider.AutoProvision = false
	f.ur.byNameErr = domain.ErrNotFound

	svc := f.svc(t)

	cookie, state := startCookie(t, f, svc)

	_, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
		Code:             testOIDCCode,
		State:            state,
		StateCookieValue: cookie,
	})
	if !errors.Is(err, auth.ErrOIDCNoAccount) {
		t.Errorf("err = %v, want ErrOIDCNoAccount", err)
	}

	if f.tr.calls != 0 {
		t.Errorf("WithinTx called %d times, want 0", f.tr.calls)
	}
}

func TestFinishOIDCLoginRecomputesIsAdminOnEveryBranch(t *testing.T) {
	t.Parallel()

	role := roleClaimKey

	tests := map[string]func(*fakes){
		"existing identity": func(f *fakes) {
			f.fir.fi = &federatedidentities.FederatedIdentity{
				ID: uuid.New(), Subject: testSubject, ProviderID: f.opr.provider.ID, UserID: f.ur.user.ID,
			}
			f.fir.getErr = nil
		},
		"lien par username": func(*fakes) {},
		"auto-provisioning": func(f *fakes) {
			f.opr.provider.AutoProvision = true
			f.ur.byNameErr = domain.ErrNotFound
		},
	}

	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newFakes()
			f.opr.provider.RoleClaim = &role
			f.opr.provider.AdminValues = []string{adminRole}
			f.oc.tokenSet.Claims[roleClaimKey] = adminRole
			setup(f)

			svc := f.svc(t)

			cookie, state := startCookie(t, f, svc)

			if _, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
				Code:             testOIDCCode,
				State:            state,
				StateCookieValue: cookie,
			}); err != nil {
				t.Fatalf("FinishOIDCLogin: %v", err)
			}

			if f.ur.updateCalls != 1 {
				t.Fatalf("UsersRepository.Update called %d times, want 1", f.ur.updateCalls)
			}

			if !f.ur.gotUpdate.IsAdmin {
				t.Error("IsAdmin not recalculated to true")
			}
		})
	}
}

func TestFinishOIDCLoginNeverDemotesWhenTheProviderMapsNoAdminRole(t *testing.T) {
	t.Parallel()

	role := roleClaimKey

	tests := map[string]func(*fakes){
		"no RoleClaim":  func(*fakes) {},
		"AdminValues empty": func(f *fakes) { f.opr.provider.RoleClaim = &role },
	}

	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newFakes()
			f.ur.user.IsAdmin = true
			f.ur.adminCount = 12
			setup(f)

			svc := f.svc(t)

			cookie, state := startCookie(t, f, svc)

			got, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
				Code:             testOIDCCode,
				State:            state,
				StateCookieValue: cookie,
			})
			if err != nil {
				t.Fatalf("FinishOIDCLogin: %v", err)
			}

			if got.Session == nil {
				t.Fatal("Session nil")
			}

			if f.ur.updateCalls != 0 {
				t.Errorf("UsersRepository.Update called %d times, want 0 (%+v)", f.ur.updateCalls, f.ur.gotUpdate)
			}
		})
	}
}

func TestFinishOIDCLoginKeepsTheLastAdminWhenTheProviderDemotesThem(t *testing.T) {
	t.Parallel()

	role := roleClaimKey

	tests := map[string]struct {
		adminCount     int
		wantUpdate     bool
		wantWarnLogged bool
	}{
		"dernier admin":  {adminCount: 1, wantUpdate: false, wantWarnLogged: true},
		"un autre admin": {adminCount: 2, wantUpdate: true, wantWarnLogged: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newFakes()
			f.opr.provider.RoleClaim = &role
			f.opr.provider.AdminValues = []string{adminRole}
			f.oc.tokenSet.Claims[roleClaimKey] = guestRole
			f.ur.user.IsAdmin = true
			f.ur.adminCount = tc.adminCount

			svc := f.svc(t)

			cookie, state := startCookie(t, f, svc)

			got, err := svc.FinishOIDCLogin(context.Background(), auth.FinishOIDCLoginOpts{
				Code:             testOIDCCode,
				State:            state,
				StateCookieValue: cookie,
			})
			if err != nil {
				t.Fatalf("FinishOIDCLogin: %v", err)
			}

			if got.Session == nil {
				t.Fatal("Session nil")
			}

			if tc.wantUpdate {
				if f.ur.updateCalls != 1 {
					t.Fatalf("UsersRepository.Update called %d times, want 1", f.ur.updateCalls)
				}

				if f.ur.gotUpdate.IsAdmin {
					t.Error("IsAdmin = true, want false: demotion must apply")
				}
			} else if f.ur.updateCalls != 0 {
				t.Errorf("last admin was demoted (%+v)", f.ur.gotUpdate)
			}

			if got := strings.Contains(f.logs.String(), "kept the last admin"); got != tc.wantWarnLogged {
				t.Errorf("warning log present = %v, want %v (logs: %s)", got, tc.wantWarnLogged, f.logs)
			}
		})
	}
}
