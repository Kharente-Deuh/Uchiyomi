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
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/core"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/httputils"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/logging"
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
	AsuraApp        *core.App
	Logger          *slog.Logger
	CoverURLBuilder func(source, slug string) string
}

func (deps *Deps) Validate() error {
	if deps.AsuraApp == nil {
		return errors.New("asuraApp is required")
	}

	if deps.Logger == nil {
		return errors.New("logger is required")
	}

	if deps.CoverURLBuilder == nil {
		return errors.New("coverURLBuilder is required")
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

	deps.Logger = deps.Logger.With("component", "sources.asurascans.gateway.http")

	c := &Controller{
		cfg:  cfg,
		deps: deps,
	}

	return c, nil
}

func (c *Controller) InitRouter(r chi.Router) {
	seriesSlugParam := "{seriesSlug:[a-z0-9-]+}"
	r.Route(c.cfg.Endpoint, func(r chi.Router) {
		for _, m := range c.cfg.Middlewares {
			r.Use(m)
		}

		r.Get("/search", c.search)
		r.Get(fmt.Sprintf("/series/%s", seriesSlugParam), c.getInfosBySlug)
		r.Get(fmt.Sprintf("/series/%s/chapters", seriesSlugParam), c.getChaptersListBySeries)
		r.Get(fmt.Sprintf("/series/%s/chapters/{chapterID}", seriesSlugParam), c.getImageURLsByChapter)
	})
}

func (c *Controller) search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	opts, err := parseSearchOpts(r.URL.Query())
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, err.Error())

		return
	}

	c.deps.Logger.Debug("parseSearchOpts", "opts", opts)

	res, err := c.deps.AsuraApp.Search(ctx, opts)
	if err != nil {
		c.deps.Logger.ErrorContext(ctx, "failt to search", "error", err)
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")

		return
	}

	buildCover := c.deps.CoverURLBuilder

	dto := searchResDTO{
		Total: res.Meta.Total,
		Items: utils.MapSlice(res.Items, func(i domain.SearchResultItem) searchResItemDTO {
			return searchResItemDTO{
				LastChapterAt: i.LastChapterAt,
				UpdatedAt:     i.UpdatedAt,
				CreatedAt:     i.CreatedAt,
				PublicURL:     i.PublicURL,
				SourceURL:     i.SourceURL,
				Cover:         buildCover(covers.SourceAsuraScans, i.Slug),
				Status:        i.Status,
				Type:          i.Type,
				Author:        i.Author,
				Artist:        i.Artist,
				Description:   i.Description,
				Slug:          i.Slug,
				Title:         i.Title,
				AltTitles:     i.AltTitles,
				Genres:        i.Genres,
				ChapterCount:  i.ChapterCount,
				Rating:        i.Rating,
				ReleaseYear:   i.ReleaseYear,
				InternalID:    i.InternalID,
				LatestChapters: utils.MapSlice(i.LatestChapters, func(lc domain.SearchResultItemChapter) searchResItemChapterDTO {
					return searchResItemChapterDTO{
						EarlyAccessUntil: lc.EarlyAccessUntil,
						PublishedAt:      lc.PublishedAt,
						Title:            lc.Title,
						ID:               lc.ID,
						Number:           lc.Number,
					}
				}),
			}
		}),
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, dto)
}

func (c *Controller) getInfosBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "seriesSlug")
	ctx := r.Context()

	s, err := c.deps.AsuraApp.GetInfosBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.deps.Logger.ErrorContext(ctx, "series not found", "slug", slug)
			httputils.WriteError(w, c.deps.Logger, http.StatusNotFound, "series not found")

			return
		}

		c.deps.Logger.ErrorContext(ctx, "failed to fetch series", "slug", slug, logging.Err(err))
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, getInfosBySlugResDTO{
		LastChapterAt: s.LastChapterAt,
		UpdatedAt:     s.UpdatedAt,
		CreatedAt:     s.CreatedAt,
		Description:   s.Description,
		Title:         s.Title,
		Cover:         c.deps.CoverURLBuilder(covers.SourceAsuraScans, s.Slug),
		Status:        s.Status,
		Type:          s.Type,
		Author:        s.Author,
		Artist:        s.Artist,
		SourceURL:     s.SourceURL,
		PublicURL:     s.PublicURL,
		Slug:          s.Slug,
		AltTitles:     s.AltTitles,
		Genres:        s.Genres,
		ChapterCount:  s.ChapterCount,
		Rating:        s.Rating,
	})
}

func (c *Controller) getChaptersListBySeries(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "seriesSlug")
	ctx := r.Context()

	chapters, err := c.deps.AsuraApp.GetChaptersListBySeries(ctx, slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.deps.Logger.ErrorContext(ctx, "series not found", "slug", slug)
			httputils.WriteError(w, c.deps.Logger, http.StatusNotFound, "series not found")

			return
		}

		c.deps.Logger.ErrorContext(ctx, "failed to fetch chapters", "slug", slug, logging.Err(err))
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")

		return
	}

	dto := utils.MapSlice(chapters, func(ch domain.Chapter) getChaptersListBySeriesResItemDTO {
		return getChaptersListBySeriesResItemDTO{
			EarlyAccessUntil: ch.EarlyAccessUntil,
			PublishedAt:      ch.PublishedAt,
			ID:               ch.ID,
			Title:            ch.Title,
			Number:           ch.Number,
		}
	})

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, dto)
}

func (c *Controller) getImageURLsByChapter(w http.ResponseWriter, r *http.Request) {
	opts := domain.GetImageURLsByChapterOpts{
		SeriesSlug: chi.URLParam(r, "seriesSlug"),
		ChapterID:  chi.URLParam(r, "chapterID"),
	}

	ctx := r.Context()

	urls, err := c.deps.AsuraApp.GetImageURLsByChapter(ctx, opts)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.deps.Logger.ErrorContext(ctx, "chapter not found", "slug", opts.SeriesSlug, "chapterId", opts.ChapterID)
			httputils.WriteError(w, c.deps.Logger, http.StatusNotFound, "chapter not found")

			return
		}

		//nolint:lll
		c.deps.Logger.ErrorContext(ctx, "failed to fetch images", "slug", opts.SeriesSlug, "chapterId", opts.ChapterID, logging.Err(err))
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, getImageURLsByChapterResDTO{URLs: urls})
}
