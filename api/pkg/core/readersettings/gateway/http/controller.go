// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	httpsession "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readersettings"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/httputils"
)

type Config struct {
	Endpoint    string
	Middlewares chi.Middlewares
}

func (cfg *Config) Validate() error {
	if cfg.Endpoint == "" {
		return errors.New("endpoint is required")
	}

	if !strings.HasPrefix(cfg.Endpoint, "/") {
		return fmt.Errorf("endpoint must start with '/', got %q", cfg.Endpoint)
	}

	hasNilMiddlewares := slices.ContainsFunc(cfg.Middlewares, func(m func(http.Handler) http.Handler) bool {
		return m == nil
	})

	if hasNilMiddlewares {
		return errors.New("all middlewares must not be nil")
	}

	return nil
}

type Deps struct {
	Service readersettings.ReaderSettingsService
	Logger  *slog.Logger
}

func (deps *Deps) Validate() error {
	if deps.Service == nil {
		return errors.New("service is required")
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

	deps.Logger = deps.Logger.With("component", "readersettings.gateway.http")

	return &Controller{deps: deps, cfg: cfg}, nil
}

func (c *Controller) InitRouter(r chi.Router) {
	r.Route(c.cfg.Endpoint, func(r chi.Router) {
		r.Use(c.cfg.Middlewares...)
		r.Get("/reader-settings", c.list)
		r.Put("/reader-settings/{type}", c.replace)
	})
}

func (c *Controller) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := httpsession.UserFrom(ctx)
	if !ok {
		httputils.WriteError(w, c.deps.Logger, http.StatusUnauthorized, "")

		return
	}

	profiles, err := c.deps.Service.ListForUser(ctx, user.ID)
	if err != nil {
		c.deps.Logger.Error("failed to list reader settings", "error", err)
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, listFromDomain(profiles))
}

func (c *Controller) replace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := httpsession.UserFrom(ctx)
	if !ok {
		httputils.WriteError(w, c.deps.Logger, http.StatusUnauthorized, "")

		return
	}

	comicType, err := sources.ParseSeriesType(chi.URLParam(r, "type"))
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "invalid comic type")

		return
	}

	req, err := httputils.DecodeJSON[replaceRequest](r)
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "invalid request body")

		return
	}

	readingMode, err := readersettings.ParseReadingMode(req.ReadingMode)
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "invalid request body")

		return
	}

	pageScale, err := readersettings.ParsePageScale(req.PageScale)
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "invalid request body")

		return
	}

	profile, err := c.deps.Service.Replace(ctx, readersettings.ReplaceOpts{
		UserID:      user.ID,
		Type:        comicType,
		ReadingMode: readingMode,
		PageScale:   pageScale,
		DoublePage:  *req.DoublePage,
	})
	if err != nil {
		if errors.Is(err, readersettings.ErrInvalid) {
			httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "invalid request body")

			return
		}

		c.deps.Logger.Error("failed to replace reader settings", "error", err)
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, profileFromDomain(profile))
}
