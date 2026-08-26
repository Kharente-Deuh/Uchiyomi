// SPDX-License-Identifier: AGPL-3.0-or-later

package domain

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/parse"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/challengesolver"
)

type Solver interface {
	Session(context.Context, string, ...challengesolver.Request) (*http.Client, *challengesolver.Solution, error)
}

type ApiClient interface {
	Search(context.Context, SearchCacheOpts) (*SearchCacheResult, error)
	GetSeriesPage(context.Context, string) (*parse.SeriesPage, error)
	GetImageURLsByChapter(context.Context, GetImageURLsByChapterOpts) (*[]string, error)
}

type SortType string

const (
	SortTypePopular SortType = "popular"
	SortTypeLatest  SortType = "latest"
	SortTypeNewest  SortType = "newest"
	SortTypeTitle   SortType = "title"
	SortTypeNone    SortType = ""
)

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
	SortOrderNone SortOrder = ""
)

type SearchCacheOpts struct {
	Search    string
	Sort      SortType
	SortOrder SortOrder
	Page      int
}

type SearchOpts struct {
	Search    string
	Sort      SortType
	SortOrder SortOrder
	Page      int
	UserID    uuid.UUID
}

type SearchResult struct {
	Items []SearchResultItem
	Meta  SearchResultMeta
}

type SearchResultItem struct {
	LastChapterAt  *time.Time
	UpdatedAt      time.Time
	CreatedAt      time.Time
	InternalID     *uuid.UUID
	Description    string
	PublicURL      string
	Status         sources.SeriesStatus
	Type           sources.SeriesType
	Author         string
	Artist         string
	SourceURL      string
	Slug           string
	Title          string
	Cover          string
	AltTitles      []string
	LatestChapters []SearchResultItemChapter
	Genres         []string
	ChapterCount   int
	ID             int
	Rating         float64
	ReleaseYear    int
}

type SearchCacheResult struct {
	Items []SearchCacheResultItem
	Meta  SearchResultMeta
}

type SearchResultMeta struct {
	HasNextPage bool
}

type SearchCacheResultItem struct {
	LastChapterAt  *time.Time
	UpdatedAt      time.Time
	CreatedAt      time.Time
	PublicURL      string
	SourceURL      string
	Cover          string
	Status         sources.SeriesStatus
	Type           sources.SeriesType
	Author         string
	Artist         string
	Description    string
	Slug           string
	Title          string
	AltTitles      []string
	Genres         []string
	LatestChapters []SearchResultItemChapter
	ChapterCount   int
	ID             int
	Rating         float64
	ReleaseYear    int
}

func (i *SearchCacheResultItem) Domain(internalID *uuid.UUID) SearchResultItem {
	return SearchResultItem{
		LastChapterAt:  utils.OptionalTime(i.LastChapterAt),
		UpdatedAt:      i.UpdatedAt,
		CreatedAt:      i.CreatedAt,
		PublicURL:      i.PublicURL,
		SourceURL:      i.SourceURL,
		Cover:          i.Cover,
		Status:         i.Status,
		Type:           i.Type,
		Author:         i.Author,
		Artist:         i.Artist,
		Description:    i.Description,
		Slug:           i.Slug,
		Title:          i.Title,
		AltTitles:      i.AltTitles,
		Genres:         i.Genres,
		LatestChapters: i.LatestChapters,
		ChapterCount:   i.ChapterCount,
		ID:             i.ID,
		Rating:         i.Rating,
		ReleaseYear:    i.ReleaseYear,
		InternalID:     internalID,
	}
}

type SearchResultItemChapter struct {
	EarlyAccessUntil *time.Time
	PublishedAt      time.Time
	Title            string
	ID               string
	Number           float64
}

type Chapter struct {
	EarlyAccessUntil *time.Time
	PublishedAt      time.Time
	InternalID       *uuid.UUID
	Download         *int
	ID               string
	Title            string
	Number           float64
	PageCount        int
}

type GetImageURLsByChapterOpts struct {
	SeriesSlug string
	ChapterID  string
}
