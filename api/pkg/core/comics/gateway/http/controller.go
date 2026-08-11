// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"
	httpsession "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
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
	ComicsService comics.ComicsService
	Logger        *slog.Logger
}

func (deps *Deps) Validate() error {
	if deps.ComicsService == nil {
		return errors.New("comics service is required")
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

		r.Post("/", c.create)
		r.Get("/", c.getMany)
		r.Get("/{id}", c.getByID)
		r.Delete("/{id}", c.deleteByID)
	})
}

func (c *Controller) create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.deps.Logger.Error("failed to decode request", "error", err)
		http.Error(w, "invalid request", http.StatusBadRequest)

		return
	}

	ctx := r.Context()

	user, ok := httpsession.UserFrom(ctx)
	if !ok {
		c.deps.Logger.Error("user not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)

		return
	}

	comic, err := c.deps.ComicsService.Create(r.Context(), comics.CreateOpts{
		Source: req.Source,
		Slug:   req.Slug,
		UserID: user.ID,
	})

	if err != nil {
		c.deps.Logger.Error("failed to create comic", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusCreated, lightComicFromDomain(comic))
}

func (c *Controller) getByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := httpsession.UserFrom(ctx)
	if !ok {
		c.deps.Logger.Error("user not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)

		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "id must be a valid UUID")

		return
	}

	comic, err := c.deps.ComicsService.GetByID(ctx, comics.GetByIDOpts{
		UserID: user.ID,
		ID:     id,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputils.WriteError(w, c.deps.Logger, http.StatusNotFound, "comic not found")

			return
		}

		c.deps.Logger.Error("failed to get comic by id", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, comicResponseFromDomain(comic))
}

func (c *Controller) getMany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := httpsession.UserFrom(ctx)
	if !ok {
		c.deps.Logger.Error("user not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)

		return
	}

	opts, err := c.getManyQuery(r)
	if err != nil {
		c.deps.Logger.Error("failed to get many comics", "error", err)
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "invalid query parameters")

		return
	}

	opts.UserID = &user.ID

	comics, err := c.deps.ComicsService.GetMany(ctx, *opts)
	if err != nil {
		c.deps.Logger.Error("failed to get many comics", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	res := make([]lightComic, len(comics))
	for i, comic := range comics {
		res[i] = lightComicFromDomain(&comic)
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, res)
}

func (c *Controller) deleteByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := httpsession.UserFrom(ctx)
	if !ok {
		c.deps.Logger.Error("user not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)

		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "id must be a valid UUID")

		return
	}

	err = c.deps.ComicsService.Delete(ctx, comics.DeleteOpts{
		UserID: user.ID,
		ID:     id,
	})
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		c.deps.Logger.Error("failed to delete comic by id", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, "")
}

func (c *Controller) getManyQuery(r *http.Request) (*comics.GetManyOpts, error) {
	opts := &comics.GetManyOpts{}

	source := r.URL.Query().Get("source")
	if source != "" {
		parsed, err := sources.ParseSourceName(source)
		if err != nil {
			return nil, fmt.Errorf("sources.ParseSourceName: %w", err)
		}

		opts.Source = &parsed
	}

	comicType := r.URL.Query().Get("type")
	if comicType != "" {
		parsed, err := sources.ParseSeriesType(comicType)
		if err != nil {
			//nolint:wrapcheck
			return nil, err
		}

		opts.Type = &parsed
	}

	comicStatus := r.URL.Query().Get("status")
	if comicStatus != "" {
		parsed, err := sources.ParseSeriesStatus(comicStatus)
		if err != nil {
			//nolint:wrapcheck
			return nil, err
		}

		opts.Status = &parsed
	}

	limit, err := parseQueryLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, fmt.Errorf("parseQueryLimit: %w", err)
	}

	opts.Limit = limit

	offset, err := parseQueryOffset(r.URL.Query().Get("offset"))
	if err != nil {
		return nil, fmt.Errorf("parseQueryOffset: %w", err)
	}

	opts.Offset = offset

	return opts, nil
}

func parseQueryLimit(q string) (int, error) {
	if q == "" {
		return 10, nil
	}

	limit, err := strconv.Atoi(q)
	if err != nil {
		return 0, fmt.Errorf("invalid limit: %w", err)
	}

	if limit < 1 {
		return 10, nil
	}

	if limit > 100 {
		return 100, nil
	}

	return limit, nil
}

func parseQueryOffset(q string) (int, error) {
	if q == "" {
		return 0, nil
	}

	offset, err := strconv.Atoi(q)
	if err != nil {
		return 0, fmt.Errorf("invalid offset: %w", err)
	}

	if offset < 0 {
		return 0, nil
	}

	return offset, nil
}
