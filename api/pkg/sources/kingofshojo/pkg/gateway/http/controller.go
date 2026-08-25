// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	httpsession "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/core"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/parse"
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
	KingOfShojoApp  *core.App
	Logger          *slog.Logger
	CoverURLBuilder func(source, slug string) string
}

func (deps *Deps) Validate() error {
	if deps.KingOfShojoApp == nil {
		return errors.New("kingOfShojoApp is required")
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

	deps.Logger = deps.Logger.With("component", "sources.kingofshojo.gateway.http")

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

	user, ok := httpsession.UserFrom(ctx)
	if !ok {
		c.deps.Logger.ErrorContext(ctx, "user not found")
		httputils.WriteError(w, c.deps.Logger, http.StatusUnauthorized, "user not found")

		return
	}

	opts, err := parseSearchOpts(r.URL.Query())
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, err.Error())

		return
	}

	opts.UserID = user.ID

	c.deps.Logger.Debug("parseSearchOpts", "opts", opts)

	res, err := c.deps.KingOfShojoApp.Search(ctx, opts)
	if err != nil {
		c.writeSourceError(w, ctx, "failed to search", err)

		return
	}

	buildCover := c.deps.CoverURLBuilder
	source := string(sources.SourceKingOfShojo)

	dto := searchResDTO{
		Total: res.Meta.Total,
		Items: utils.MapSlice(res.Items, func(i domain.SearchResultItem) searchResItemDTO {
			return searchResItemDTO{
				LastChapterAt: utils.OptionalTime(i.LastChapterAt),
				UpdatedAt:     i.UpdatedAt,
				CreatedAt:     i.CreatedAt,
				PublicURL:     i.PublicURL,
				SourceURL:     i.SourceURL,
				Cover:         buildCover(source, i.Slug),
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
				Rating:        0,
				ReleaseYear:   i.ReleaseYear,
				InternalID:    i.InternalID,
				LatestChapters: utils.MapSlice(i.LatestChapters, func(lc domain.SearchResultItemChapter) searchResItemChapterDTO {
					return searchResItemChapterDTO{
						EarlyAccessUntil: utils.OptionalTime(lc.EarlyAccessUntil),
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

	user, ok := httpsession.UserFrom(ctx)
	if !ok {
		c.deps.Logger.ErrorContext(ctx, "user not found")
		httputils.WriteError(w, c.deps.Logger, http.StatusUnauthorized, "user not found")

		return
	}

	s, err := c.deps.KingOfShojoApp.GetInfosBySlug(ctx, sources.GetInfosBySlugOpts{
		Slug:   slug,
		UserID: user.ID,
	})
	if err != nil {
		c.writeSourceError(w, ctx, "failed to fetch series", err, "slug", slug)

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, getInfosBySlugResDTO{
		LastChapterAt: utils.OptionalTime(s.LastChapterAt),
		UpdatedAt:     s.UpdatedAt,
		CreatedAt:     s.CreatedAt,
		Description:   s.Description,
		Title:         s.Title,
		Cover:         c.deps.CoverURLBuilder(string(sources.SourceKingOfShojo), s.Slug),
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
		Rating:        0,
		InternalID:    s.InternalID,
	})
}

func (c *Controller) getChaptersListBySeries(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "seriesSlug")
	ctx := r.Context()

	user, ok := httpsession.UserFrom(ctx)
	if !ok {
		c.deps.Logger.ErrorContext(ctx, "user not found")
		httputils.WriteError(w, c.deps.Logger, http.StatusUnauthorized, "user not found")

		return
	}

	chapters, err := c.deps.KingOfShojoApp.GetChaptersListBySeries(ctx, slug, user.ID)
	if err != nil {
		c.writeSourceError(w, ctx, "failed to fetch chapters", err, "slug", slug)

		return
	}

	dto := utils.MapSlice(chapters, func(ch domain.Chapter) getChaptersListBySeriesResItemDTO {
		return getChaptersListBySeriesResItemDTO{
			EarlyAccessUntil: utils.OptionalTime(ch.EarlyAccessUntil),
			PublishedAt:      ch.PublishedAt,
			InternalID:       ch.InternalID,
			Download:         ch.Download,
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

	urls, err := c.deps.KingOfShojoApp.GetImageURLsByChapter(ctx, opts)
	if err != nil {
		c.writeSourceError(w, ctx, "failed to fetch images", err, "slug", opts.SeriesSlug, "chapterId", opts.ChapterID)

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, getImageURLsByChapterResDTO{URLs: urls})
}

func (c *Controller) writeSourceError(
	w http.ResponseWriter,
	ctx context.Context,
	msg string,
	err error,
	args ...any,
) {
	switch {
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, parse.ErrUnsupportedType):
		c.deps.Logger.ErrorContext(ctx, msg, append(args, logging.Err(err))...)
		httputils.WriteError(w, c.deps.Logger, http.StatusNotFound, "")

	case errors.Is(err, domain.ErrChallenge):
		c.deps.Logger.ErrorContext(ctx, msg, append(args, logging.Err(err))...)
		httputils.WriteError(w, c.deps.Logger, http.StatusServiceUnavailable, "")

	default:
		c.deps.Logger.ErrorContext(ctx, msg, append(args, logging.Err(err))...)
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")
	}
}
