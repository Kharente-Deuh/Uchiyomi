// SPDX-License-Identifier: AGPL-3.0-or-later

package refresh

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/logging"
)

type Config struct {
	Interval time.Duration
}

func (cfg *Config) Validate() error {
	if cfg.Interval < time.Minute {
		return errors.New("interval must be at least 1 minute")
	}

	return nil
}

type Deps struct {
	ComicsService comics.ComicsService
	Logger        *slog.Logger
}

func (deps *Deps) Validate() error {
	if deps.ComicsService == nil {
		return errors.New("comicsService is required")
	}

	return nil
}

type App struct {
	deps    Deps
	cfg     Config
	running atomic.Bool
}

func New(cfg Config, deps Deps) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}

	deps.Logger = deps.Logger.With("component", "comics.refresh")

	return &App{deps: deps, cfg: cfg}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.deps.Logger.Info("chapter list refresh started")

	//nolint:wrapcheck
	return utils.Loop(ctx, utils.LoopOpts{
		Interval: a.cfg.Interval,
		Fn:       a.refresh,
	})
}

func (a *App) refresh(ctx context.Context) error {
	if !a.running.CompareAndSwap(false, true) {
		a.deps.Logger.WarnContext(ctx, "chapter list refresh skipped: already running")

		return nil
	}

	defer a.running.Store(false)

	if err := a.deps.ComicsService.RefreshChapterLists(ctx); err != nil {
		a.deps.Logger.ErrorContext(ctx, "chapter list refresh failed", logging.Err(err))
	}

	return nil
}
