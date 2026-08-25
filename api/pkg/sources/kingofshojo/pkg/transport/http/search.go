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

const kosPerPage = 40

func kosPage(offset, _ int) (page int, pageOffset int) {
	return offset/kosPerPage + 1, offset % kosPerPage
}

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

	cards, total, err := c.fetchSearchCards(ctx, withDefaults)
	if err != nil {
		return nil, fmt.Errorf("c.fetchSearchCards: %w", err)
	}

	items := make([]domain.SearchCacheResultItem, len(cards))
	for i, card := range cards {
		items[i] = searchCardToItem(card, c.cfg.BaseURL)
	}

	hasMore := withDefaults.Offset+len(items) < total

	return &domain.SearchCacheResult{
		Items: items,
		Meta: domain.SearchResultMeta{
			Total:   total,
			PerPage: withDefaults.Limit,
			HasMore: hasMore,
		},
	}, nil
}

func searchOptsWithDefaults(opts domain.SearchCacheOpts) domain.SearchCacheOpts {
	withDefaults := opts
	if withDefaults.Limit == 0 {
		withDefaults.Limit = 20
	}

	if withDefaults.Sort == "" {
		withDefaults.Sort = domain.SortTypePopular
	}

	if withDefaults.Sort == domain.SortTypeTitle && withDefaults.SortOrder == "" {
		withDefaults.SortOrder = domain.SortOrderAsc
	}

	return withDefaults
}

func (c *Client) fetchSearchCards(ctx context.Context, opts domain.SearchCacheOpts) ([]parse.SearchCard, int, error) {
	pageNum, start := kosPage(opts.Offset, opts.Limit)

	firstPage, err := c.fetchSearchPage(ctx, pageNum, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("c.fetchSearchPage: %w", err)
	}

	total := 0
	if len(firstPage.Items) > 0 {
		total = firstPage.LastPage * len(firstPage.Items)
	}

	filtered := filterSearchCards(firstPage.Items)

	take1 := opts.Limit
	if take1 > len(filtered)-start {
		take1 = len(filtered) - start
	}

	if take1 < 0 {
		take1 = 0
	}

	var result []parse.SearchCard
	if start < len(filtered) {
		result = append(result, filtered[start:start+take1]...)
	}

	remaining := opts.Limit - len(result)
	if remaining > 0 && start+opts.Limit > kosPerPage {
		secondPage, err := c.fetchSearchPage(ctx, pageNum+1, opts)
		if err != nil {
			return nil, 0, fmt.Errorf("c.fetchSearchPage: %w", err)
		}

		secondFiltered := filterSearchCards(secondPage.Items)
		take2 := remaining
		if take2 > len(secondFiltered) {
			take2 = len(secondFiltered)
		}

		result = append(result, secondFiltered[:take2]...)
	}

	return result, total, nil
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
	pageNum int,
	opts domain.SearchCacheOpts,
) (parse.SearchPage, error) {
	targetURL := c.searchPageURL(pageNum, opts)

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

func (c *Client) searchPageURL(pageNum int, opts domain.SearchCacheOpts) string {
	var path string
	if pageNum <= 1 {
		path = "/manga/"
	} else {
		path = "/manga/page/" + strconv.Itoa(pageNum) + "/"
	}

	if q := c.buildSearchQuery(opts); q != "" {
		return c.mangaURL(path) + "?" + q
	}

	return c.mangaURL(path)
}

func (c *Client) buildSearchQuery(opts domain.SearchCacheOpts) string {
	q := url.Values{}

	if opts.Search != "" {
		q.Set("title", opts.Search)
	}

	if order := kosOrder(opts.Sort, opts.SortOrder); order != "" {
		q.Set("order", order)
	}

	return q.Encode()
}

func searchCardToItem(card parse.SearchCard, baseURL string) domain.SearchCacheResultItem {
	seriesURL := baseURL + "/manga/" + card.Slug + "/"

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
