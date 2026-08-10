// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
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

type providersLister interface {
	List(ctx context.Context) ([]oidcproviders.LightOIDCProvider, error)
}

type Deps struct {
	AuthService      auth.AuthService
	Cookies          *sessionshttp.CookieManager
	OIDCStateCookies *sessionshttp.CookieManager
	ProvidersLister  providersLister
	Logger           *slog.Logger
	Now              func() time.Time
}

func (deps *Deps) Validate() error {
	if deps.Cookies == nil {
		return errors.New("cookies is required")
	}

	if deps.OIDCStateCookies == nil {
		return errors.New("oidcStateCookies is required")
	}

	if deps.Logger == nil {
		return errors.New("logger is required")
	}

	if deps.AuthService == nil {
		return errors.New("authService is required")
	}

	if deps.ProvidersLister == nil {
		return errors.New("providersLister is required")
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
		r.Post("/logout", c.logout)
		r.Get("/providers", c.listProviders)
		r.Get("/oidc/{id}/start", c.startOIDCLogin)
		r.Get("/oidc/callback", c.oidcCallback)
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

func (c *Controller) logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if token := c.deps.Cookies.Read(r); token != "" {
		if err := c.deps.AuthService.Logout(ctx, token); err != nil {
			c.deps.Logger.ErrorContext(ctx, "failed to logout", logging.Err(err))
			httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")

			return
		}
	}

	c.deps.Cookies.Clear(w)
	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) listProviders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	providers, err := c.deps.ProvidersLister.List(ctx)
	if err != nil {
		c.deps.Logger.ErrorContext(ctx, "failed to list oidc providers", logging.Err(err))
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")

		return
	}

	res := make([]ProviderSummaryResponse, 0, len(providers))
	for _, p := range providers {
		res = append(res, ProviderSummaryResponse{
			ID:          p.ID.String(),
			DisplayName: p.DisplayName,
		})
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, res)
}

func (c *Controller) startOIDCLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Redirect(w, r, "/login?error=oidcUnavailable", http.StatusFound)

		return
	}

	res, err := c.deps.AuthService.StartOIDCLogin(ctx, auth.StartOIDCLoginOpts{
		ProviderID: id,
		Redirect:   r.URL.Query().Get("redirect"),
	})
	if err != nil {
		c.logOIDCError(ctx, "failed to start oidc login", err)
		http.Redirect(w, r, "/login?error="+oidcErrorCode(err), http.StatusFound)

		return
	}

	c.deps.OIDCStateCookies.Set(w, res.StateCookieValue, res.ExpiresAt, c.deps.Now())
	http.Redirect(w, r, res.AuthCodeURL, http.StatusFound)
}

func (c *Controller) oidcCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stateCookieValue := c.deps.OIDCStateCookies.Read(r)
	c.deps.OIDCStateCookies.Clear(w)

	query := r.URL.Query()

	res, err := c.deps.AuthService.FinishOIDCLogin(ctx, auth.FinishOIDCLoginOpts{
		Code:             query.Get("code"),
		State:            query.Get("state"),
		ErrorParam:       query.Get("error"),
		StateCookieValue: stateCookieValue,
	})
	if err != nil {
		c.logOIDCError(ctx, "failed to finish oidc login", err)
		http.Redirect(w, r, "/login?error="+oidcErrorCode(err), http.StatusFound)

		return
	}

	c.deps.Cookies.Set(w, res.Session.Token, res.Session.ExpiresAt, c.deps.Now())
	http.Redirect(w, r, res.Redirect, http.StatusFound)
}

func (c *Controller) logOIDCError(ctx context.Context, msg string, err error) {
	if isExpectedOIDCOutcome(err) {
		c.deps.Logger.WarnContext(ctx, msg, logging.Err(err))

		return
	}

	c.deps.Logger.ErrorContext(ctx, msg, logging.Err(err))
}

func isExpectedOIDCOutcome(err error) bool {
	return errors.Is(err, auth.ErrOIDCDenied) ||
		errors.Is(err, auth.ErrOIDCState) ||
		errors.Is(err, auth.ErrOIDCNotAllowed) ||
		errors.Is(err, auth.ErrOIDCNoAccount)
}

func oidcErrorCode(err error) string {
	switch {
	case errors.Is(err, auth.ErrOIDCDenied):
		return "oidcDenied"
	case errors.Is(err, auth.ErrOIDCNotAllowed):
		return "oidcNotAllowed"
	case errors.Is(err, auth.ErrOIDCNoAccount):
		return "oidcNoAccount"
	case errors.Is(err, auth.ErrOIDCState):
		return "oidcState"
	case errors.Is(err, auth.ErrOIDCUnavailable):
		return "oidcUnavailable"
	default:
		return "oidcUnavailable"
	}
}
