// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/health"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/httputils"
)

const DefaultProbeTimeout = 2 * time.Second

type Config struct {
	ProbeTimeout time.Duration
}

func (cfg *Config) Validate() error {
	if cfg.ProbeTimeout <= 0 {
		return errors.New("probeTimeout must be greater than 0")
	}

	return nil
}

type Deps struct {
	Registry *health.Registry
	Logger   *slog.Logger
}

func (deps *Deps) Validate() error {
	if deps.Registry == nil {
		return errors.New("registry is required")
	}

	if deps.Logger == nil {
		return errors.New("logger is required")
	}

	return nil
}

type Controller struct {
	deps Deps
	cfg  Config
}

func New(cfg Config, deps Deps) (*Controller, error) {
	var err error
	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err = deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	deps.Logger = deps.Logger.With("component", "health.gateway.http")

	c := &Controller{
		cfg:  cfg,
		deps: deps,
	}

	return c, nil
}

func (c *Controller) InitRouter(r chi.Router) {
	r.Get("/healthz", c.getHealthz)
	r.Get("/readyz", c.getReadyz)
}

func (c *Controller) getHealthz(w http.ResponseWriter, _ *http.Request) {
	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, statusResponse{Status: string(health.StatusOK)})
}

func (c *Controller) getReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), c.cfg.ProbeTimeout)
	defer cancel()

	rep := c.deps.Registry.Snapshot(ctx)

	status := http.StatusOK
	if rep.Status != health.StatusOK {
		status = http.StatusServiceUnavailable
	}

	c.logProbeFailures(ctx, rep)

	httputils.WriteJSON(w, c.deps.Logger, status, newReadyzResponse(rep))
}

func (c *Controller) logProbeFailures(ctx context.Context, rep health.Report) {
	for name, comp := range rep.Components {
		if !comp.Probe || comp.Reason == "" {
			continue
		}

		c.deps.Logger.WarnContext(ctx, "readiness probe failed", "probe", name, "reason", comp.Reason)
	}
}
