// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

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
	PublicURL      string                    `json:"publicUrl"`
	SourceURL      string                    `json:"sourceUrl"`
	Cover          string                    `json:"cover"`
	Status         domain.SeriesStatus       `json:"status"`
	Type           domain.SeriesType         `json:"type"`
	Author         string                    `json:"author"`
	Artist         string                    `json:"artist"`
	Description    string                    `json:"description"`
	Slug           string                    `json:"slug"`
	Title          string                    `json:"title"`
	AltTitles      []string                  `json:"altTitles"`
	Genres         []domain.SeriesGenre      `json:"genres"`
	LatestChapters []searchResItemChapterDTO `json:"latestChapters"`
	ChapterCount   int                       `json:"chapterCount"`
	Rating         float64                   `json:"rating"`
	ReleaseYear    int                       `json:"releaseYear"`
}

type searchResItemChapterDTO struct {
	EarlyAccessUntil time.Time `json:"earlyAccessUntil"`
	PublishedAt      time.Time `json:"publishedAt"`
	Title            string    `json:"title"`
	ID               string    `json:"id"`
	Number           int       `json:"number"`
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
	if !domain.IsSeriesStatus(status) {
		return domain.SearchOpts{}, fmt.Errorf("invalid status %q", status)
	}

	opts.Status = domain.SeriesStatus(status)

	seriesType := q.Get("type")
	if !domain.IsSeriesType(seriesType) {
		return domain.SearchOpts{}, fmt.Errorf("invalid type %q", seriesType)
	}

	opts.Type = domain.SeriesType(seriesType)

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

func parseGenres(raw string) ([]domain.SeriesGenre, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	genres := make([]domain.SeriesGenre, 0, len(parts))

	for _, part := range parts {
		genre := strings.TrimSpace(part)
		if genre == "" {
			continue
		}

		if !domain.IsSeriesGenre(genre) {
			return nil, fmt.Errorf("invalid genre %q", genre)
		}

		genres = append(genres, domain.SeriesGenre(genre))
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
	Description   string               `json:"description"`
	Title         string               `json:"title"`
	Cover         string               `json:"cover"`
	Status        domain.SeriesStatus  `json:"status"`
	Type          domain.SeriesType    `json:"type"`
	Author        string               `json:"author"`
	Artist        string               `json:"artist"`
	SourceURL     string               `json:"sourceUrl"`
	PublicURL     string               `json:"publicUrl"`
	Slug          string               `json:"slug"`
	AltTitles     []string             `json:"altTitles"`
	Genres        []domain.SeriesGenre `json:"genres"`
	ChapterCount  int                  `json:"chapterCount"`
	Rating        float64              `json:"rating"`
}

type getChaptersListBySeriesResItemDTO struct {
	EarlyAccessUntil time.Time `json:"earlyAccessUntil"`
	PublishedAt      time.Time `json:"publishedAt"`
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Number           int       `json:"number"`
}

type getImageURLsByChapterResDTO struct {
	URLs []string `json:"urls"`
}
