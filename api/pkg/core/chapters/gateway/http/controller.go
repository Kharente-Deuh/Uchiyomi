// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	httpsession "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readingprogress"
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

type ProgressReader interface {
	MapByChapterIDs(context.Context, readingprogress.MapOpts) (map[uuid.UUID]readingprogress.Progress, error)
}

type Deps struct {
	ChaptersService chapters.ChaptersService
	Progress        ProgressReader
	Logger          *slog.Logger
}

func (deps *Deps) Validate() error {
	if deps.ChaptersService == nil {
		return errors.New("chapters service is required")
	}

	if deps.Progress == nil {
		return errors.New("progress is required")
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

		r.Get("/", c.listForLibrary)
		r.Get("/{id}/pages/{index}", c.servePage)
		r.Get("/{id}", c.getByID)
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

	c.writeChapterList(w, r, user.ID, res)
}

func (c *Controller) listForLibrary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := httpsession.UserFrom(ctx)
	if !ok {
		c.deps.Logger.Error("user not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)

		return
	}

	rawComicID := strings.TrimSpace(r.URL.Query().Get("comicId"))
	if rawComicID == "" {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "comicId is required")

		return
	}

	comicID, err := uuid.Parse(rawComicID)
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "comicId must be a valid UUID")

		return
	}

	res, err := c.deps.ChaptersService.ListForLibrary(ctx, chapters.ListForLibraryOpts{
		UserID:  user.ID,
		ComicID: comicID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputils.WriteError(w, c.deps.Logger, http.StatusNotFound, "comic not found")

			return
		}

		if errors.Is(err, domain.ErrForbidden) {
			httputils.WriteError(w, c.deps.Logger, http.StatusForbidden, "comic not in library")

			return
		}

		c.deps.Logger.Error("failed to list chapters", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	c.writeChapterList(w, r, user.ID, res)
}

func (c *Controller) getByID(w http.ResponseWriter, r *http.Request) {
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

	chapter, err := c.deps.ChaptersService.GetDetailForLibrary(ctx, chapters.GetForLibraryOpts{
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

		c.deps.Logger.Error("failed to get chapter", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	item, err := c.chapterDetailItem(r.Context(), user.ID, *chapter)
	if err != nil {
		c.deps.Logger.Error("failed to map reading progress", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, item)
}

func (c *Controller) servePage(w http.ResponseWriter, r *http.Request) {
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

	index, err := strconv.Atoi(chi.URLParam(r, "index"))
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusNotFound, "page not found")

		return
	}

	diskPath, contentType, err := c.deps.ChaptersService.ServePage(ctx, chapters.ServePageOpts{
		UserID:    user.ID,
		ChapterID: id,
		Index:     index,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httputils.WriteError(w, c.deps.Logger, http.StatusNotFound, "page not found")

			return
		}

		if errors.Is(err, domain.ErrForbidden) {
			httputils.WriteError(w, c.deps.Logger, http.StatusForbidden, "comic not in library")

			return
		}

		c.deps.Logger.Error("failed to serve chapter page", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	f, err := os.Open(diskPath)
	if err != nil {
		c.deps.Logger.Error("failed to open chapter page", "path", diskPath, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	defer f.Close()

	w.Header().Set("Content-Type", contentType)
	http.ServeContent(w, r, diskPath, time.Time{}, f)
}

func (c *Controller) writeChapterList(
	w http.ResponseWriter, r *http.Request, userID uuid.UUID, res []chapters.Chapter,
) {
	byID, err := c.progressByIDs(r.Context(), userID, res)
	if err != nil {
		c.deps.Logger.Error("failed to map reading progress", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, chapterHTTPList(res, byID))
}

func (c *Controller) chapterDetailItem(
	ctx context.Context, userID uuid.UUID, detail chapters.ChapterDetail,
) (chapterDetailResponse, error) {
	item, err := c.chapterItem(ctx, userID, detail.Chapter)
	if err != nil {
		return chapterDetailResponse{}, fmt.Errorf("c.chapterItem: %w", err)
	}

	return chapterDetailResponse{
		postListResponseChapter: item,
		PageURLs:                pageURLs(detail.Chapter.ID, detail.Chapter.Download, detail.Chapter.PagesNb),
		NextChapterID:           detail.NextID,
		PreviousChapterID:       detail.PreviousID,
	}, nil
}

func (c *Controller) chapterItem(
	ctx context.Context, userID uuid.UUID, chapter chapters.Chapter,
) (postListResponseChapter, error) {
	byID, err := c.progressByIDs(ctx, userID, []chapters.Chapter{chapter})
	if err != nil {
		return postListResponseChapter{}, err
	}

	return chapterHTTPItem(chapter, byID), nil
}

func (c *Controller) progressByIDs(
	ctx context.Context, userID uuid.UUID, res []chapters.Chapter,
) (map[uuid.UUID]readingprogress.Progress, error) {
	ids := make([]uuid.UUID, len(res))
	for i := range res {
		ids[i] = res[i].ID
	}

	byID, err := c.deps.Progress.MapByChapterIDs(ctx, readingprogress.MapOpts{
		UserID: userID,
		IDs:    ids,
	})
	if err != nil {
		return nil, fmt.Errorf("c.deps.Progress.MapByChapterIDs: %w", err)
	}

	return byID, nil
}

func chapterHTTPList(
	res []chapters.Chapter, byID map[uuid.UUID]readingprogress.Progress,
) []postListResponseChapter {
	out := make([]postListResponseChapter, 0, len(res))
	for _, chapter := range res {
		out = append(out, chapterHTTPItem(chapter, byID))
	}

	return out
}

func chapterHTTPItem(
	chapter chapters.Chapter, byID map[uuid.UUID]readingprogress.Progress,
) postListResponseChapter {
	return postListResponseChapter{
		PublishedAt:       chapter.PublishedAt,
		EarlyAccessUntil:  utils.OptionalTime(chapter.EarlyAccessUntil),
		Progress:          progressPayload(chapter.ID, chapter.PagesNb, byID),
		SourceChapterSlug: chapter.SourceChapterSlug,
		Title:             chapter.Title,
		Number:            chapter.Number,
		PagesNb:           chapter.PagesNb,
		Download:          chapter.Download,
		ID:                chapter.ID,
		ComicID:           chapter.ComicID,
	}
}
