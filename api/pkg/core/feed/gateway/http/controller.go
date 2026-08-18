// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	httpsession "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/feed"
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
	FeedService feed.FeedService
	Logger      *slog.Logger
}

func (deps *Deps) Validate() error {
	if deps.FeedService == nil {
		return errors.New("feed service is required")
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

		r.Get("/", c.get)
	})
}

func (c *Controller) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := httpsession.UserFrom(ctx)
	if !ok {
		c.deps.Logger.Error("user not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)

		return
	}

	opts, err := c.getQuery(r)
	if err != nil {
		c.deps.Logger.Error("failed to get feed", "error", err)
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "invalid query parameters")

		return
	}

	opts.UserID = user.ID

	page, err := c.deps.FeedService.Get(ctx, opts)
	if err != nil {
		c.deps.Logger.Error("failed to get feed", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, listFromPage(page))
}

func (c *Controller) getQuery(r *http.Request) (feed.GetOpts, error) {
	opts := feed.GetOpts{}

	source := r.URL.Query().Get("source")
	if source != "" {
		parsed, err := sources.ParseSourceName(source)
		if err != nil {
			return feed.GetOpts{}, fmt.Errorf("sources.ParseSourceName: %w", err)
		}

		opts.Source = &parsed
	}

	comicType := r.URL.Query().Get("type")
	if comicType != "" {
		parsed, err := sources.ParseSeriesType(comicType)
		if err != nil {
			//nolint:wrapcheck
			return feed.GetOpts{}, err
		}

		opts.Type = &parsed
	}

	comicStatus := r.URL.Query().Get("status")
	if comicStatus != "" {
		parsed, err := sources.ParseSeriesStatus(comicStatus)
		if err != nil {
			//nolint:wrapcheck
			return feed.GetOpts{}, err
		}

		opts.Status = &parsed
	}

	limit, err := parseQueryLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return feed.GetOpts{}, fmt.Errorf("parseQueryLimit: %w", err)
	}

	opts.Limit = limit

	offset, err := parseQueryOffset(r.URL.Query().Get("offset"))
	if err != nil {
		return feed.GetOpts{}, fmt.Errorf("parseQueryOffset: %w", err)
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
