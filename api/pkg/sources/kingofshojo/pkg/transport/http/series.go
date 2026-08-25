// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/parse"
)

func (c *Client) GetSeriesPage(ctx context.Context, slug string) (*parse.SeriesPage, error) {
	targetURL := c.mangaURL("/manga/" + slug + "/")

	status, body, err := c.get(ctx, targetURL)
	if err != nil {
		return nil, fmt.Errorf("c.get: %w", err)
	}

	if status == http.StatusNotFound {
		return nil, domain.ErrNotFound
	}

	page, err := parse.ParseSeries(string(body), slug)
	if err != nil {
		return nil, fmt.Errorf("parse.ParseSeries: %w", err)
	}

	return &page, nil
}
