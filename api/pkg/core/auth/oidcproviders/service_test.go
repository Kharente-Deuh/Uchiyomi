// SPDX-License-Identifier: AGPL-3.0-or-later

package oidcproviders_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
)

const (
	testIssuerURL     = "https://sso.example.com"
	testRedirect      = "https://manga.example.com/api/auth/oidc/callback"
	testSecret        = "s3cr3t"
	testDisplayName   = "Keycloak"
	testClientID      = "uchiyomi"
	testUsernameClaim = "preferred_username"
	testScope         = "openid"
)

type fakeRepository struct {
	created  *oidcproviders.CreateOIDCProviderOpts
	updated  *oidcproviders.UpdateOIDCProviderOpts
	provider *oidcproviders.OIDCProvider
	all      []oidcproviders.LightOIDCProvider
	users    []oidcproviders.OIDCProviderUser
	err      error
	deleted  []uuid.UUID
}

func (f *fakeRepository) GetByID(context.Context, uuid.UUID) (*oidcproviders.OIDCProvider, error) {
	return f.provider, f.err
}

func (f *fakeRepository) GetByIssuerURL(context.Context, string) (*oidcproviders.OIDCProvider, error) {
	return f.provider, f.err
}

//nolint:lll
func (f *fakeRepository) Create(_ context.Context, opts oidcproviders.CreateOIDCProviderOpts) (*oidcproviders.OIDCProvider, error) {
	f.created = &opts

	return f.provider, f.err
}

//nolint:lll
func (f *fakeRepository) Update(_ context.Context, _ uuid.UUID, opts oidcproviders.UpdateOIDCProviderOpts) (*oidcproviders.OIDCProvider, error) {
	f.updated = &opts

	return f.provider, f.err
}

func (f *fakeRepository) DeleteByID(_ context.Context, id uuid.UUID) error {
	f.deleted = append(f.deleted, id)

	return f.err
}

func (f *fakeRepository) GetAll(context.Context) ([]oidcproviders.LightOIDCProvider, error) {
	return f.all, f.err
}

func (f *fakeRepository) GetUsers(context.Context, uuid.UUID) ([]oidcproviders.OIDCProviderUser, error) {
	return f.users, f.err
}

type fakeCipher struct {
	err   error
	seals int
}

func (f *fakeCipher) Seal(plaintext []byte) ([]byte, error) {
	f.seals++

	if f.err != nil {
		return nil, f.err
	}

	return append([]byte("sealed:"), plaintext...), nil
}

type fakeDiscoverer struct {
	discovery *oidcproviders.Discovery
	err       error
	issuers   []string
}

func (f *fakeDiscoverer) Discover(_ context.Context, issuerURL string) (*oidcproviders.Discovery, error) {
	f.issuers = append(f.issuers, issuerURL)

	if f.err != nil {
		return nil, f.err
	}

	if f.discovery != nil {
		return f.discovery, nil
	}

	return &oidcproviders.Discovery{
		Issuer:                issuerURL,
		AuthorizationEndpoint: issuerURL + "/auth",
		TokenEndpoint:         issuerURL + "/token",
	}, nil
}

func newService(
	t *testing.T,
	repo *fakeRepository,
	cipher *fakeCipher,
	disco *fakeDiscoverer,
) *oidcproviders.Service {
	t.Helper()

	s, err := oidcproviders.NewService(
		oidcproviders.ServiceConfig{RedirectURI: testRedirect},
		oidcproviders.ServiceDeps{Repository: repo, Cipher: cipher, Discoverer: disco},
	)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	return s
}

func createOpts() oidcproviders.CreateOpts {
	return oidcproviders.CreateOpts{
		DisplayName:   testDisplayName,
		IssuerURL:     testIssuerURL,
		ClientID:      testClientID,
		ClientSecret:  testSecret,
		Scopes:        []string{testScope, "profile"},
		UsernameClaim: testUsernameClaim,
	}
}

func TestCreateEncryptsTheClientSecret(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{provider: &oidcproviders.OIDCProvider{ID: uuid.New()}}
	s := newService(t, repo, &fakeCipher{}, &fakeDiscoverer{})

	if _, err := s.Create(context.Background(), createOpts()); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if repo.created == nil {
		t.Fatal("the repository was never called")
	}

	if bytes.Contains(repo.created.ClientSecretEnc, []byte(testSecret)) &&
		!bytes.HasPrefix(repo.created.ClientSecretEnc, []byte("sealed:")) {
		t.Errorf("the secret reached the repository unencrypted: %q", repo.created.ClientSecretEnc)
	}

	if !bytes.Equal(repo.created.ClientSecretEnc, []byte("sealed:"+testSecret)) {
		t.Errorf("ClientSecretEnc = %q, want the sealed secret", repo.created.ClientSecretEnc)
	}
}

func TestCreateNeverReturnsTheSecret(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{provider: &oidcproviders.OIDCProvider{
		ID:              uuid.New(),
		ClientSecretEnc: []byte("sealed:s3cr3t"),
	}}
	s := newService(t, repo, &fakeCipher{}, &fakeDiscoverer{})

	got, err := s.Create(context.Background(), createOpts())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if got.ClientSecretEnc != nil {
		t.Errorf("ClientSecretEnc = %q, want nil", got.ClientSecretEnc)
	}
}

func TestCreateRejectsAnIssuerThatFailsDiscovery(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}
	s := newService(t, repo, &fakeCipher{}, &fakeDiscoverer{err: errors.New("connection refused")})

	_, err := s.Create(context.Background(), createOpts())
	if !errors.Is(err, oidcproviders.ErrUnreachableIssuer) {
		t.Fatalf("Create = %v, want ErrUnreachableIssuer", err)
	}

	if repo.created != nil {
		t.Error("the provider was written despite the discovery failure")
	}
}

func TestCreateRejectsAnIncompleteDiscoveryDocument(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}
	s := newService(t, repo, &fakeCipher{}, &fakeDiscoverer{err: oidcproviders.ErrIncompleteDiscovery})

	_, err := s.Create(context.Background(), createOpts())
	if !errors.Is(err, oidcproviders.ErrIncompleteIssuer) {
		t.Fatalf("Create = %v, want ErrIncompleteIssuer", err)
	}

	if errors.Is(err, oidcproviders.ErrUnreachableIssuer) {
		t.Error("an incomplete document must not also read as ErrUnreachableIssuer")
	}

	if repo.created != nil {
		t.Error("the provider was written despite the incomplete discovery document")
	}
}

func TestCreatePropagatesADuplicateIssuer(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{err: domain.ErrAlreadyExists}
	s := newService(t, repo, &fakeCipher{}, &fakeDiscoverer{})

	if _, err := s.Create(context.Background(), createOpts()); !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("Create = %v, want domain.ErrAlreadyExists", err)
	}
}

func TestUpdateNeverSealsASecret(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{provider: &oidcproviders.OIDCProvider{ID: uuid.New()}}
	cipher := &fakeCipher{}
	s := newService(t, repo, cipher, &fakeDiscoverer{})

	_, err := s.Update(context.Background(), uuid.New(), oidcproviders.UpdateOpts{
		DisplayName:   testDisplayName,
		IssuerURL:     testIssuerURL,
		ClientID:      testClientID,
		Scopes:        []string{testScope},
		UsernameClaim: testUsernameClaim,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if repo.updated == nil {
		t.Fatal("the repository was never called")
	}

	if cipher.seals != 0 {
		t.Errorf("the cipher was called %d times, want 0: an update must not rotate the secret", cipher.seals)
	}
}

func TestUpdateRejectsAnIssuerThatFailsDiscovery(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}
	s := newService(t, repo, &fakeCipher{}, &fakeDiscoverer{err: errors.New("no such host")})

	_, err := s.Update(context.Background(), uuid.New(), oidcproviders.UpdateOpts{IssuerURL: testIssuerURL})
	if !errors.Is(err, oidcproviders.ErrUnreachableIssuer) {
		t.Fatalf("Update = %v, want ErrUnreachableIssuer", err)
	}

	if repo.updated != nil {
		t.Error("the provider was written despite the discovery failure")
	}
}

func TestListReturnsTheLightShape(t *testing.T) {
	t.Parallel()

	id1, id2 := uuid.New(), uuid.New()
	repo := &fakeRepository{all: []oidcproviders.LightOIDCProvider{
		{ID: id1, DisplayName: "Authentik"},
		{ID: id2, DisplayName: testDisplayName},
	}}
	s := newService(t, repo, &fakeCipher{}, &fakeDiscoverer{})

	got, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(got) != 2 || got[0].ID != id1 || got[1].ID != id2 {
		t.Errorf("List() = %+v, want the repository's light providers in order", got)
	}
}

func TestGetByIDNeverReturnsTheSecret(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{provider: &oidcproviders.OIDCProvider{
		ID:              uuid.New(),
		DisplayName:     testDisplayName,
		ClientSecretEnc: []byte("sealed:s3cr3t"),
	}}
	s := newService(t, repo, &fakeCipher{}, &fakeDiscoverer{})

	got, err := s.GetByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.Provider.ClientSecretEnc != nil {
		t.Errorf("ClientSecretEnc = %q, want nil", got.Provider.ClientSecretEnc)
	}

	if got.Provider.DisplayName != testDisplayName {
		t.Errorf("DisplayName = %q, want the rest of the provider to come through", got.Provider.DisplayName)
	}
}

func TestGetByIDReturnsTheLinkedUsers(t *testing.T) {
	t.Parallel()

	linkedAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	userID := uuid.New()
	repo := &fakeRepository{
		provider: &oidcproviders.OIDCProvider{ID: uuid.New(), DisplayName: testDisplayName},
		users: []oidcproviders.OIDCProviderUser{
			{ID: userID, Username: "alice", LinkedAt: linkedAt, IsAdmin: true},
		},
	}
	s := newService(t, repo, &fakeCipher{}, &fakeDiscoverer{})

	got, err := s.GetByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if len(got.Users) != 1 {
		t.Fatalf("Users = %+v, want the repository's users", got.Users)
	}

	if got.Users[0].ID != userID || got.Users[0].Username != "alice" ||
		!got.Users[0].IsAdmin || !got.Users[0].LinkedAt.Equal(linkedAt) {
		t.Errorf("Users[0] = %+v", got.Users[0])
	}
}

func TestGetByIDPropagatesNotFound(t *testing.T) {
	t.Parallel()

	s := newService(t, &fakeRepository{err: domain.ErrNotFound}, &fakeCipher{}, &fakeDiscoverer{})

	if _, err := s.GetByID(context.Background(), uuid.New()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID = %v, want domain.ErrNotFound", err)
	}
}

func TestDeleteForwardsTheID(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}
	s := newService(t, repo, &fakeCipher{}, &fakeDiscoverer{})

	id := uuid.New()
	if err := s.Delete(context.Background(), id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if len(repo.deleted) != 1 || repo.deleted[0] != id {
		t.Errorf("deleted = %v, want [%s]", repo.deleted, id)
	}
}

func TestDeletePropagatesNotFound(t *testing.T) {
	t.Parallel()

	s := newService(t, &fakeRepository{err: domain.ErrNotFound}, &fakeCipher{}, &fakeDiscoverer{})

	if err := s.Delete(context.Background(), uuid.New()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete = %v, want domain.ErrNotFound", err)
	}
}

func TestProbeReportsTheEndpointsAndTheRedirectURI(t *testing.T) {
	t.Parallel()

	disco := &fakeDiscoverer{discovery: &oidcproviders.Discovery{
		Issuer:                testIssuerURL,
		AuthorizationEndpoint: testIssuerURL + "/auth",
		TokenEndpoint:         testIssuerURL + "/token",
		UserInfoEndpoint:      testIssuerURL + "/userinfo",
		EndSessionEndpoint:    testIssuerURL + "/logout",
	}}
	s := newService(t, &fakeRepository{}, &fakeCipher{}, disco)

	got, err := s.Probe(context.Background(), testIssuerURL)
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}

	if !got.SupportsRPInitiatedLogout {
		t.Error("SupportsRPInitiatedLogout = false, want true")
	}

	if got.RedirectURI != testRedirect {
		t.Errorf("RedirectURI = %q, want %q", got.RedirectURI, testRedirect)
	}

	if got.TokenEndpoint != testIssuerURL+"/token" {
		t.Errorf("TokenEndpoint = %q", got.TokenEndpoint)
	}
}

func TestProbeReportsAProviderWithoutEndSession(t *testing.T) {
	t.Parallel()

	s := newService(t, &fakeRepository{}, &fakeCipher{}, &fakeDiscoverer{})

	got, err := s.Probe(context.Background(), testIssuerURL)
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}

	if got.SupportsRPInitiatedLogout {
		t.Error("SupportsRPInitiatedLogout = true, want false")
	}
}

func TestProbeRejectsAnUnreachableIssuer(t *testing.T) {
	t.Parallel()

	s := newService(t, &fakeRepository{}, &fakeCipher{}, &fakeDiscoverer{err: errors.New("timeout")})

	if _, err := s.Probe(context.Background(), testIssuerURL); !errors.Is(err, oidcproviders.ErrUnreachableIssuer) {
		t.Errorf("Probe = %v, want ErrUnreachableIssuer", err)
	}
}

func TestProbeRejectsAnIncompleteDiscoveryDocument(t *testing.T) {
	t.Parallel()

	s := newService(t, &fakeRepository{}, &fakeCipher{}, &fakeDiscoverer{err: oidcproviders.ErrIncompleteDiscovery})

	if _, err := s.Probe(context.Background(), testIssuerURL); !errors.Is(err, oidcproviders.ErrIncompleteIssuer) {
		t.Errorf("Probe = %v, want ErrIncompleteIssuer", err)
	}
}

func TestNewServiceValidates(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		deps    oidcproviders.ServiceDeps
		cfg     oidcproviders.ServiceConfig
		wantErr string
	}{
		"no redirect uri": {
			cfg: oidcproviders.ServiceConfig{},
			deps: oidcproviders.ServiceDeps{
				Repository: &fakeRepository{}, Cipher: &fakeCipher{}, Discoverer: &fakeDiscoverer{},
			},
			wantErr: "cfg.Validate: redirectURI is required",
		},
		"no repository": {
			cfg:     oidcproviders.ServiceConfig{RedirectURI: testRedirect},
			deps:    oidcproviders.ServiceDeps{Cipher: &fakeCipher{}, Discoverer: &fakeDiscoverer{}},
			wantErr: "deps.Validate: repository is required",
		},
		"no cipher": {
			cfg:     oidcproviders.ServiceConfig{RedirectURI: testRedirect},
			deps:    oidcproviders.ServiceDeps{Repository: &fakeRepository{}, Discoverer: &fakeDiscoverer{}},
			wantErr: "deps.Validate: cipher is required",
		},
		"no discoverer": {
			cfg:     oidcproviders.ServiceConfig{RedirectURI: testRedirect},
			deps:    oidcproviders.ServiceDeps{Repository: &fakeRepository{}, Cipher: &fakeCipher{}},
			wantErr: "deps.Validate: discoverer is required",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := oidcproviders.NewService(tc.cfg, tc.deps)
			if err == nil || err.Error() != tc.wantErr {
				t.Errorf("NewService() = %v, want %q", err, tc.wantErr)
			}
		})
	}
}
