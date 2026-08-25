// SPDX-License-Identifier: AGPL-3.0-or-later

package parse

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

type SeriesPage struct {
	Chapters []SeriesChapter
	Infos    SeriesInfos
}

type SeriesInfos struct {
	UpdatedAt    time.Time
	CreatedAt    time.Time
	Description  string
	Title        string
	Cover        string
	Status       sources.SeriesStatus
	Type         sources.SeriesType
	Author       string
	Artist       string
	Slug         string
	AltTitles    []string
	Genres       []string
	ChapterCount int
}

type SeriesChapter struct {
	PublishedAt time.Time
	ID          string
	Title       string
	Number      float64
	PageCount   int
}

var (
	seriesDateRe       = regexp.MustCompile(`[A-Z][a-z]+ \d{1,2}, \d{4}`)
	chapterNumberLabel = regexp.MustCompile(`(?i)chapter\s+(\d+(?:\.\d+)?)`)
)

func isTypeGenreLabel(genre string) bool {
	switch genre {
	case labelManga, labelManhwa, labelManhua, labelMangatoon, labelComic, labelNovel:
		return true
	default:
		return false
	}
}

func ParseSeries(html, slug string) (SeriesPage, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return SeriesPage{}, fmt.Errorf("parse.ParseSeries: %w", err)
	}

	seriesType, err := MapSeriesType(tableValue(doc, "Type"))
	if err != nil {
		return SeriesPage{}, fmt.Errorf("parse.ParseSeries: %w", err)
	}

	createdAt, err := parseSeriesDate(tableValue(doc, "Posted On"))
	if err != nil {
		return SeriesPage{}, fmt.Errorf("parse.ParseSeries: %w", err)
	}

	updatedAt, err := parseSeriesDate(tableValue(doc, "Updated On"))
	if err != nil {
		return SeriesPage{}, fmt.Errorf("parse.ParseSeries: %w", err)
	}

	chapters, err := parseSeriesChapters(doc, slug)
	if err != nil {
		return SeriesPage{}, fmt.Errorf("parse.ParseSeries: %w", err)
	}

	return SeriesPage{
		Infos: SeriesInfos{
			Title:        seriesTitle(doc),
			Cover:        seriesCover(doc),
			Status:       MapSeriesStatus(tableValue(doc, "Status")),
			Type:         seriesType,
			Author:       CleanPerson(tableValue(doc, "Author")),
			Artist:       CleanPerson(tableValue(doc, "Artist")),
			Description:  seriesDescription(doc),
			Slug:         slug,
			AltTitles:    CleanAltTitles(tableValue(doc, "Alternative")),
			Genres:       seriesGenres(doc),
			ChapterCount: len(chapters),
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		},
		Chapters: chapters,
	}, nil
}

func ParseChapterNumber(label string) (float64, bool) {
	label = strings.TrimSpace(label)
	if label == "" {
		return 0, false
	}

	if number, err := strconv.ParseFloat(label, 64); err == nil {
		return number, true
	}

	match := chapterNumberLabel.FindStringSubmatch(label)
	if len(match) < 2 {
		return 0, false
	}

	number, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, false
	}

	return number, true
}

func tableValue(doc *goquery.Document, label string) string {
	var value string

	doc.Find(".infotable tr").Each(func(_ int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() < 2 {
			return
		}

		if !strings.EqualFold(strings.TrimSpace(cells.First().Text()), label) {
			return
		}

		value = strings.TrimSpace(cells.Last().Text())
	})

	return value
}

func seriesTitle(doc *goquery.Document) string {
	if title := doc.Find("h1.entry-title").First(); title.Length() > 0 {
		return strings.TrimSpace(title.Text())
	}

	return ""
}

func seriesCover(doc *goquery.Document) string {
	if img := doc.Find("img.attachment-post-thumbnail").First(); img.Length() > 0 {
		return imageURL(img)
	}

	if img := doc.Find(".thumb img").First(); img.Length() > 0 {
		return imageURL(img)
	}

	return ""
}

func imageURL(img *goquery.Selection) string {
	if src, exists := img.Attr("src"); exists && src != "" {
		return src
	}

	if dataSrc, exists := img.Attr("data-src"); exists {
		return dataSrc
	}

	return ""
}

func seriesDescription(doc *goquery.Document) string {
	if paragraph := doc.Find(".desc p").First(); paragraph.Length() > 0 {
		return strings.TrimSpace(paragraph.Text())
	}

	var description string

	doc.Find(".postbody p").EachWithBreak(func(_ int, paragraph *goquery.Selection) bool {
		text := strings.TrimSpace(paragraph.Text())
		if text == "" {
			return true
		}

		description = text

		return false
	})

	return description
}

func seriesGenres(doc *goquery.Document) []string {
	links := doc.Find(".mgen a")
	if links.Length() == 0 {
		links = doc.Find(".seriestugenre a")
	}

	genres := make([]string, 0, links.Length())

	links.Each(func(_ int, link *goquery.Selection) {
		genre := strings.ToLower(strings.TrimSpace(link.Text()))
		if genre == "" {
			return
		}

		if isTypeGenreLabel(genre) {
			return
		}

		genres = append(genres, genre)
	})

	return genres
}

func parseSeriesChapters(doc *goquery.Document, slug string) ([]SeriesChapter, error) {
	items := doc.Find("#chapterlist li")
	if items.Length() == 0 {
		items = doc.Find(".eplister li")
	}

	chapters := make([]SeriesChapter, 0, items.Length())

	var err error

	items.Each(func(_ int, item *goquery.Selection) {
		if err != nil {
			return
		}

		numberText, _ := item.Attr("data-num")
		if numberText == "" {
			numberText = item.Find(".chapternum").First().Text()
		}

		if numberText == "" {
			numberText = item.Text()
		}

		number, ok := ParseChapterNumber(numberText)
		if !ok {
			return
		}

		dateText := strings.TrimSpace(item.Find(".chapterdate").First().Text())
		if dateText == "" {
			dateText = chapterDateFromText(item.Text())
		}

		publishedAt, parseErr := parseSeriesDate(dateText)
		if parseErr != nil {
			err = parseErr

			return
		}

		chapters = append(chapters, SeriesChapter{
			ID:          SourceChapterSlug(slug, number),
			Number:      number,
			Title:       "",
			PageCount:   0,
			PublishedAt: publishedAt,
		})
	})

	if err != nil {
		return nil, err
	}

	return chapters, nil
}

func chapterDateFromText(text string) string {
	match := seriesDateRe.FindString(text)
	if match == "" {
		return ""
	}

	return match
}

func parseSeriesDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}

	parsed, err := time.Parse("January 2, 2006", raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("time.Parse: %w", err)
	}

	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC), nil
}
