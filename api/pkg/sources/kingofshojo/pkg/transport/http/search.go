// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/parse"
)

const mangaPath = "/manga/"

func kosOrder(sort domain.SortType, order domain.SortOrder) string {
	switch sort {
	case domain.SortTypeLatest:
		return "updated"
	case domain.SortTypeNewest:
		return "added"
	case domain.SortTypeTitle:
		if order == domain.SortOrderDesc {
			return "titlereverse"
		}

		return "title"
	case domain.SortTypePopular, domain.SortTypeNone:
		return string(domain.SortTypePopular)
	default:
		return string(domain.SortTypePopular)
	}
}

func (c *Client) Search(ctx context.Context, opts domain.SearchCacheOpts) (*domain.SearchCacheResult, error) {
	withDefaults := searchOptsWithDefaults(opts)

	page, err := c.fetchSearchPage(ctx, withDefaults)
	if err != nil {
		return nil, fmt.Errorf("c.fetchSearchPage: %w", err)
	}

	cards := filterSearchCards(page.Items)
	items := make([]domain.SearchCacheResultItem, len(cards))
	for i, card := range cards {
		items[i] = searchCardToItem(card, c.cfg.BaseURL)
	}

	return &domain.SearchCacheResult{
		Items: items,
		Meta: domain.SearchResultMeta{
			HasNextPage: page.HasNext,
		},
	}, nil
}

func searchOptsWithDefaults(opts domain.SearchCacheOpts) domain.SearchCacheOpts {
	withDefaults := opts

	if withDefaults.Page < 1 {
		withDefaults.Page = 1
	}

	if withDefaults.Sort == "" {
		withDefaults.Sort = domain.SortTypePopular
	}

	if withDefaults.Sort == domain.SortTypeTitle && withDefaults.SortOrder == "" {
		withDefaults.SortOrder = domain.SortOrderAsc
	}

	return withDefaults
}

func filterSearchCards(cards []parse.SearchCard) []parse.SearchCard {
	filtered := make([]parse.SearchCard, 0, len(cards))
	for _, card := range cards {
		if card.Skip {
			continue
		}

		filtered = append(filtered, card)
	}

	return filtered
}

func (c *Client) fetchSearchPage(
	ctx context.Context,
	opts domain.SearchCacheOpts,
) (parse.SearchPage, error) {
	targetURL := c.searchPageURL(opts)

	status, body, err := c.get(ctx, targetURL)
	if err != nil {
		return parse.SearchPage{}, fmt.Errorf("c.get: %w", err)
	}

	if status == http.StatusNotFound {
		return parse.SearchPage{}, domain.ErrNotFound
	}

	page, err := parse.ParseSearch(string(body))
	if err != nil {
		return parse.SearchPage{}, fmt.Errorf("parse.ParseSearch: %w", err)
	}

	return page, nil
}

func (c *Client) searchPageURL(opts domain.SearchCacheOpts) string {
	if q := c.buildSearchQuery(opts); q != "" {
		return c.mangaURL(mangaPath) + "?" + q
	}

	return c.mangaURL(mangaPath)
}

func (c *Client) buildSearchQuery(opts domain.SearchCacheOpts) string {
	q := url.Values{}

	if opts.Search != "" {
		q.Set("title", opts.Search)
	}

	if order := kosOrder(opts.Sort, opts.SortOrder); order != "" {
		q.Set("order", order)
	}

	if opts.Page > 1 {
		q.Set("page", strconv.Itoa(opts.Page))
	}

	return q.Encode()
}

func searchCardToItem(card parse.SearchCard, baseURL string) domain.SearchCacheResultItem {
	seriesURL := baseURL + mangaPath + card.Slug + "/"

	return domain.SearchCacheResultItem{
		Slug:      card.Slug,
		Title:     card.Title,
		Cover:     card.Cover,
		PublicURL: seriesURL,
		SourceURL: seriesURL,
		AltTitles: []string{},
		Genres:    []string{},
		LatestChapters: []domain.SearchResultItemChapter{
			{
				ID:     parse.SourceChapterSlug(card.Slug, card.LastChapter),
				Number: card.LastChapter,
			},
		},
	}
}
