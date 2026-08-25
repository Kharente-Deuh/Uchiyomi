// SPDX-License-Identifier: AGPL-3.0-or-later

package domain

import (
	"context"
	"net/http"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/parse"
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
	Offset    int
	Limit     int
}

type SearchCacheResult struct {
	Items []SearchCacheResultItem
	Meta  SearchResultMeta
}

type SearchResultMeta struct {
	Total   int
	PerPage int
	HasMore bool
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

type SearchResultItemChapter struct {
	EarlyAccessUntil *time.Time
	PublishedAt      time.Time
	Title            string
	ID               string
	Number           float64
}

type GetImageURLsByChapterOpts struct {
	SeriesSlug string
	ChapterID  string
}
