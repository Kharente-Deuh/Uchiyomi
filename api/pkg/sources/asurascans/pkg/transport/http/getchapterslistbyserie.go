// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
)

//nolint:lll
func (c *Client) GetChaptersListBySerie(ctx context.Context, seriesSlug string) (*[]domain.Chapter, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/series/%s/chapters", c.cfg.AsuraURL, seriesSlug), nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	res, err := c.deps.Http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("c.deps.http: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, domain.ErrNotFound
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("c.deps.http: %d", res.StatusCode)
	}

	var parsed getSeriesChaptersHttpResponse
	if err = json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("json.Decode: %w", err)
	}

	ch := parsed.Domain()

	return &ch, nil
}

type getSeriesChaptersHttpResponse struct {
	Chapters []getSeriesChaptersHttpChapter `json:"data"`
}

func (r *getSeriesChaptersHttpResponse) Domain() []domain.Chapter {
	chapters := make([]domain.Chapter, len(r.Chapters))
	for i, c := range r.Chapters {
		chapters[i] = domain.Chapter{
			ID:               c.Slug,
			Number:           c.Number,
			EarlyAccessUntil: utils.OptionalTime(c.EarlyAccessUntil),
			PublishedAt:      c.PublishedAt,
			Title:            c.Title,
			PageCount:        c.PageCount,
		}
	}

	return chapters
}

type getSeriesChaptersHttpChapter struct {
	EarlyAccessUntil *time.Time `json:"early_access_until,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	PublishedAt      time.Time  `json:"published_at"`
	Title            string     `json:"title,omitempty"`
	SeriesSlug       string     `json:"series_slug"`
	Slug             string     `json:"slug"`
	ViewCount        int        `json:"view_count"`
	ID               int        `json:"id"`
	PageCount        int        `json:"page_count"`
	Number           float64    `json:"number"`
	SeriesID         int        `json:"series_id"`
	CommentsEnabled  bool       `json:"comments_enabled"`
	IsPremium        bool       `json:"is_premium"`
	IsLocked         bool       `json:"is_locked"`
}
