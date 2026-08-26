// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"github.com/google/uuid"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

type lightComic struct {
	Title        string               `json:"title"`
	Slug         string               `json:"slug"`
	Cover        string               `json:"cover"`
	Source       sources.SourceName   `json:"source"`
	Status       sources.SeriesStatus `json:"status"`
	Type         sources.SeriesType   `json:"type"`
	ChapterCount int                  `json:"chapterCount"`
	ID           uuid.UUID            `json:"id"`
}

func comicCoverURL(id uuid.UUID) string {
	return "/api/comics/" + id.String() + "/cover"
}

func lightComicFromDomain(comic *comics.Comic) lightComic {
	return lightComic{
		ID:           comic.ID,
		ChapterCount: comic.ChapterCount,
		Title:        comic.Title,
		Slug:         comic.Slug,
		Cover:        comicCoverURL(comic.ID),
		Source:       comic.Source,
		Status:       comic.Status,
		Type:         comic.Type,
	}
}

type comicListResponse struct {
	Items []lightComic `json:"items"`
	Total int64        `json:"total"`
}

func comicListFromPage(page comics.Page) comicListResponse {
	items := make([]lightComic, len(page.Items))
	for i := range page.Items {
		items[i] = lightComicFromDomain(&page.Items[i])
	}

	return comicListResponse{Items: items, Total: page.Total}
}

type createRequest struct {
	Source sources.SourceName `json:"source" validate:"required"`
	Slug   string             `json:"slug" validate:"required"`
}

type updateComicRequest struct {
	Type sources.SeriesType `json:"type"`
}

type comicResponse struct {
	Artist       string               `json:"artist"`
	Type         sources.SeriesType   `json:"type"`
	Description  string               `json:"description"`
	Source       sources.SourceName   `json:"source"`
	Author       string               `json:"author"`
	Status       sources.SeriesStatus `json:"status"`
	Slug         string               `json:"slug"`
	Title        string               `json:"title"`
	Cover        string               `json:"cover"`
	Genres       []string             `json:"genres"`
	AltTitles    []string             `json:"altTitles"`
	ChapterCount int                  `json:"chapterCount"`
	ID           uuid.UUID            `json:"id"`
}

func comicResponseFromDomain(comic *comics.Comic) comicResponse {
	return comicResponse{
		ID:           comic.ID,
		Artist:       comic.Artist,
		ChapterCount: comic.ChapterCount,
		Type:         comic.Type,
		Description:  comic.Description,
		Source:       comic.Source,
		Author:       comic.Author,
		Status:       comic.Status,
		Slug:         comic.Slug,
		Title:        comic.Title,
		Cover:        comicCoverURL(comic.ID),
		Genres:       comic.Genres,
		AltTitles:    comic.AltTitles,
	}
}

type retryChaptersRequest struct {
	ChapterIDs []uuid.UUID `json:"chapterIds"`
}
