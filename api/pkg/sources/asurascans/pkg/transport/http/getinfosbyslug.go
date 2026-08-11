// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
)

//nolint:lll
func (c *Client) GetInfosBySlug(ctx context.Context, slug string) (*domain.GetInfosBySlugResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/series/%s", c.cfg.AsuraURL, slug), nil)
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

	var parsed getInfosBySlugHttpResponse
	if err = json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("json.Decode: %w", err)
	}

	return parsed.Domain(), nil
}

type getInfosBySlugHttpResponse struct {
	RecommendedSeries []getInfosBySlugHttpRecommendedSeries `json:"recommended_series"`
	Series            getInfosBySlugHttpSeries              `json:"series"`
}

func (r *getInfosBySlugHttpResponse) Domain() *domain.GetInfosBySlugResponse {
	genres := make([]string, len(r.Series.Genres))
	for i, g := range r.Series.Genres {
		genres[i] = g.Slug
	}

	return &domain.GetInfosBySlugResponse{
		Slug:          r.Series.Slug,
		Title:         r.Series.Title,
		AltTitles:     r.Series.AltTitles,
		Description:   r.Series.Description,
		Cover:         r.Series.Cover,
		Status:        sources.SeriesStatus(r.Series.Status),
		Type:          sources.SeriesType(r.Series.Type),
		Author:        r.Series.Author,
		Artist:        r.Series.Artist,
		Rating:        r.Series.Rating,
		ChapterCount:  r.Series.ChapterCount,
		LastChapterAt: r.Series.LastChapterAt,
		CreatedAt:     r.Series.CreatedAt,
		UpdatedAt:     r.Series.UpdatedAt,
		PublicURL:     r.Series.PublicURL,
		SourceURL:     r.Series.SourceURL,
		Genres:        genres,
	}
}

type getInfosBySlugHttpRecommendedSeries struct {
	Slug         string  `json:"slug"`
	Title        string  `json:"title"`
	CoverURL     string  `json:"cover_url"`
	Type         string  `json:"type"`
	Status       string  `json:"status"`
	PublicURL    string  `json:"public_url"`
	SourceURL    string  `json:"source_url"`
	ID           int     `json:"id"`
	ChapterCount int     `json:"chapter_count"`
	Rating       float64 `json:"rating"`
}

type getInfosBySlugHttpSeries struct {
	LastChapterAt     time.Time `json:"last_chapter_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	CreatedAt         time.Time `json:"created_at"`
	Author            string    `json:"author"`
	AlternativeTitles string    `json:"alternative_titles"`
	Description       string    `json:"description"`
	Cover             string    `json:"cover"`
	Banner            string    `json:"banner"`
	Status            string    `json:"status"`
	Type              string    `json:"type"`
	SourceURL         string    `json:"source_url"`
	Artist            string    `json:"artist"`
	PublicURL         string    `json:"public_url"`
	Slug              string    `json:"slug"`
	Title             string    `json:"title"`
	AltTitles         []string  `json:"alt_titles"`
	Genres            []struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
		ID   int    `json:"id"`
	} `json:"genres"`
	ChapterCount   int     `json:"chapter_count"`
	Rating         float64 `json:"rating"`
	BookmarkCount  int     `json:"bookmark_count"`
	PopularityRank int     `json:"popularity_rank"`
	ID             int     `json:"id"`
	PdfAvailable   bool    `json:"pdf_available"`
}
