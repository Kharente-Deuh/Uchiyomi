// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/fncache"
	"golang.org/x/sync/errgroup"
)

type Dependencies struct {
	Logger                       *slog.Logger
	SearchCache                  *fncache.Cache[domain.SearchOpts, domain.SearchResult]
	GetInfosBySlugCache          *fncache.Cache[string, domain.GetInfosBySlugResponse]
	GetChaptersListBySeriesCache *fncache.Cache[string, []domain.Chapter]
	GetImageURLsByChapter        *fncache.Cache[domain.GetImageURLsByChapterOpts, []string]
}

func (deps *Dependencies) Validate() error {
	if deps.SearchCache == nil {
		return errors.New("searchCache is required")
	}

	if deps.GetInfosBySlugCache == nil {
		return errors.New("getInfosBySlugCache is required")
	}

	if deps.GetChaptersListBySeriesCache == nil {
		return errors.New("getChaptersListBySeriesCache is required")
	}

	if deps.GetImageURLsByChapter == nil {
		return errors.New("getImageURLsByChapter is required")
	}

	return nil
}

type App struct {
	deps Dependencies
}

func New(deps Dependencies) (*App, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	deps.Logger = deps.Logger.With("component", "sources.asurascans")

	app := &App{
		deps: deps,
	}

	return app, nil
}

func (a *App) Run(ctx context.Context) error {
	errG, errCtx := errgroup.WithContext(ctx)

	for _, run := range []func(context.Context) error{
		a.deps.SearchCache.Run,
		a.deps.GetInfosBySlugCache.Run,
		a.deps.GetChaptersListBySeriesCache.Run,
		a.deps.GetImageURLsByChapter.Run,
	} {
		errG.Go(func() error { return run(errCtx) })
	}

	a.deps.Logger.Info("app started")

	//nolint:wrapcheck
	return errG.Wait()
}

func (a *App) Search(ctx context.Context, opts domain.SearchOpts) (*domain.SearchResult, error) {
	//nolint:wrapcheck
	return a.deps.SearchCache.Get(ctx, opts)
}

func (a *App) GetInfosBySlug(ctx context.Context, slug string) (*domain.GetInfosBySlugResponse, error) {
	//nolint:wrapcheck
	return a.deps.GetInfosBySlugCache.Get(ctx, slug)
}

func (a *App) GetChaptersListBySeries(ctx context.Context, slug string) ([]domain.Chapter, error) {
	res, err := a.deps.GetChaptersListBySeriesCache.Get(ctx, slug)
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	return slices.Clone(*res), nil
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
