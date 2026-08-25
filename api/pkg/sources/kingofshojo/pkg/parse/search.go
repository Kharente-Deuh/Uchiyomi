// SPDX-License-Identifier: AGPL-3.0-or-later

package parse

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SearchPage struct {
	Items    []SearchCard
	LastPage int
}

type SearchCard struct {
	Slug        string
	Title       string
	Cover       string
	LastChapter float64
	Skip        bool
}

var (
	chapterRe  = regexp.MustCompile(`(?i)chapter\s+(\d+(?:\.\d+)?)`)
	lastPageRe = regexp.MustCompile(`(?i)page\s+\d+\s+of\s+(\d+)`)
)

func ParseSearch(html string) (SearchPage, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return SearchPage{}, fmt.Errorf("parse.ParseSearch: %w", err)
	}

	page := SearchPage{}

	if m := lastPageRe.FindStringSubmatch(doc.Text()); len(m) > 1 {
		page.LastPage, _ = strconv.Atoi(m[1])
	}

	cards := doc.Find(".listupd .bsx")
	if cards.Length() == 0 {
		cards = doc.Find(".bsx")
	}

	cards.Each(func(_ int, card *goquery.Selection) {
		link := card.Find(`a[href*="/manga/"]`)
		if link.Length() == 0 {
			return
		}

		href, _ := link.Attr("href")
		slug := mangaSlugFromHref(href)
		if slug == "" {
			return
		}

		title := strings.TrimSpace(card.Find(".tt").First().Text())

		cover := ""
		if img := card.Find("img").First(); img.Length() > 0 {
			if src, exists := img.Attr("src"); exists && src != "" {
				cover = src
			} else if dataSrc, exists := img.Attr("data-src"); exists {
				cover = dataSrc
			}
		}

		var lastChapter float64
		if m := chapterRe.FindStringSubmatch(card.Text()); len(m) > 1 {
			lastChapter, _ = strconv.ParseFloat(m[1], 64)
		}

		skip := cardHasComicOrNovelType(card)

		page.Items = append(page.Items, SearchCard{
			Slug:        slug,
			Title:       title,
			Cover:       cover,
			LastChapter: lastChapter,
			Skip:        skip,
		})
	})

	return page, nil
}

func mangaSlugFromHref(href string) string {
	idx := strings.Index(href, "/manga/")
	if idx < 0 {
		return ""
	}

	path := strings.Trim(href[idx+len("/manga/"):], "/")
	if path == "" {
		return ""
	}

	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}

	return ""
}

func cardHasComicOrNovelType(card *goquery.Selection) bool {
	typeText := strings.ToLower(strings.TrimSpace(card.Find(".type").First().Text()))

	return typeText == labelComic || typeText == labelNovel
}
