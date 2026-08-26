// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/parse"
)

const (
	kosPerPage = 40
	mangaPath  = "/manga/"
)

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

func searchTotal(offset, limit, nItems int, hasNext bool) int {
	if hasNext {
		return offset + limit + 1
	}

	return offset + nItems
}

func (c *Client) fetchSearchCards(ctx context.Context, opts domain.SearchCacheOpts) ([]parse.SearchCard, int, error) {
	pageNum, start := kosPage(opts.Offset, opts.Limit)

	page, err := c.fetchSearchPage(ctx, pageNum, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("c.fetchSearchPage: %w", err)
	}

	result, leftover := appendSearchCards(nil, page.Items, start, opts.Limit)
	hasNext := leftover || page.HasNext
	fullPage := len(page.Items) >= kosPerPage

	for len(result) < opts.Limit && fullPage {
		pageNum++

		nextPage, err := c.fetchSearchPage(ctx, pageNum, opts)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				hasNext = leftover

				break
			}

			return nil, 0, fmt.Errorf("c.fetchSearchPage: %w", err)
		}

		if len(nextPage.Items) == 0 {
			hasNext = leftover

			break
		}

		page = nextPage
		result, leftover = appendSearchCards(result, page.Items, 0, opts.Limit-len(result))
		hasNext = leftover || page.HasNext
		fullPage = len(page.Items) >= kosPerPage
	}

	return result, searchTotal(opts.Offset, opts.Limit, len(result), hasNext), nil
}

func appendSearchCards(dst, cards []parse.SearchCard, start, remaining int) ([]parse.SearchCard, bool) {
	filtered := filterSearchCards(cards)
	if start < 0 {
		start = 0
	}

	if start > len(filtered) {
		start = len(filtered)
	}

	available := filtered[start:]
	leftover := remaining < len(available)
	if remaining > len(available) {
		remaining = len(available)
	}

	if remaining <= 0 {
		return dst, leftover
	}

	return append(dst, available[:remaining]...), leftover
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
		path = mangaPath
	} else {
		path = mangaPath + "page/" + strconv.Itoa(pageNum) + "/"
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
