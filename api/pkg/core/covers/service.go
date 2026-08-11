// SPDX-License-Identifier: AGPL-3.0-or-later

package covers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/imgcache"
)

type ServiceConfig struct {
	ProxyPathPrefix string
}

func (cfg *ServiceConfig) Validate() error {
	if cfg.ProxyPathPrefix == "" {
		return errors.New("proxyPathPrefix is required")
	}

	return nil
}

type ServiceDeps struct {
	Cache      *imgcache.Cache
	Resolvers  map[string]CoverResolver
	HTTPClient *http.Client
	Logger     *slog.Logger
}

func (deps *ServiceDeps) Validate() error {
	if deps.Cache == nil {
		return errors.New("cache is required")
	}

	if len(deps.Resolvers) == 0 {
		return errors.New("resolvers is required")
	}

	if deps.HTTPClient == nil {
		return errors.New("HTTPClient is required")
	}

	if deps.Logger == nil {
		return errors.New("logger is required")
	}

	return nil
}

type Service struct {
	deps ServiceDeps
	cfg  ServiceConfig
}

func NewService(cfg ServiceConfig, deps ServiceDeps) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	deps.Logger = deps.Logger.With("component", "covers.service")

	return &Service{cfg: cfg, deps: deps}, nil
}

func (s *Service) BuildProxyURL(source, slug string) string {
	return fmt.Sprintf("%s/%s?source=%s", s.cfg.ProxyPathPrefix, slug, source)
}

func (s *Service) Serve(ctx context.Context, source, slug string) (string, string, error) {
	resolver, ok := s.deps.Resolvers[source]
	if !ok {
		return "", "", ErrUnknownSource
	}

	externalURL, err := resolver.ResolveExternalURL(ctx, slug)
	if err != nil {
		if errors.Is(err, ErrSeriesNotFound) {
			return "", "", ErrSeriesNotFound
		}

		return "", "", fmt.Errorf("resolver.ResolveExternalURL: %w", err)
	}

	ext := ExtensionFromURL(externalURL)
	if ext == "" {
		ext, err = probeExtension(ctx, s.deps.HTTPClient, externalURL)
		if err != nil {
			if errors.Is(err, ErrSeriesNotFound) {
				return "", "", ErrSeriesNotFound
			}

			return "", "", fmt.Errorf("%w: %w", ErrDownloadFailed, err)
		}
	}

	key := cacheKey(source, slug, ext)

	diskPath, err := s.deps.Cache.Ensure(ctx, key)
	if err != nil {
		if errors.Is(err, ErrSeriesNotFound) {
			return "", "", ErrSeriesNotFound
		}

		return "", "", fmt.Errorf("%w: %w", ErrDownloadFailed, err)
	}

	contentType := MIMEForExtension(filepath.Ext(diskPath))

	return diskPath, contentType, nil
}

func NewFetchFn(resolvers map[string]CoverResolver) func(context.Context, string) (io.ReadCloser, error) {
	return func(ctx context.Context, key string) (io.ReadCloser, error) {
		source, slug, _, err := parseCacheKey(key)
		if err != nil {
			return nil, fmt.Errorf("parseCacheKey: %w", err)
		}

		resolver, ok := resolvers[source]
		if !ok {
			return nil, ErrUnknownSource
		}

		externalURL, err := resolver.ResolveExternalURL(ctx, slug)
		if err != nil {
			return nil, fmt.Errorf("resolver.ResolveExternalURL: %w", err)
		}

		rc, err := resolver.Fetch(ctx, externalURL)
		if err != nil {
			if errors.Is(err, ErrSeriesNotFound) {
				return nil, ErrSeriesNotFound
			}

			return nil, fmt.Errorf("%w: %w", ErrDownloadFailed, err)
		}

		return rc, nil
	}
}
