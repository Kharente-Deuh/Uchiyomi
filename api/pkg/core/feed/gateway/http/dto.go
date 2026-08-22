// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"time"

	"github.com/google/uuid"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/feed"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
)

type latestChapterResponse struct {
	PublishedAt      time.Time  `json:"publishedAt"`
	EarlyAccessUntil *time.Time `json:"earlyAccessUntil,omitempty"`
	Title            string     `json:"title"`
	Number           float64    `json:"number"`
	HasProgress      bool       `json:"hasProgress"`
	Download         int        `json:"download"`
	ID               uuid.UUID  `json:"id"`
}

type itemResponse struct {
	Title          string                  `json:"title"`
	Slug           string                  `json:"slug"`
	Cover          string                  `json:"cover"`
	Source         sources.SourceName      `json:"source"`
	Status         sources.SeriesStatus    `json:"status"`
	Type           sources.SeriesType      `json:"type"`
	LatestChapters []latestChapterResponse `json:"latestChapters"`
	ID             uuid.UUID               `json:"id"`
}

type listResponse struct {
	Items []itemResponse `json:"items"`
	Total int64          `json:"total"`
}

func comicCoverURL(id uuid.UUID) string {
	return "/api/comics/" + id.String() + "/cover"
}

func listFromPage(page feed.Page) listResponse {
	items := make([]itemResponse, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, itemFromDomain(&page.Items[i]))
	}

	return listResponse{Items: items, Total: page.Total}
}

func itemFromDomain(item *feed.Item) itemResponse {
	chs := make([]latestChapterResponse, 0, len(item.LatestChapters))
	for i := range item.LatestChapters {
		ch := item.LatestChapters[i]
		chs = append(chs, latestChapterResponse{
			ID:               ch.ID,
			Title:            ch.Title,
			Number:           ch.Number,
			PublishedAt:      ch.PublishedAt,
			EarlyAccessUntil: utils.OptionalTime(ch.EarlyAccessUntil),
			Download:         ch.Download,
			HasProgress:      ch.HasProgress,
		})
	}

	return itemResponse{
		ID:             item.ID,
		Title:          item.Title,
		Slug:           item.Slug,
		Cover:          comicCoverURL(item.ID),
		Source:         item.Source,
		Status:         item.Status,
		Type:           item.Type,
		LatestChapters: chs,
	}
}
