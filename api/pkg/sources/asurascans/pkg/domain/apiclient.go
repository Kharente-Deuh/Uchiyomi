// SPDX-License-Identifier: AGPL-3.0-or-later

package domain

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

type ApiClient interface {
	Search(context.Context, SearchCacheOpts) (*SearchCacheResult, error)
	GetInfosBySlug(context.Context, string) (*GetInfosBySlugResponse, error)
	GetChaptersListBySerie(context.Context, string) (*[]Chapter, error)
	GetImageURLsByChapter(context.Context, GetImageURLsByChapterOpts) (*[]string, error)
}

type SortType string

const (
	SortTypePopular SortType = "popular"
	SortTypeLatest  SortType = "latest"
	SortTypeRating  SortType = "rating"
	SortTypeTitle   SortType = "title"
	SortTypeNewest  SortType = "newest"
	SortTypeNone    SortType = ""
)

func IsSeriesSortType(s string) bool {
	values := []SortType{
		SortTypePopular,
		SortTypeLatest,
		SortTypeRating,
		SortTypeTitle,
		SortTypeNewest,
		SortTypeNone,
	}

	return slices.Contains(values, SortType(s))
}

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
	SortOrderNone SortOrder = ""
)

func IsSortOrder(s string) bool {
	values := []SortOrder{
		SortOrderAsc,
		SortOrderDesc,
		SortOrderNone,
	}

	return slices.Contains(values, SortOrder(s))
}

type SearchCacheOpts struct {
	Search      string
	Sort        SortType
	SortOrder   SortOrder
	Status      sources.SeriesStatus
	Type        sources.SeriesType
	Artist      string
	Genres      []string
	Offset      int
	Limit       int
	MinChapters int
}

type SearchOpts struct {
	Search      string
	Sort        SortType
	SortOrder   SortOrder
	Status      sources.SeriesStatus
	Type        sources.SeriesType
	Artist      string
	Genres      []string
	Offset      int
	Limit       int
	MinChapters int
	UserID      uuid.UUID
}

type SearchCacheResult struct {
	Items []SearchCacheResultItem
	Meta  SearchResultMeta
}

type SearchResult struct {
	Items []SearchResultItem
	Meta  SearchResultMeta
}

type SearchResultMeta struct {
	Total   int
	PerPage int
	HasMore bool
}

type SearchResultItem struct {
	LastChapterAt  time.Time
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

type SearchCacheResultItem struct {
	LastChapterAt  time.Time
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
		LastChapterAt:  i.LastChapterAt,
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
	EarlyAccessUntil time.Time
	PublishedAt      time.Time
	Title            string
	ID               string
	Number           float64
}

type GetInfosBySlugResponse struct {
	LastChapterAt time.Time
	UpdatedAt     time.Time
	CreatedAt     time.Time
	Description   string
	Title         string
	Cover         string
	Status        sources.SeriesStatus
	Type          sources.SeriesType
	Author        string
	Artist        string
	SourceURL     string
	PublicURL     string
	Slug          string
	AltTitles     []string
	Genres        []string
	ChapterCount  int
	Rating        float64
}

func (r *GetInfosBySlugResponse) Source(internalURL *uuid.UUID) sources.GetInfosBySlugResponse {
	return sources.GetInfosBySlugResponse{
		LastChapterAt: r.LastChapterAt,
		UpdatedAt:     r.UpdatedAt,
		CreatedAt:     r.CreatedAt,
		Description:   r.Description,
		Title:         r.Title,
		Cover:         r.Cover,
		Status:        r.Status,
		Type:          r.Type,
		Author:        r.Author,
		Artist:        r.Artist,
		SourceURL:     r.SourceURL,
		PublicURL:     r.PublicURL,
		Slug:          r.Slug,
		AltTitles:     r.AltTitles,
		Genres:        r.Genres,
		ChapterCount:  r.ChapterCount,
		Rating:        r.Rating,
		InternalID:    internalURL,
	}
}

type Chapter struct {
	EarlyAccessUntil time.Time
	PublishedAt      time.Time
	ID               string
	Title            string
	Number           float64
}

type GetImageURLsByChapterOpts struct {
	SeriesSlug string
	ChapterID  string
}
