// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
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
	Service *covers.Service
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
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	deps.Logger = deps.Logger.With("component", "covers.gateway.http")

	return &Controller{cfg: cfg, deps: deps}, nil
}

func (c *Controller) InitRouter(r chi.Router) {
	r.Route(c.cfg.Endpoint, func(r chi.Router) {
		for _, m := range c.cfg.Middlewares {
			r.Use(m)
		}

		r.Get("/{slug}", c.serveCover)
	})
}

func (c *Controller) serveCover(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	source := strings.TrimSpace(r.URL.Query().Get("source"))
	slug := chi.URLParam(r, "slug")

	if source == "" {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "source is required")

		return
	}

	diskPath, contentType, err := c.deps.Service.Serve(ctx, source, slug)
	if err != nil {
		switch {
		case errors.Is(err, covers.ErrUnknownSource):
			httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, err.Error())
		case errors.Is(err, covers.ErrSeriesNotFound):
			httputils.WriteError(w, c.deps.Logger, http.StatusNotFound, "series not found")
		case errors.Is(err, covers.ErrDownloadFailed):
			httputils.WriteError(w, c.deps.Logger, http.StatusBadGateway, "cover download failed")
		case errors.Is(err, covers.ErrLocalCoverMissing):
			c.deps.Logger.ErrorContext(ctx, "local cover missing", "source", source, "slug", slug, "error", err)
			httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")
		default:
			c.deps.Logger.ErrorContext(ctx, "failed to serve cover", "source", source, "slug", slug, "error", err)
			httputils.WriteError(w, c.deps.Logger, http.StatusBadGateway, "cover download failed")
		}

		return
	}

	f, err := os.Open(diskPath)
	if err != nil {
		c.deps.Logger.ErrorContext(ctx, "failed to open cached cover", "path", diskPath, "error", err)
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")

		return
	}

	defer f.Close()

	w.Header().Set("Content-Type", contentType)
	http.ServeContent(w, r, diskPath, mustStatModTime(f), f)
}

func mustStatModTime(f *os.File) (modTime time.Time) {
	st, err := f.Stat()
	if err != nil {
		return time.Time{}
	}

	return st.ModTime()
}
