// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	sessionshttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/setup"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/httputils"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/logging"
)

type Config struct {
	Endpoint string
}

func (cfg *Config) Validate() error {
	if cfg.Endpoint == "" {
		return errors.New("endpoint is required")
	}

	if !strings.HasPrefix(cfg.Endpoint, "/") {
		return fmt.Errorf("endpoint must start with '/', got %q", cfg.Endpoint)
	}

	return nil
}

type Deps struct {
	SetupService setup.SetupService
	Cookies      *sessionshttp.CookieManager
	Logger       *slog.Logger
	Now          func() time.Time
}

func (deps *Deps) Validate() error {
	if deps.SetupService == nil {
		return errors.New("setupService is required")
	}

	if deps.Cookies == nil {
		return errors.New("cookies is required")
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

	if deps.Now == nil {
		deps.Now = time.Now
	}

	deps.Logger = deps.Logger.With("component", "setup.gateway.http")

	c := &Controller{
		cfg:  cfg,
		deps: deps,
	}

	return c, nil
}

func (c *Controller) InitRouter(r chi.Router) {
	r.Route(c.cfg.Endpoint, func(r chi.Router) {
		r.Get("/status", c.getSetupStatus)
		r.Post("/", c.doSetup)
	})
}

func (c *Controller) getSetupStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	required, err := c.deps.SetupService.IsSetupRequired(ctx)
	if err != nil {
		c.deps.Logger.ErrorContext(ctx, "failed to check setup status", logging.Err(err))
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, GetStatusResponse{Required: required})
}

func (c *Controller) doSetup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := httputils.DecodeJSON[DoSetupRequest](r)
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "invalid request body")

		return
	}

	session, err := c.deps.SetupService.DoSetup(ctx, setup.DoSetupOpts{
		Username: req.Username.String(),
		Password: req.Password.String(),
	})

	switch {
	case errors.Is(err, setup.ErrSetupNotNeeded):
		httputils.WriteError(w, c.deps.Logger, http.StatusConflict, "setup has already been completed")

		return
	case errors.Is(err, setup.ErrSessionNotIssued):
		c.deps.Logger.ErrorContext(ctx, "failed to issue setup session", logging.Err(err))
	case err != nil:
		c.deps.Logger.ErrorContext(ctx, "failed to complete setup", logging.Err(err))
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")

		return
	}

	if session != nil {
		c.deps.Cookies.Set(w, session.Token, session.ExpiresAt, c.deps.Now())
	}

	w.WriteHeader(http.StatusCreated)
}
