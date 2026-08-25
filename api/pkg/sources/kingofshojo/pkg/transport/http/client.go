// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/domain"
)

var _ domain.ApiClient = (*Client)(nil)

const defaultBaseURL = "https://kingofshojo.com"

type Config struct {
	BaseURL string
}

func (cfg *Config) Validate() error {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	_, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return errors.New("baseURL is not a valid URL")
	}

	cfg.BaseURL = strings.TrimRight(baseURL, "/")

	return nil
}

type Deps struct {
	HTTP   *http.Client
	Logger *slog.Logger
	Solver domain.Solver
}

func (d *Deps) Validate() error {
	if d.HTTP == nil {
		return errors.New("http is required")
	}

	if d.Logger == nil {
		return errors.New("logger is required")
	}

	return nil
}

type Client struct {
	deps        Deps
	retryClient *http.Client
	cfg         Config
	mu          sync.Mutex
}

func New(cfg Config, deps Deps) (*Client, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	deps.Logger = deps.Logger.With("component", "sources.kingofshojo.httpclient")

	return &Client{
		deps: deps,
		cfg:  cfg,
	}, nil
}

func (c *Client) httpClient() *http.Client {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.retryClient != nil {
		return c.retryClient
	}

	return c.deps.HTTP
}

func (c *Client) setRetryClient(hc *http.Client) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.retryClient = hc
}

func (c *Client) mangaURL(path string) string {
	return c.cfg.BaseURL + path
}
