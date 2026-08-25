// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	coredomain "github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/parse"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/fncache"
	"golang.org/x/sync/errgroup"
)

var _ sources.Source = (*App)(nil)

type Config struct {
	SourceName sources.SourceName
	BaseURL    string
}

func (c *Config) Validate() error {
	_, err := sources.ParseSourceName(string(c.SourceName))
	if err != nil {
		return fmt.Errorf("sources.ParseSourceName: %w", err)
	}

	if c.BaseURL == "" {
		return errors.New("baseURL is required")
	}

	return nil
}

type Deps struct {
	Logger                *slog.Logger
	SearchCache           *fncache.Cache[domain.SearchCacheOpts, domain.SearchCacheResult]
	SeriesCache           *fncache.Cache[string, parse.SeriesPage]
	GetImageURLsByChapter *fncache.Cache[domain.GetImageURLsByChapterOpts, []string]
	ComicsRepository      comics.ComicsRepository
	ChaptersService       chapters.ChaptersService
}

func (deps *Deps) Validate() error {
	if deps.SearchCache == nil {
		return errors.New("searchCache is required")
	}

	if deps.SeriesCache == nil {
		return errors.New("seriesCache is required")
	}

	if deps.GetImageURLsByChapter == nil {
		return errors.New("getImageURLsByChapter is required")
	}

	if deps.ComicsRepository == nil {
		return errors.New("comicsRepository is required")
	}

	return nil
}

type App struct {
	deps Deps
	cfg  Config
}

func New(cfg Config, deps Deps) (*App, error) {
	var err error
	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err = deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	deps.Logger = deps.Logger.With("component", "sources.kingofshojo")

	app := &App{
		cfg:  cfg,
		deps: deps,
	}

	return app, nil
}

func (a *App) BindChaptersService(svc chapters.ChaptersService) {
	a.deps.ChaptersService = svc
}

func (a *App) Run(ctx context.Context) error {
	errG, errCtx := errgroup.WithContext(ctx)

	for _, run := range []func(context.Context) error{
		a.deps.SearchCache.Run,
		a.deps.SeriesCache.Run,
		a.deps.GetImageURLsByChapter.Run,
	} {
		errG.Go(func() error { return run(errCtx) })
	}

	a.deps.Logger.Info("app started")

	//nolint:wrapcheck
	return errG.Wait()
}

func (a *App) Search(ctx context.Context, opts domain.SearchOpts) (*domain.SearchResult, error) {
	res, err := a.deps.SearchCache.Get(ctx, domain.SearchCacheOpts{
		Search:    opts.Search,
		Sort:      opts.Sort,
		SortOrder: opts.SortOrder,
		Offset:    opts.Offset,
		Limit:     opts.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("a.deps.SearchCache.Get: %w", err)
	}

	slugs := make([]string, len(res.Items))
	for i, item := range res.Items {
		slugs[i] = item.Slug
	}

	foundComics, err := a.deps.ComicsRepository.GetBySlugsAndSource(ctx, comics.GetBySlugsAndSource{
		Source: a.cfg.SourceName,
		Slugs:  slugs,
		UserID: opts.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("a.deps.ComicsRepository.GetBySlugsAndSource: %w", err)
	}

	items := make([]domain.SearchResultItem, len(res.Items))
	for i, item := range res.Items {
		var internalID *uuid.UUID = nil

		for _, comic := range foundComics {
			if comic.Slug == item.Slug {
				internalID = &comic.ID

				break
			}
		}

		items[i] = item.Domain(internalID)
	}

	return &domain.SearchResult{
		Items: items,
		Meta:  res.Meta,
	}, nil
}

func (a *App) GetInfosBySlug(
	ctx context.Context,
	opts sources.GetInfosBySlugOpts,
) (*sources.GetInfosBySlugResponse, error) {
	page, err := a.seriesPage(ctx, opts.Slug, opts.Fresh)
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	res := seriesInfosResponse(page.Infos, page.Chapters, a.cfg.BaseURL)

	var internalID *uuid.UUID = nil
	if opts.UserID != uuid.Nil {
		comic, err := a.deps.ComicsRepository.GetBySourceSlug(ctx, comics.GetBySourceSlugOpts{
			Source: a.cfg.SourceName,
			Slug:   opts.Slug,
			UserID: opts.UserID,
		})
		if err != nil && !errors.Is(err, coredomain.ErrNotFound) {
			return nil, fmt.Errorf("a.deps.ComicsRepository.GetBySourceSlug: %w", err)
		}

		if comic != nil {
			internalID = &comic.ID
		}
	}

	res.InternalID = internalID

	return &res, nil
}

func (a *App) GetChaptersBySlug(
	ctx context.Context,
	opts sources.GetChaptersBySlugOpts,
) ([]sources.SourceChapter, error) {
	chs, err := a.chaptersListBySeries(ctx, opts.Slug, uuid.Nil, opts.Fresh)
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	result := make([]sources.SourceChapter, len(chs))
	for i, ch := range chs {
		result[i] = sources.SourceChapter{
			EarlyAccessUntil:  utils.OptionalTime(ch.EarlyAccessUntil),
			PublishedAt:       ch.PublishedAt,
			SourceChapterSlug: ch.ID,
			Title:             ch.Title,
			Number:            ch.Number,
			PageCount:         ch.PageCount,
		}
	}

	return result, nil
}

func (a *App) GetChaptersListBySeries(
	ctx context.Context,
	slug string,
	userID uuid.UUID,
) ([]domain.Chapter, error) {
	return a.chaptersListBySeries(ctx, slug, userID, false)
}

func (a *App) seriesPage(ctx context.Context, slug string, fresh bool) (*parse.SeriesPage, error) {
	var (
		page *parse.SeriesPage
		err  error
	)

	if fresh {
		page, err = a.deps.SeriesCache.Fetch(ctx, slug)
	} else {
		page, err = a.deps.SeriesCache.Get(ctx, slug)
	}
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	return page, nil
}

func (a *App) chaptersListBySeries(
	ctx context.Context,
	slug string,
	userID uuid.UUID,
	fresh bool,
) ([]domain.Chapter, error) {
	page, err := a.seriesPage(ctx, slug, fresh)
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	chapterList := slices.Clone(seriesChapters(page.Chapters))

	if userID == uuid.Nil || a.deps.ChaptersService == nil {
		return chapterList, nil
	}

	comic, err := a.deps.ComicsRepository.GetBySourceSlug(ctx, comics.GetBySourceSlugOpts{
		Source: a.cfg.SourceName,
		Slug:   slug,
		UserID: userID,
	})
	if err != nil && !errors.Is(err, coredomain.ErrNotFound) {
		return nil, fmt.Errorf("a.deps.ComicsRepository.GetBySourceSlug: %w", err)
	}

	if comic == nil {
		return chapterList, nil
	}

	dbChapters, err := a.deps.ChaptersService.ListByComicID(ctx, comic.ID)
	if err != nil {
		return nil, fmt.Errorf("a.deps.ChaptersService.ListByComicID: %w", err)
	}

	bySlug := make(map[string]chapters.Chapter, len(dbChapters))
	for _, chapter := range dbChapters {
		bySlug[chapter.SourceChapterSlug] = chapter
	}

	for i := range chapterList {
		dbChapter, ok := bySlug[chapterList[i].ID]
		if !ok {
			continue
		}

		internalID := dbChapter.ID
		download := dbChapter.Download
		chapterList[i].InternalID = &internalID
		chapterList[i].Download = &download
	}

	return chapterList, nil
}

func (a *App) GetImageURLsByChapter(ctx context.Context, opts domain.GetImageURLsByChapterOpts) ([]string, error) {
	res, err := a.deps.GetImageURLsByChapter.Get(ctx, opts)
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	return slices.Clone(*res), nil
}

func (a *App) GetPageURLsByChapter(ctx context.Context, opts sources.GetPageURLsByChapterOpts) ([]string, error) {
	urls, err := a.GetImageURLsByChapter(ctx, domain.GetImageURLsByChapterOpts{
		SeriesSlug: opts.SeriesSlug,
		ChapterID:  opts.ChapterSlug,
	})
	if err != nil {
		return nil, fmt.Errorf("a.GetImageURLsByChapter: %w", err)
	}

	return urls, nil
}

func seriesInfosResponse(
	infos parse.SeriesInfos,
	chapters []parse.SeriesChapter,
	baseURL string,
) sources.GetInfosBySlugResponse {
	seriesURL := baseURL + "/manga/" + infos.Slug + "/"

	return sources.GetInfosBySlugResponse{
		LastChapterAt: maxChapterPublishedAt(chapters),
		UpdatedAt:     infos.UpdatedAt,
		CreatedAt:     infos.CreatedAt,
		Description:   infos.Description,
		Title:         infos.Title,
		Cover:         infos.Cover,
		Status:        infos.Status,
		Type:          infos.Type,
		Author:        infos.Author,
		Artist:        infos.Artist,
		SourceURL:     seriesURL,
		PublicURL:     seriesURL,
		Slug:          infos.Slug,
		AltTitles:     infos.AltTitles,
		Genres:        infos.Genres,
		ChapterCount:  infos.ChapterCount,
		Rating:        0,
	}
}

func seriesChapters(chapters []parse.SeriesChapter) []domain.Chapter {
	result := make([]domain.Chapter, len(chapters))
	for i, ch := range chapters {
		result[i] = domain.Chapter{
			PublishedAt: ch.PublishedAt,
			ID:          ch.ID,
			Title:       ch.Title,
			Number:      ch.Number,
			PageCount:   ch.PageCount,
		}
	}

	return result
}

func maxChapterPublishedAt(chapters []parse.SeriesChapter) *time.Time {
	if len(chapters) == 0 {
		return nil
	}

	maxNumber := chapters[0].Number
	maxAt := chapters[0].PublishedAt

	for _, ch := range chapters[1:] {
		if ch.Number > maxNumber {
			maxNumber = ch.Number
			maxAt = ch.PublishedAt
		}
	}

	if maxAt.IsZero() {
		return nil
	}

	t := maxAt

	return &t
}
