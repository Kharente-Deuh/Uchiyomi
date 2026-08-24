// SPDX-License-Identifier: AGPL-3.0-or-later

package oidc_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/federatedidentities"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidc"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
)

type stubRevalidator struct {
	done  chan struct{}
	errs  []error
	calls int
}

func (s *stubRevalidator) RevalidateFederatedIdentity(
	context.Context,
	oidcproviders.OIDCProvider,
	federatedidentities.FederatedIdentity,
) error {
	var err error

	if s.calls < len(s.errs) {
		err = s.errs[s.calls]
	}

	s.calls++

	if s.done != nil && s.calls == len(s.errs) {
		close(s.done)
	}

	return err
}

type stubFederatedIdentitiesRepo struct {
	identities []federatedidentities.FederatedIdentity
}

func (r *stubFederatedIdentitiesRepo) Create(
	context.Context,
	federatedidentities.CreateFederatedIdentityOpts,
) (*federatedidentities.FederatedIdentity, error) {
	panic("Create is not used by the revalidation app")
}

func (r *stubFederatedIdentitiesRepo) Get(
	context.Context,
	federatedidentities.GetFederatedIdentityOpts,
) (*federatedidentities.FederatedIdentity, error) {
	panic("Get is not used by the revalidation app")
}

func (r *stubFederatedIdentitiesRepo) Update(context.Context, federatedidentities.UpdateFederatedIdentityOpts) error {
	panic("Update is not used by the revalidation app")
}

func (r *stubFederatedIdentitiesRepo) ListDueForRevalidation(
	context.Context,
	time.Time,
) ([]federatedidentities.FederatedIdentity, error) {
	return r.identities, nil
}

type stubOIDCProvidersRepo struct {
	providers map[uuid.UUID]*oidcproviders.OIDCProvider
}

func (r *stubOIDCProvidersRepo) GetByID(_ context.Context, id uuid.UUID) (*oidcproviders.OIDCProvider, error) {
	p, ok := r.providers[id]
	if !ok {
		return nil, errors.New("not found")
	}

	return p, nil
}

func (r *stubOIDCProvidersRepo) GetByIssuerURL(context.Context, string) (*oidcproviders.OIDCProvider, error) {
	panic("GetByIssuerURL is not used by the revalidation app")
}

func (r *stubOIDCProvidersRepo) GetBySlug(context.Context, string) (*oidcproviders.OIDCProvider, error) {
	panic("GetBySlug is not used by the revalidation app")
}

func (r *stubOIDCProvidersRepo) Create(
	context.Context,
	oidcproviders.CreateOIDCProviderOpts,
) (*oidcproviders.OIDCProvider, error) {
	panic("Create is not used by the revalidation app")
}

func (r *stubOIDCProvidersRepo) Update(
	context.Context,
	uuid.UUID,
	oidcproviders.UpdateOIDCProviderOpts,
) (*oidcproviders.OIDCProvider, error) {
	panic("Update is not used by the revalidation app")
}

func (r *stubOIDCProvidersRepo) DeleteByID(context.Context, uuid.UUID) error {
	panic("DeleteByID is not used by the revalidation app")
}

func (r *stubOIDCProvidersRepo) GetAll(context.Context) ([]oidcproviders.LightOIDCProvider, error) {
	panic("GetAll is not used by the revalidation app")
}

func (r *stubOIDCProvidersRepo) GetUsers(context.Context, uuid.UUID) ([]oidcproviders.OIDCProviderUser, error) {
	panic("GetUsers is not used by the revalidation app")
}

func TestRevalidationAppContinuesAfterOneProviderFails(t *testing.T) {
	t.Parallel()

	providerA := uuid.New()
	providerB := uuid.New()
	revalidated := make(chan struct{})
	revalidator := &stubRevalidator{
		errs: []error{
			errors.New("provider A is down"),
			nil,
		},
		done: revalidated,
	}

	app, err := oidc.NewRevalidationApp(
		oidc.RevalidationConfig{Interval: time.Hour},
		oidc.RevalidationDeps{
			Logger:      slog.New(slog.DiscardHandler),
			Revalidator: revalidator,
			FederatedIdentitiesRepository: &stubFederatedIdentitiesRepo{identities: []federatedidentities.FederatedIdentity{
				{ID: uuid.New(), ProviderID: providerA, UserID: uuid.New()},
				{ID: uuid.New(), ProviderID: providerB, UserID: uuid.New()},
			}},
			OIDCProvidersRepository: &stubOIDCProvidersRepo{providers: map[uuid.UUID]*oidcproviders.OIDCProvider{
				providerA: {ID: providerA},
				providerB: {ID: providerB},
			}},
			Now: func() time.Time { return time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC) },
		},
	)
	if err != nil {
		t.Fatalf("NewRevalidationApp: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runDone := make(chan error, 1)

	go func() {
		runDone <- app.Run(ctx)
	}()

	<-revalidated
	cancel()

	if err := <-runDone; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run = %v, want context.Canceled", err)
	}

	if revalidator.calls != 2 {
		t.Errorf("revalidator calls = %d, want 2", revalidator.calls)
	}
}
