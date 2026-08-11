// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"github.com/google/uuid"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

type lightComic struct {
	ID           uuid.UUID          `json:"id"`
	Title        string             `json:"title"`
	Slug         string             `json:"slug"`
	Source       sources.SourceName `json:"source"`
	Status       sources.SeriesStatus `json:"status"`
	ChapterCount int                  `json:"chapter_count"`
}

func lightComicFromDomain(comic *comics.Comic) lightComic {
	return lightComic{
		ID:           comic.ID,
		ChapterCount: comic.ChapterCount,
		Title:        comic.Title,
		Slug:         comic.Slug,
		Source:       comic.Source,
		Status:       comic.Status,
	}
}

type createRequest struct {
	Source sources.SourceName `json:"source" validate:"required"`
	Slug   string             `json:"slug" validate:"required"`
}

type comicResponse struct {
	ID           uuid.UUID          `json:"id"`
	Artist       string             `json:"artist"`
	Type         sources.SeriesType   `json:"type"`
	Description  string               `json:"description"`
	Source       sources.SourceName   `json:"source"`
	Author       string               `json:"author"`
	Status       sources.SeriesStatus `json:"status"`
	Slug         string               `json:"slug"`
	Title        string               `json:"title"`
	Genres       []string             `json:"genres"`
	AltTitles    []string             `json:"alt_titles"`
	ChapterCount int                  `json:"chapter_count"`
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
		Genres:       comic.Genres,
		AltTitles:    comic.AltTitles,
	}
}
