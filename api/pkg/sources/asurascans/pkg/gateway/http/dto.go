// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
)

type searchResDTO struct {
	Items []searchResItemDTO `json:"items"`
	Total int                `json:"total"`
}

type searchResItemDTO struct {
	LastChapterAt  time.Time                 `json:"lastChapterAt"`
	UpdatedAt      time.Time                 `json:"updatedAt"`
	CreatedAt      time.Time                 `json:"createdAt"`
	InternalID     *uuid.UUID                `json:"internalId"`
	Description    string                    `json:"description"`
	Slug           string                    `json:"slug"`
	Status         sources.SeriesStatus      `json:"status"`
	Type           sources.SeriesType        `json:"type"`
	Author         string                    `json:"author"`
	Artist         string                    `json:"artist"`
	SourceURL      string                    `json:"sourceUrl"`
	Cover          string                    `json:"cover"`
	Title          string                    `json:"title"`
	PublicURL      string                    `json:"publicUrl"`
	Genres         []string                  `json:"genres"`
	LatestChapters []searchResItemChapterDTO `json:"latestChapters"`
	AltTitles      []string                  `json:"altTitles"`
	ChapterCount   int                       `json:"chapterCount"`
	Rating         float64                   `json:"rating"`
	ReleaseYear    int                       `json:"releaseYear"`
}

type searchResItemChapterDTO struct {
	EarlyAccessUntil time.Time `json:"earlyAccessUntil"`
	PublishedAt      time.Time `json:"publishedAt"`
	Title            string    `json:"title"`
	ID               string    `json:"id"`
	Number           float64   `json:"number"`
}

func parseSearchOpts(q url.Values) (domain.SearchOpts, error) {
	opts := domain.SearchOpts{
		Search: strings.TrimSpace(q.Get("search")),
		Artist: strings.TrimSpace(q.Get("artist")),
	}

	sort := q.Get("sort")
	if !domain.IsSeriesSortType(sort) {
		return domain.SearchOpts{}, fmt.Errorf("invalid sort %q", sort)
	}

	opts.Sort = domain.SortType(sort)

	order := q.Get("order")
	if !domain.IsSortOrder(order) {
		return domain.SearchOpts{}, fmt.Errorf("invalid order %q", order)
	}

	opts.SortOrder = domain.SortOrder(order)

	status := q.Get("status")
	if status == "" {
		opts.Status = ""
	} else {
		parsedStatus, err := sources.ParseSeriesStatus(status)
		if err != nil {
			return domain.SearchOpts{}, fmt.Errorf("invalid status %q: %w", status, err)
		}

		opts.Status = parsedStatus
	}

	seriesType := q.Get("type")
	if seriesType == "" {
		opts.Type = ""
	} else {
		parsedType, err := sources.ParseSeriesType(seriesType)
		if err != nil {
			return domain.SearchOpts{}, fmt.Errorf("invalid type %q: %w", seriesType, err)
		}

		opts.Type = parsedType
	}

	genres, err := parseGenres(q.Get("genres"))
	if err != nil {
		return domain.SearchOpts{}, err
	}

	opts.Genres = genres

	for _, p := range []struct {
		dst *int
		key string
	}{
		{key: "offset", dst: &opts.Offset},
		{key: "limit", dst: &opts.Limit},
		{key: "min_chapters", dst: &opts.MinChapters},
	} {
		if *p.dst, err = parsePositiveInt(q.Get(p.key)); err != nil {
			return domain.SearchOpts{}, fmt.Errorf("invalid %s: %w", p.key, err)
		}
	}

	if opts.Limit > 100 {
		opts.Limit = 100
	}

	if opts.Limit <= 0 {
		opts.Limit = 20
	}

	if opts.Offset < 0 {
		opts.Offset = 0
	}

	if opts.MinChapters < 0 {
		opts.MinChapters = 0
	}

	return opts, nil
}

func parseGenres(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	genres := make([]string, 0, len(parts))

	for _, part := range parts {
		genre := strings.TrimSpace(part)
		if genre == "" {
			return nil, fmt.Errorf("genre is empty")
		}

		genres = append(genres, genre)
	}

	return genres, nil
}

func parsePositiveInt(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}

	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%q is not a number", raw)
	}

	if n < 0 {
		return 0, fmt.Errorf("%d is negative", n)
	}

	return n, nil
}

type getInfosBySlugResDTO struct {
	LastChapterAt time.Time            `json:"lastChapterAt"`
	UpdatedAt     time.Time            `json:"updatedAt"`
	CreatedAt     time.Time            `json:"createdAt"`
	InternalID    *uuid.UUID           `json:"internalId"`
	Author        string               `json:"author"`
	PublicURL     string               `json:"publicUrl"`
	Status        sources.SeriesStatus `json:"status"`
	Type          sources.SeriesType   `json:"type"`
	Title         string               `json:"title"`
	Artist        string               `json:"artist"`
	SourceURL     string               `json:"sourceUrl"`
	Cover         string               `json:"cover"`
	Slug          string               `json:"slug"`
	Description   string               `json:"description"`
	Genres        []string             `json:"genres"`
	AltTitles     []string             `json:"altTitles"`
	ChapterCount  int                  `json:"chapterCount"`
	Rating        float64              `json:"rating"`
}

type getChaptersListBySeriesResItemDTO struct {
	EarlyAccessUntil time.Time `json:"earlyAccessUntil"`
	PublishedAt      time.Time `json:"publishedAt"`
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Number           float64   `json:"number"`
}

type getImageURLsByChapterResDTO struct {
	URLs []string `json:"urls"`
}
