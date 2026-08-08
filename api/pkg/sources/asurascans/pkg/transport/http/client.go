// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
)

var _ domain.ApiClient = (*Client)(nil)

type Config struct {
	AsuraURL string
}

func (cfg *Config) Validate() error {
	_, err := url.ParseRequestURI(cfg.AsuraURL)
	if err != nil {
		return errors.New("asuraURL is not a valid URL")
	}

	return nil
}

type Deps struct {
	Http *http.Client
}

func (d *Deps) Validate() error {
	if d.Http == nil {
		return errors.New("SearchCache is required")
	}

	return nil
}

type Client struct {
	deps Deps
	cfg  Config
}

func New(deps Deps, cfg Config) (*Client, error) {
	var err error
	if err = deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	c := &Client{
		deps: deps,
		cfg:  cfg,
	}

	return c, nil
}
