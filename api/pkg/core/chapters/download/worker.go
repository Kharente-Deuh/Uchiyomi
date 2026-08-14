// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
	"golang.org/x/sync/errgroup"
)

const defaultPageRetries = 3

type Config struct {
	SourceRateLimits map[sources.SourceName]time.Duration
	Dir              string
	RateLimit        time.Duration
}

func (cfg *Config) Validate() error {
	if cfg.Dir == "" {
		return errors.New("dir is required")
	}

	if cfg.RateLimit <= 0 {
		return errors.New("rateLimit must be greater than 0")
	}

	return nil
}

type Deps struct {
	ChaptersRepository chapters.ChaptersRepository
	ComicsRepository   comics.ComicsRepository
	Sources            sources.SourceMap
	HTTPClient         *http.Client
	Logger             *slog.Logger
	Publisher          ProgressPublisher
}

func (deps *Deps) Validate() error {
	if deps.ChaptersRepository == nil {
		return errors.New("chaptersRepository is required")
	}

	if deps.ComicsRepository == nil {
		return errors.New("comicsRepository is required")
	}

	if deps.Sources == nil {
		return errors.New("sources is required")
	}

	for _, source := range deps.Sources {
		if source == nil {
			return errors.New("source is required")
		}
	}

	if deps.HTTPClient == nil {
		return errors.New("httpClient is required")
	}

	return nil
}

type Worker struct {
	deps          Deps
	queue         *queue
	throttles     map[sources.SourceName]*sourceThrottle
	comicTrackers map[uuid.UUID]*comicTracker
	comicCond     sync.Cond
	cfg           Config
	comicMu       sync.Mutex
}

func New(cfg Config, deps Deps) (*Worker, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}

	if deps.Publisher == nil {
		deps.Publisher = noopPublisher{}
	}

	deps.Logger = deps.Logger.With("component", "chapters.download")

	worker := &Worker{
		cfg:           cfg,
		deps:          deps,
		queue:         newQueue(),
		throttles:     buildSourceThrottles(deps.Sources, cfg.RateLimit, cfg.SourceRateLimits),
		comicTrackers: make(map[uuid.UUID]*comicTracker),
	}
	worker.comicCond.L = &worker.comicMu

	return worker, nil
}

func (w *Worker) Enqueue(ctx context.Context, chapterList []chapters.Chapter) error {
	byComic := make(map[uuid.UUID][]chapters.Chapter)
	for _, chapter := range chapterList {
		byComic[chapter.ComicID] = append(byComic[chapter.ComicID], chapter)
	}

	for comicID, comicChapters := range byComic {
		comic, err := w.deps.ComicsRepository.FindByID(ctx, comicID)
		if err != nil {
			return fmt.Errorf("w.deps.ComicsRepository.FindByID: %w", err)
		}

		w.queue.enqueue(comic.Source, comicChapters)
	}

	return nil
}

func (w *Worker) ResetAndEnqueue(ctx context.Context, chapterID uuid.UUID) error {
	chapter, err := w.deps.ChaptersRepository.GetByID(ctx, chapterID)
	if err != nil {
		return fmt.Errorf("w.deps.ChaptersRepository.GetByID: %w", err)
	}

	dir := chapterDir(w.cfg.Dir, chapter.ComicID, chapter.Number)
	if err = deleteChapterDir(dir); err != nil {
		return fmt.Errorf("deleteChapterDir: %w", err)
	}

	if err = w.deps.ChaptersRepository.UpdateDownload(ctx, chapterID, 0); err != nil {
		return fmt.Errorf("w.deps.ChaptersRepository.UpdateDownload: %w", err)
	}

	return w.Enqueue(ctx, []chapters.Chapter{*chapter})
}

func (w *Worker) Resume(ctx context.Context, chapterID uuid.UUID) error {
	chapter, err := w.deps.ChaptersRepository.GetByID(ctx, chapterID)
	if err != nil {
		return fmt.Errorf("w.deps.ChaptersRepository.GetByID: %w", err)
	}

	return w.Enqueue(ctx, []chapters.Chapter{*chapter})
}

func (w *Worker) CleanupComic(ctx context.Context, comicID uuid.UUID, chapterList []chapters.Chapter) error {
	ids := make(map[uuid.UUID]struct{}, len(chapterList))
	for _, chapter := range chapterList {
		if chapter.ComicID != comicID {
			continue
		}

		ids[chapter.ID] = struct{}{}
	}

	w.queue.removeComicChapters(ids)
	w.cancelComicWork(comicID)
	w.waitComicWorkDone(comicID)

	dir := comicDir(w.cfg.Dir, comicID)
	if err := deleteComicDir(dir); err != nil {
		w.deps.Logger.ErrorContext(
			ctx,
			"failed to delete comic download dir",
			"comicID", comicID,
			loggingErr(err),
		)
	}

	return nil
}

func (w *Worker) Run(ctx context.Context) error {
	if err := utils.EnsureDir(w.cfg.Dir); err != nil {
		return fmt.Errorf("utils.EnsureDir: %w", err)
	}

	errG, errCtx := errgroup.WithContext(ctx)
	for source := range w.deps.Sources {
		errG.Go(func() error {
			return w.runSource(errCtx, source)
		})
	}

	w.deps.Logger.Info("worker started")

	//nolint:wrapcheck
	return errG.Wait()
}

func (w *Worker) runSource(ctx context.Context, source sources.SourceName) error {
	for {
		item, ok := w.queue.pop(ctx, source)
		if !ok {
			return nil
		}

		w.processChapter(ctx, source, item.ChapterID)
		w.queue.done(item.ChapterID)
	}
}

func (w *Worker) processChapter(ctx context.Context, source sources.SourceName, chapterID uuid.UUID) {
	chapter, err := w.deps.ChaptersRepository.GetByID(ctx, chapterID)
	if err != nil {
		w.deps.Logger.ErrorContext(ctx, "failed to load chapter", "chapterID", chapterID, loggingErr(err))

		return
	}

	comicCtx, releaseComic := w.beginComic(ctx, chapter.ComicID)
	defer releaseComic()

	if comicCtx.Err() != nil {
		return
	}

	if chapter.Download >= 100 {
		return
	}

	comic, err := w.deps.ComicsRepository.FindByID(comicCtx, chapter.ComicID)
	if err != nil {
		w.deps.Logger.ErrorContext(comicCtx, "failed to load comic", "chapterID", chapterID, loggingErr(err))

		return
	}

	src, ok := w.deps.Sources[source]
	if !ok {
		w.deps.Logger.ErrorContext(comicCtx, "source not configured", "source", source)

		return
	}

	pageURLs, err := src.GetPageURLsByChapter(comicCtx, sources.GetPageURLsByChapterOpts{
		SeriesSlug:  comic.Slug,
		ChapterSlug: chapter.SourceChapterSlug,
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}

		w.markError(comicCtx, chapterID, err)

		return
	}

	pagesNb := len(pageURLs)
	if pagesNb == 0 {
		w.markError(comicCtx, chapterID, errors.New("source returned no pages"))

		return
	}

	if pagesNb != chapter.PagesNb {
		if err = w.deps.ChaptersRepository.UpdatePagesNb(comicCtx, chapterID, pagesNb); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			w.deps.Logger.ErrorContext(comicCtx, "failed to update pages_nb", "chapterID", chapterID, loggingErr(err))

			return
		}

		chapter.PagesNb = pagesNb
	}

	dir := chapterDir(w.cfg.Dir, chapter.ComicID, chapter.Number)
	downloaded, err := listDownloadedPageIndices(dir)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}

		w.markError(comicCtx, chapterID, err)

		return
	}

	if len(downloaded) == pagesNb {
		w.markCompleted(comicCtx, chapterID)

		return
	}

	if err = w.publishProgress(comicCtx, chapterID, progressPercent(len(downloaded), pagesNb)); err != nil {
		return
	}

	throttle, ok := w.throttles[source]
	if !ok {
		w.deps.Logger.ErrorContext(comicCtx, "source throttle not configured", "source", source)
		w.markError(comicCtx, chapterID, errors.New("source throttle not configured"))

		return
	}

	var progressMu sync.Mutex
	downloadedCount := len(downloaded)

	missing := make([]int, 0, pagesNb-len(downloaded))
	for pageIndex := 1; pageIndex <= pagesNb; pageIndex++ {
		if _, ok := downloaded[pageIndex]; !ok {
			missing = append(missing, pageIndex)
		}
	}

	errG, errCtx := errgroup.WithContext(comicCtx)
	for _, pageIndex := range missing {
		errG.Go(func() error {
			imageURL := pageURLs[pageIndex-1]
			ext := pageExtension(imageURL)
			destPath := filepath.Join(dir, pageFilename(pageIndex, ext))

			if err := w.downloadPageWithRetries(errCtx, throttle, imageURL, destPath); err != nil {
				return err
			}

			progressMu.Lock()
			downloadedCount++
			current := downloadedCount
			progressMu.Unlock()

			return w.publishProgress(errCtx, chapterID, progressPercent(current, pagesNb))
		})
	}

	if err = errG.Wait(); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}

		w.markError(comicCtx, chapterID, err)

		return
	}

	w.markCompleted(comicCtx, chapterID)
}

func (w *Worker) downloadPageWithRetries(
	ctx context.Context,
	throttle *sourceThrottle,
	imageURL, destPath string,
) error {
	var lastErr error

	for attempt := 1; attempt <= defaultPageRetries; attempt++ {
		if err := throttle.wait(ctx); err != nil {
			return fmt.Errorf("throttle.wait: %w", err)
		}

		lastErr = downloadPage(ctx, w.deps.HTTPClient, imageURL, destPath)
		if lastErr == nil {
			return nil
		}

		w.deps.Logger.WarnContext(
			ctx,
			"page download failed",
			"url", imageURL,
			"attempt", attempt,
			loggingErr(lastErr),
		)
	}

	return fmt.Errorf("page download failed after %d attempts: %w", defaultPageRetries, lastErr)
}

func (w *Worker) publishProgress(ctx context.Context, chapterID uuid.UUID, download int) error {
	if err := w.deps.ChaptersRepository.UpdateDownload(ctx, chapterID, download); err != nil {
		w.deps.Logger.ErrorContext(ctx, "failed to update download progress", "chapterID", chapterID, loggingErr(err))

		return fmt.Errorf("w.deps.ChaptersRepository.UpdateDownload: %w", err)
	}

	w.deps.Publisher.Publish(ctx, chapterID, ProgressEvent{
		Status:   ProgressStatusDownloading,
		Download: download,
	})

	return nil
}

func (w *Worker) markCompleted(ctx context.Context, chapterID uuid.UUID) {
	if err := w.deps.ChaptersRepository.UpdateDownload(ctx, chapterID, 100); err != nil {
		w.deps.Logger.ErrorContext(ctx, "failed to mark chapter completed", "chapterID", chapterID, loggingErr(err))

		return
	}

	w.deps.Publisher.Publish(ctx, chapterID, ProgressEvent{
		Status:   ProgressStatusCompleted,
		Download: 100,
	})
}

func (w *Worker) markError(ctx context.Context, chapterID uuid.UUID, err error) {
	w.deps.Logger.ErrorContext(ctx, "chapter download failed", "chapterID", chapterID, loggingErr(err))

	if updateErr := w.deps.ChaptersRepository.UpdateDownload(ctx, chapterID, -1); updateErr != nil {
		w.deps.Logger.ErrorContext(ctx, "failed to mark chapter error", "chapterID", chapterID, loggingErr(updateErr))

		return
	}

	w.deps.Publisher.Publish(ctx, chapterID, ProgressEvent{
		Status:   ProgressStatusError,
		Download: -1,
	})
}

func loggingErr(err error) slog.Attr {
	return slog.Any("error", err)
}
