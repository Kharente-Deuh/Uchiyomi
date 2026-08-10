// SPDX-License-Identifier: AGPL-3.0-or-later

package oidc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/federatedidentities"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
)

type FederatedRevalidator interface {
	RevalidateFederatedIdentity(
		context.Context,
		oidcproviders.OIDCProvider,
		federatedidentities.FederatedIdentity,
	) error
}

type RevalidationConfig struct {
	Interval time.Duration
}

func (cfg *RevalidationConfig) Validate() error {
	if cfg.Interval < time.Minute {
		return errors.New("interval must be at least 1 minute")
	}

	return nil
}

type RevalidationDeps struct {
	Revalidator                   FederatedRevalidator
	FederatedIdentitiesRepository federatedidentities.FederatedIdentitiesRepository
	OIDCProvidersRepository       oidcproviders.OIDCProvidersRepository
	Logger                        *slog.Logger
	Now                           func() time.Time
}

func (deps *RevalidationDeps) Validate() error {
	if deps.Revalidator == nil {
		return errors.New("revalidator is required")
	}

	if deps.FederatedIdentitiesRepository == nil {
		return errors.New("federatedIdentitiesRepository is required")
	}

	if deps.OIDCProvidersRepository == nil {
		return errors.New("oidcProvidersRepository is required")
	}

	if deps.Logger == nil {
		return errors.New("logger is required")
	}

	return nil
}

type RevalidationApp struct {
	deps RevalidationDeps
	cfg  RevalidationConfig
}

func NewRevalidationApp(cfg RevalidationConfig, deps RevalidationDeps) (*RevalidationApp, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	if deps.Now == nil {
		deps.Now = time.Now
	}

	deps.Logger = deps.Logger.With("component", "oidc.revalidation")

	return &RevalidationApp{deps: deps, cfg: cfg}, nil
}

func (a *RevalidationApp) Run(ctx context.Context) error {
	//nolint:wrapcheck
	return utils.Loop(ctx, utils.LoopOpts{
		Interval: a.cfg.Interval,
		Fn:       a.revalidate,
	})
}

func (a *RevalidationApp) revalidate(ctx context.Context) error {
	before := a.deps.Now().Add(-a.cfg.Interval)

	identities, err := a.deps.FederatedIdentitiesRepository.ListDueForRevalidation(ctx, before)
	if err != nil {
		a.deps.Logger.ErrorContext(ctx, "failed to list federated identities for revalidation", "err", err)

		return nil
	}

	for _, fi := range identities {
		if err := a.revalidateOne(ctx, fi); err != nil {
			a.deps.Logger.ErrorContext(ctx, "federated identity revalidation failed",
				"err", err, "federatedIdentityID", fi.ID, "providerID", fi.ProviderID)
		}
	}

	return nil
}

func (a *RevalidationApp) revalidateOne(ctx context.Context, fi federatedidentities.FederatedIdentity) error {
	provider, err := a.deps.OIDCProvidersRepository.GetByID(ctx, fi.ProviderID)
	if err != nil {
		return fmt.Errorf("a.deps.OIDCProvidersRepository.GetByID: %w", err)
	}

	if err := a.deps.Revalidator.RevalidateFederatedIdentity(ctx, *provider, fi); err != nil {
		return fmt.Errorf("a.deps.Revalidator.RevalidateFederatedIdentity: %w", err)
	}

	return nil
}
