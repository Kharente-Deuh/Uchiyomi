// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
)

type Config struct {
	RemoveExpiredSessionsInterval time.Duration
}

func (cfg *Config) Validate() error {
	if cfg.RemoveExpiredSessionsInterval < time.Minute {
		return fmt.Errorf("cfg.RemoveExpiredSessionsInterval must be at leas 1 min")
	}

	return nil
}

type Deps struct {
	SessionsRepository SessionsRepository
	Logger             *slog.Logger
}

func (deps *Deps) Validate() error {
	if deps.SessionsRepository == nil {
		return errors.New("sessionsRepository is required")
	}

	if deps.Logger == nil {
		return errors.New("logger is required")
	}

	return nil
}

type App struct {
	deps Deps
	cfg  Config
}

func New(cfg Config, deps Deps) (*App, error) {
	var err error
	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err = deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	return &App{deps: deps, cfg: cfg}, nil
}

func (a *App) Run(ctx context.Context) error {
	//nolint:wrapcheck // Run délègue tel quel à utils.Loop, comme le reste du dépôt (voir fncache.Run).
	return utils.Loop(ctx, utils.LoopOpts{
		Interval: a.cfg.RemoveExpiredSessionsInterval,
		Fn:       a.removeExpiredSessions,
	})
}

func (a *App) removeExpiredSessions(ctx context.Context) error {
	count, err := a.deps.SessionsRepository.DeleteExpired(ctx, time.Now())
	if err != nil {
		a.deps.Logger.ErrorContext(ctx, "sessions cleanup failed", "err", err)
	} else {
		a.deps.Logger.Info("expired sessions removed", "count", count)
	}

	return nil
}
