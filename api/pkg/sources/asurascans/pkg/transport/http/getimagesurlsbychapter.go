// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
)

//nolint:lll
func (c *Client) GetImageURLsByChapter(ctx context.Context, opts domain.GetImageURLsByChapterOpts) (*[]string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s/series/%s/chapters/%s", c.cfg.AsuraURL, opts.SeriesSlug, opts.ChapterID),
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	req.Header.Add("Accept", "application/json")
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

	var parsed getChapterHttpResponse
	if err = json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("json.Decode: %w", err)
	}

	urls := parsed.ToDomain()

	return &urls, nil
}

type getChapterHttpResponse struct {
	Data struct {
		UnlockTime  interface{} `json:"unlock_time"`
		AccessGate  string      `json:"access_gate"`
		ChapterList []struct {
			CreatedAt       time.Time `json:"created_at"`
			Slug            string    `json:"slug"`
			Title           string    `json:"title,omitempty"`
			ID              int       `json:"id"`
			SeriesID        int       `json:"series_id"`
			Number          int       `json:"number"`
			PageCount       int       `json:"page_count"`
			ViewCount       int       `json:"view_count"`
			IsPremium       bool      `json:"is_premium"`
			CommentsEnabled bool      `json:"comments_enabled"`
		} `json:"chapter_list"`
		Series struct {
			CreatedAt     time.Time `json:"created_at"`
			UpdatedAt     time.Time `json:"updated_at"`
			Cover         string    `json:"cover"`
			Status        string    `json:"status"`
			Type          string    `json:"type"`
			Title         string    `json:"title"`
			Slug          string    `json:"slug"`
			PublicURL     string    `json:"public_url"`
			SourceURL     string    `json:"source_url"`
			ID            int       `json:"id"`
			BookmarkCount int       `json:"bookmark_count"`
			Rating        int       `json:"rating"`
			ChapterCount  int       `json:"chapter_count"`
		} `json:"series"`
		NextChapter struct {
			CreatedAt       time.Time `json:"created_at"`
			Slug            string    `json:"slug"`
			ID              int       `json:"id"`
			SeriesID        int       `json:"series_id"`
			Number          int       `json:"number"`
			PageCount       int       `json:"page_count"`
			ViewCount       int       `json:"view_count"`
			IsPremium       bool      `json:"is_premium"`
			CommentsEnabled bool      `json:"comments_enabled"`
		} `json:"next_chapter"`
		PrevChapter struct {
			CreatedAt       time.Time `json:"created_at"`
			Slug            string    `json:"slug"`
			ID              int       `json:"id"`
			SeriesID        int       `json:"series_id"`
			Number          int       `json:"number"`
			PageCount       int       `json:"page_count"`
			ViewCount       int       `json:"view_count"`
			IsPremium       bool      `json:"is_premium"`
			CommentsEnabled bool      `json:"comments_enabled"`
		} `json:"prev_chapter"`
		Chapter struct {
			PublishedAt time.Time `json:"published_at"`
			CreatedAt   time.Time `json:"created_at"`
			Slug        string    `json:"slug"`
			SeriesSlug  string    `json:"series_slug"`
			Pages       []struct {
				URL    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"pages"`
			ID              int  `json:"id"`
			SeriesID        int  `json:"series_id"`
			Number          int  `json:"number"`
			PageCount       int  `json:"page_count"`
			ViewCount       int  `json:"view_count"`
			IsPremium       bool `json:"is_premium"`
			CommentsEnabled bool `json:"comments_enabled"`
		} `json:"chapter"`
		CommentCount int  `json:"comment_count"`
		IsLocked     bool `json:"is_locked"`
	} `json:"data"`
}

func (c *getChapterHttpResponse) ToDomain() []string {
	urls := make([]string, len(c.Data.Chapter.Pages))
	for i, p := range c.Data.Chapter.Pages {
		urls[i] = p.URL
	}

	return urls
}
