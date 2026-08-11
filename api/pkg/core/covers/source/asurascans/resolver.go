// SPDX-License-Identifier: AGPL-3.0-or-later

package asurascans

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	asuradomain "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
)

const DefaultCDNBaseURL = "https://gg.asuracomic.net"

type InfosBySlugGetter interface {
	GetInfosBySlug(context.Context, string) (*sources.GetInfosBySlugResponse, error)
}

type Config struct {
	CDNBaseURL string
}

func (cfg *Config) Validate() error {
	if cfg.CDNBaseURL == "" {
		return errors.New("CDNBaseURL is required")
	}

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
	cfg  Config
}

func New(cfg Config, deps Deps) (*Resolver, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	deps.Logger = deps.Logger.With("component", "covers.source.asurascans")

	return &Resolver{cfg: cfg, deps: deps}, nil
}

func (r *Resolver) ResolveExternalURL(ctx context.Context, slug string) (string, error) {
	infos, err := r.deps.Getter.GetInfosBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, asuradomain.ErrNotFound) {
			return "", covers.ErrSeriesNotFound
		}

		return "", fmt.Errorf("getter.GetInfosBySlug: %w", err)
	}

	if infos.Cover == "" {
		return "", fmt.Errorf("series %q has no cover", slug)
	}

	return covers.ResolveAbsoluteURL(r.cfg.CDNBaseURL, infos.Cover), nil
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
