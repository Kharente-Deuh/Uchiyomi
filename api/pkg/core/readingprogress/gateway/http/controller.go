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
	"github.com/google/uuid"
	httpsession "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readingprogress"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/httputils"
)

type Config struct {
	Endpoint         string
	ChaptersEndpoint string
	Middlewares      chi.Middlewares
}

func (cfg *Config) Validate() error {
	if cfg.Endpoint == "" {
		return errors.New("endpoint is required")
	}

	if !strings.HasPrefix(cfg.Endpoint, "/") {
		return fmt.Errorf("endpoint must start with '/', got %q", cfg.Endpoint)
	}

	if cfg.ChaptersEndpoint == "" {
		return errors.New("chapters endpoint is required")
	}

	if !strings.HasPrefix(cfg.ChaptersEndpoint, "/") {
		return fmt.Errorf("chapters endpoint must start with '/', got %q", cfg.ChaptersEndpoint)
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
	Service readingprogress.ReadingProgressService
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

	deps.Logger = deps.Logger.With("component", "readingprogress.gateway.http")

	return &Controller{deps: deps, cfg: cfg}, nil
}

func (c *Controller) InitRouter(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(c.cfg.Middlewares...)
		r.Get(c.cfg.Endpoint+"/{id}/progress", c.get)
		r.Post(c.cfg.Endpoint+"/{id}/progress", c.post)
		r.Put(c.cfg.ChaptersEndpoint+"/{id}/progress", c.put)
	})
}

func (c *Controller) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := httpsession.UserFrom(ctx)
	if !ok {
		httputils.WriteError(w, c.deps.Logger, http.StatusUnauthorized, "")

		return
	}

	comicID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "id must be a valid UUID")

		return
	}

	result, err := c.deps.Service.List(ctx, readingprogress.ListOpts{
		UserID:  user.ID,
		ComicID: comicID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputils.WriteError(w, c.deps.Logger, http.StatusNotFound, "")

			return
		}

		if errors.Is(err, domain.ErrForbidden) {
			httputils.WriteError(w, c.deps.Logger, http.StatusForbidden, "comic not in library")

			return
		}

		c.deps.Logger.Error("failed to list reading progress", "error", err)
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, listFromDomain(result))
}

func (c *Controller) post(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := httpsession.UserFrom(ctx)
	if !ok {
		httputils.WriteError(w, c.deps.Logger, http.StatusUnauthorized, "")

		return
	}

	comicID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "id must be a valid UUID")

		return
	}

	req, err := httputils.DecodeJSON[markReadRequest](r)
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "invalid request body")

		return
	}

	result, err := c.deps.Service.MarkRead(ctx, readingprogress.MarkReadOpts{
		UserID:     user.ID,
		ComicID:    comicID,
		ChapterIDs: req.ChapterIDs,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputils.WriteError(w, c.deps.Logger, http.StatusNotFound, "")

			return
		}

		if errors.Is(err, domain.ErrForbidden) {
			httputils.WriteError(w, c.deps.Logger, http.StatusForbidden, "comic not in library")

			return
		}

		if errors.Is(err, readingprogress.ErrInvalid) {
			httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "invalid request body")

			return
		}

		c.deps.Logger.Error("failed to mark reading progress", "error", err)
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, listFromDomain(result))
}

func (c *Controller) put(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := httpsession.UserFrom(ctx)
	if !ok {
		httputils.WriteError(w, c.deps.Logger, http.StatusUnauthorized, "")

		return
	}

	chapterID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "id must be a valid UUID")

		return
	}

	req, err := httputils.DecodeJSON[saveRequest](r)
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "invalid request body")

		return
	}

	if *req.Page < 1 {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "invalid request body")

		return
	}

	saved, err := c.deps.Service.Save(ctx, readingprogress.SaveOpts{
		UserID:    user.ID,
		ChapterID: chapterID,
		Page:      *req.Page,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputils.WriteError(w, c.deps.Logger, http.StatusNotFound, "")

			return
		}

		if errors.Is(err, domain.ErrForbidden) {
			httputils.WriteError(w, c.deps.Logger, http.StatusForbidden, "comic not in library")

			return
		}

		if errors.Is(err, readingprogress.ErrInvalid) {
			httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "invalid request body")

			return
		}

		c.deps.Logger.Error("failed to save reading progress", "error", err)
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, progressFromDomain(saved))
}
