// SPDX-License-Identifier: AGPL-3.0-or-later

package covers

import (
	"context"
	"errors"
	"fmt"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/imgcache"
)

type AppDeps struct {
	Cache *imgcache.Cache
}

func (deps *AppDeps) Validate() error {
	if deps.Cache == nil {
		return errors.New("cache is required")
	}

	return nil
}

type App struct {
	deps AppDeps
}

func NewApp(deps AppDeps) (*App, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	return &App{deps: deps}, nil
}

func (a *App) Run(ctx context.Context) error {
	//nolint:wrapcheck
	return a.deps.Cache.Run(ctx)
}
