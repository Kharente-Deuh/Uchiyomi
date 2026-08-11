// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
)

func (c *Client) Search(ctx context.Context, opts domain.SearchCacheOpts) (*domain.SearchCacheResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/series", c.cfg.AsuraURL), nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	req.URL.RawQuery = c.builSearchQuery(opts)

	fmt.Printf("[INFO] GET %s\n", req.URL.String())
	res, err := c.deps.Http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("c.deps.http: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, domain.ErrNotFound
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("c.deps.http: %d", res.StatusCode)
	}

	var parsedResult searchHTTPResponse
	if err = json.NewDecoder(res.Body).Decode(&parsedResult); err != nil {
		return nil, fmt.Errorf("json.Decode: %w", err)
	}

	return parsedResult.Domain(), nil
}

func (c *Client) builSearchQuery(opts domain.SearchCacheOpts) string {
	withDefaults := c.getSearchOptsWithDefaults(opts)
	q := url.Values{}
	q.Add("offset", strconv.Itoa(withDefaults.Offset))
	q.Add("limit", strconv.Itoa(withDefaults.Limit))
	q.Add("sort", string(withDefaults.Sort))
	q.Add("order", string(withDefaults.SortOrder))

	if withDefaults.Search != "" {
		q.Add("search", withDefaults.Search)
	}

	if withDefaults.Status != "" {
		q.Add("status", string(withDefaults.Status))
	}

	if withDefaults.Type != "" {
		q.Add("type", string(withDefaults.Type))
	}

	if withDefaults.Artist != "" {
		q.Add("artist", withDefaults.Artist)
	}

	if len(withDefaults.Genres) > 0 {
		q.Add("genres", strings.Join(utils.MapSlice(
			withDefaults.Genres,
			func(g string) string {
				return string(g)
			}),
			","))
	}

	return q.Encode()
}

func (c *Client) getSearchOptsWithDefaults(opts domain.SearchCacheOpts) domain.SearchCacheOpts {
	withDefaults := opts
	if withDefaults.Limit == 0 {
		withDefaults.Limit = 20
	}

	if withDefaults.Sort == "" {
		withDefaults.Sort = domain.SortTypePopular
	}

	if withDefaults.SortOrder == "" {
		withDefaults.SortOrder = utils.Ternary(opts.Sort == domain.SortTypeTitle, domain.SortOrderAsc, domain.SortOrderDesc)
	}

	return withDefaults
}

type searchHTTPResponse struct {
	Data []searchHTTPResponseData `json:"data"`
	Meta searchHTTPResponseMeta   `json:"meta"`
}

func (r *searchHTTPResponse) Domain() *domain.SearchCacheResult {
	meta := domain.SearchResultMeta{
		Total:   r.Meta.Total,
		PerPage: r.Meta.PerPage,
		HasMore: r.Meta.HasMore,
	}

	items := make([]domain.SearchCacheResultItem, len(r.Data))
	for i, data := range r.Data {
		genres := make([]string, len(data.Genres))
		for j, g := range data.Genres {
			genres[j] = string(g.Slug)
		}

		latestChapters := make([]domain.SearchResultItemChapter, len(r.Data[i].LatestChapters))
		for j, c := range r.Data[i].LatestChapters {
			latestChapters[j] = domain.SearchResultItemChapter{
				Number:           c.Number,
				ID:               c.Slug,
				Title:            c.Title,
				EarlyAccessUntil: c.EarlyAccessUntil,
				PublishedAt:      c.PublishedAt,
			}
		}

		items[i] = domain.SearchCacheResultItem{
			ID:             data.ID,
			Slug:           data.Slug,
			Title:          data.Title,
			AltTitles:      data.AltTitles,
			Description:    data.Description,
			Cover:          data.Cover,
			Status:         sources.SeriesStatus(data.Status),
			Type:           sources.SeriesType(data.Type),
			Author:         data.Author,
			Artist:         data.Artist,
			Rating:         data.Rating,
			ChapterCount:   data.ChapterCount,
			LastChapterAt:  data.LastChapterAt,
			CreatedAt:      data.CreatedAt,
			UpdatedAt:      data.UpdatedAt,
			PublicURL:      data.PublicURL,
			SourceURL:      data.SourceURL,
			Genres:         genres,
			LatestChapters: latestChapters,
		}
	}

	return &domain.SearchCacheResult{
		Meta:  meta,
		Items: items,
	}
}

type searchHTTPResponseMeta struct {
	Total   int  `json:"total"`
	PerPage int  `json:"per_page"`
	HasMore bool `json:"has_more"`
}

func (meta *searchHTTPResponseMeta) Domain() domain.SearchResultMeta {
	return domain.SearchResultMeta{
		Total:   meta.Total,
		PerPage: meta.PerPage,
		HasMore: meta.HasMore,
	}
}

type searchHTTPResponseData struct {
	LastChapterAt  time.Time                   `json:"last_chapter_at"`
	UpdatedAt      time.Time                   `json:"updated_at"`
	CreatedAt      time.Time                   `json:"created_at"`
	SourceURL      string                      `json:"source_url"`
	Banner         string                      `json:"banner,omitempty"`
	Status         string                      `json:"status"`
	Type           string                      `json:"type"`
	Author         string                      `json:"author"`
	Artist         string                      `json:"artist"`
	Cover          string                      `json:"cover"`
	PublicURL      string                      `json:"public_url"`
	Slug           string                      `json:"slug"`
	Title          string                      `json:"title"`
	Description    string                      `json:"description"`
	Genres         []searchHTTPResponseGenre   `json:"genres"`
	LatestChapters []searchHTTPResponseChapter `json:"latest_chapters"`
	AltTitles      []string                    `json:"alt_titles,omitempty"`
	PopularityRank int                         `json:"popularity_rank"`
	ID             int                         `json:"id"`
	ChapterCount   int                         `json:"chapter_count"`
	Rating         float64                     `json:"rating"`
	BookmarkCount  int                         `json:"bookmark_count"`
	ReleaseYear    int                         `json:"release_year,omitempty"`
	IsPinned       bool                        `json:"is_pinned,omitempty"`
}

type searchHTTPResponseGenre struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	ID   int    `json:"id"`
}

type searchHTTPResponseChapter struct {
	EarlyAccessUntil time.Time `json:"early_access_until"`
	PublishedAt      time.Time `json:"published_at"`
	CreatedAt        time.Time `json:"created_at"`
	Title            string    `json:"title"`
	Slug             string    `json:"slug"`
	ID               int       `json:"id"`
	SeriesID         int       `json:"series_id"`
	Number           int       `json:"number"`
	PageCount        int       `json:"page_count"`
	ViewCount        int       `json:"view_count"`
	IsPremium        bool      `json:"is_premium"`
	CommentsEnabled  bool      `json:"comments_enabled"`
}
