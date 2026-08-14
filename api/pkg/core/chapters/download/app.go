// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
	"golang.org/x/sync/errgroup"
)

type AppConfig struct {
	ScanInterval time.Duration
}

func (cfg *AppConfig) Validate() error {
	if cfg.ScanInterval < time.Minute {
		return errors.New("scanInterval must be at least 1 minute")
	}

	return nil
}

type WorkerRunner interface {
	Run(context.Context) error
}

type AppDeps struct {
	Worker          WorkerRunner
	ChaptersService chapters.ChaptersService
	Logger          *slog.Logger
}

func (deps *AppDeps) Validate() error {
	if deps.Worker == nil {
		return ErrWorkerRequired
	}

	if deps.ChaptersService == nil {
		return errors.New("chaptersService is required")
	}

	return nil
}

type App struct {
	deps AppDeps
	cfg  AppConfig
}

func NewApp(cfg AppConfig, deps AppDeps) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}

	deps.Logger = deps.Logger.With("component", "chapters.download.app")

	return &App{deps: deps, cfg: cfg}, nil
}

func (a *App) Run(ctx context.Context) error {
	if err := a.deps.ChaptersService.EnqueueResumable(ctx); err != nil {
		return fmt.Errorf("a.deps.ChaptersService.EnqueueResumable: %w", err)
	}

	errG, errCtx := errgroup.WithContext(ctx)

	errG.Go(func() error {
		//nolint:wrapcheck
		return a.deps.Worker.Run(errCtx)
	})

	errG.Go(func() error {
		//nolint:wrapcheck // Run delegates unchanged to utils.Loop, like sessions and oidc revalidation.
		return utils.Loop(errCtx, utils.LoopOpts{
			Interval: a.cfg.ScanInterval,
			Fn:       a.scanEarlyAccess,
		})
	})

	a.deps.Logger.Info("download app started")

	//nolint:wrapcheck
	return errG.Wait()
}

func (a *App) scanEarlyAccess(ctx context.Context) error {
	if err := a.deps.ChaptersService.ScanEarlyAccess(ctx); err != nil {
		a.deps.Logger.ErrorContext(ctx, "early-access scan failed", loggingErr(err))
	}

	return nil
}
