// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"strconv"
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

type chapterDetailResponse struct {
	NextChapterID     *uuid.UUID `json:"nextChapterId,omitempty"`
	PreviousChapterID *uuid.UUID `json:"previousChapterId,omitempty"`
	PageURLs          []string   `json:"pageUrls"`
	postListResponseChapter
}

func pageURLs(id uuid.UUID, download, pagesNb int) []string {
	if download != 100 || pagesNb <= 0 {
		return []string{}
	}

	urls := make([]string, pagesNb)
	for i := range pagesNb {
		urls[i] = "/api/chapters/" + id.String() + "/pages/" + strconv.Itoa(i+1)
	}

	return urls
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
