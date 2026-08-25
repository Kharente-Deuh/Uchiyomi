// SPDX-License-Identifier: AGPL-3.0-or-later

package kingofshojo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	kosdomain "github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/domain"
)

type InfosBySlugGetter interface {
	GetInfosBySlug(context.Context, sources.GetInfosBySlugOpts) (*sources.GetInfosBySlugResponse, error)
}

type Config struct{}

func (cfg *Config) Validate() error {
	return nil
}

type Deps struct {
	Getter     InfosBySlugGetter
	HTTPClient *http.Client
	Logger     *slog.Logger
}

func (deps *Deps) Validate() error {
	if deps.Getter == nil {
		return errors.New("getter is required")
	}

	if deps.HTTPClient == nil {
		return errors.New("HTTPClient is required")
	}

	if deps.Logger == nil {
		return errors.New("logger is required")
	}

	return nil
}

type Resolver struct {
	deps Deps
}

func New(cfg Config, deps Deps) (*Resolver, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	deps.Logger = deps.Logger.With("component", "covers.source.kingofshojo")

	return &Resolver{deps: deps}, nil
}

func (r *Resolver) ResolveExternalURL(ctx context.Context, slug string) (string, error) {
	infos, err := r.deps.Getter.GetInfosBySlug(ctx, sources.GetInfosBySlugOpts{
		Slug:   slug,
		UserID: uuid.Nil,
	})
	if err != nil {
		if errors.Is(err, kosdomain.ErrNotFound) {
			return "", covers.ErrSeriesNotFound
		}

		return "", fmt.Errorf("getter.GetInfosBySlug: %w", err)
	}

	if infos.Cover == "" {
		return "", fmt.Errorf("series %q has no cover", slug)
	}

	if !strings.HasPrefix(infos.Cover, "http://") && !strings.HasPrefix(infos.Cover, "https://") {
		return "", fmt.Errorf("series %q has no absolute cover URL", slug)
	}

	return infos.Cover, nil
}

func (r *Resolver) Fetch(ctx context.Context, externalURL string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, externalURL, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	res, err := r.deps.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTPClient.Do: %w", err)
	}

	if res.StatusCode == http.StatusNotFound {
		res.Body.Close()

		return nil, covers.ErrSeriesNotFound
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		res.Body.Close()

		return nil, fmt.Errorf("unexpected status %d", res.StatusCode)
	}

	return res.Body, nil
}
