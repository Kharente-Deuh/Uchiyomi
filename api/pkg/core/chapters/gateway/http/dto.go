// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readingprogress"
)

type postListBody struct {
	IDs []uuid.UUID `json:"ids" validate:"required,min=1,dive,required"`
}

type chapterProgressResponse struct {
	UpdatedAt time.Time `json:"updatedAt"`
	Page      int       `json:"page"`
}

type postListResponseChapter struct {
	PublishedAt       time.Time                `json:"publishedAt"`
	EarlyAccessUntil  *time.Time               `json:"earlyAccessUntil"`
	Progress          *chapterProgressResponse `json:"progress"`
	SourceChapterSlug string                   `json:"sourceChapterSlug"`
	Title             string                   `json:"title"`
	Number            float64                  `json:"number"`
	PagesNb           int                      `json:"pagesNb"`
	Download          int                      `json:"download"`
	ID                uuid.UUID                `json:"id"`
	ComicID           uuid.UUID                `json:"comicId"`
}

func progressPayload(
	chapterID uuid.UUID,
	pagesNb int,
	byID map[uuid.UUID]readingprogress.Progress,
) *chapterProgressResponse {
	p, ok := byID[chapterID]
	if !ok {
		return nil
	}

	return &chapterProgressResponse{
		UpdatedAt: p.UpdatedAt,
		Page:      readingprogress.ClampPage(p.Page, pagesNb),
	}
}
