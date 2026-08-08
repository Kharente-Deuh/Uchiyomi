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
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth"
	sessionshttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
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
	AuthService auth.AuthService
	Cookies     *sessionshttp.CookieManager
	Logger      *slog.Logger
	Now         func() time.Time
}

func (deps *Deps) Validate() error {
	if deps.Cookies == nil {
		return errors.New("cookies is required")
	}

	if deps.Logger == nil {
		return errors.New("logger is required")
	}

	if deps.AuthService == nil {
		return errors.New("authService is required")
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

	deps.Logger = deps.Logger.With("component", "auth.gateway.http")

	c := &Controller{
		cfg:  cfg,
		deps: deps,
	}

	return c, nil
}

func (c *Controller) InitRouter(r chi.Router) {
	r.Route(c.cfg.Endpoint, func(r chi.Router) {
		r.Post("/login", c.loginWithPwd)
	})
}

func (c *Controller) loginWithPwd(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := httputils.DecodeJSON[LoginWithPwdRequest](r)
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "invalid request body")

		return
	}

	res, err := c.deps.AuthService.LoginWithPwd(ctx, auth.LoginWithPwdOpts{
		Username: req.Username.String(),
		Password: req.Password.String(),
	})

	switch {
	case errors.Is(err, auth.ErrInvalidLoginPwd):
		httputils.WriteError(w, c.deps.Logger, http.StatusUnauthorized, "invalid login/password")

		return
	case err != nil:
		c.deps.Logger.ErrorContext(ctx, "failed to login", logging.Err(err))
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")

		return
	}

	c.deps.Cookies.Set(w, res.Session.Token, res.Session.ExpiresAt, c.deps.Now())

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, LoginWithPwdResponse{
		ID:       res.User.ID.String(),
		Username: res.User.Name,
		IsAdmin:  res.User.IsAdmin,
	})
}
