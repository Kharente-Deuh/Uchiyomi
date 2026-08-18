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
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
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
	ChaptersService chapters.ChaptersService
	Logger          *slog.Logger
}

func (deps *Deps) Validate() error {
	if deps.ChaptersService == nil {
		return errors.New("chapters service is required")
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

	return &Controller{deps: deps, cfg: cfg}, nil
}

func (c *Controller) InitRouter(r chi.Router) {
	r.Route(c.cfg.Endpoint, func(r chi.Router) {
		r.Use(c.cfg.Middlewares...)

		r.Post("/{id}/retry", c.retryDownload)
		r.Post("/list", c.postList)
	})
}

func (c *Controller) retryDownload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := httpsession.UserFrom(ctx)
	if !ok {
		c.deps.Logger.Error("user not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)

		return
	}

	chapterID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "id must be a valid UUID")

		return
	}

	err = c.deps.ChaptersService.RetryDownload(ctx, chapters.RetryDownloadOpts{
		UserID:    user.ID,
		ChapterID: chapterID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputils.WriteError(w, c.deps.Logger, http.StatusNotFound, "chapter not found")

			return
		}

		if errors.Is(err, domain.ErrForbidden) {
			httputils.WriteError(w, c.deps.Logger, http.StatusForbidden, "comic not in library")

			return
		}

		if errors.Is(err, domain.ErrConflict) {
			httputils.WriteError(w, c.deps.Logger, http.StatusConflict, "chapter already downloaded")

			return
		}

		c.deps.Logger.Error("failed to retry chapter download", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusAccepted, "")
}

func (c *Controller) postList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := httpsession.UserFrom(ctx)
	if !ok {
		c.deps.Logger.Error("user not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)

		return
	}

	req, err := httputils.DecodeJSON[postListBody](r)
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "invalid request body")

		return
	}

	res, err := c.deps.ChaptersService.GetByIds(ctx, chapters.GetByIdsOpts{
		UserID: user.ID,
		IDs:    req.IDs,
	})
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "failed to get chapters")

		return
	}

	chapters := make([]postListResponseChapter, 0, len(res))
	for _, chapter := range res {
		chapters = append(chapters, postListResponseChapter{
			PublishedAt:       chapter.PublishedAt,
			EarlyAccessUntil:  utils.OptionalTime(chapter.EarlyAccessUntil),
			SourceChapterSlug: chapter.SourceChapterSlug,
			Title:             chapter.Title,
			Number:            chapter.Number,
			PagesNb:           chapter.PagesNb,
			Download:          chapter.Download,
			ID:                chapter.ID,
			ComicID:           chapter.ComicID,
		})
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, chapters)
}
