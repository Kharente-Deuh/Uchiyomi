// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/parse"
)

func (c *Client) GetImageURLsByChapter(
	ctx context.Context,
	opts domain.GetImageURLsByChapterOpts,
) (*[]string, error) {
	targetURL := c.mangaURL("/" + opts.ChapterID + "/")

	status, body, err := c.get(ctx, targetURL)
	if err != nil {
		return nil, fmt.Errorf("c.get: %w", err)
	}

	if status == http.StatusNotFound {
		return nil, domain.ErrNotFound
	}

	urls := parse.ParsePageURLs(string(body))
	if urls == nil {
		return nil, fmt.Errorf("parse.ParsePageURLs: invalid reader HTML")
	}

	return &urls, nil
}
