// SPDX-License-Identifier: AGPL-3.0-or-later

package parse_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/parse"
)

func TestParseSeries(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("testdata", "series.html"))
	if err != nil {
		t.Fatal(err)
	}

	const slug = "tears-on-a-withered-flower"

	page, err := parse.ParseSeries(string(raw), slug)
	if err != nil {
		t.Fatal(err)
	}

	wantDate := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)

	infos := page.Infos
	if infos.Title != "Tears on a Withered Flower" {
		t.Fatalf("title = %q", infos.Title)
	}

	if infos.Cover != "https://cdn.example/series-cover.jpg" {
		t.Fatalf("cover = %q", infos.Cover)
	}

	if infos.Status != sources.SeriesStatusOngoing {
		t.Fatalf("status = %q", infos.Status)
	}

	if infos.Type != sources.SeriesTypeManhwa {
		t.Fatalf("type = %q", infos.Type)
	}

	if infos.Author != "" {
		t.Fatalf("author = %q", infos.Author)
	}

	if infos.Artist != "Some Artist" {
		t.Fatalf("artist = %q", infos.Artist)
	}

	if infos.Description != "A floral romance story that spans seasons." {
		t.Fatalf("description = %q", infos.Description)
	}

	if infos.Slug != slug {
		t.Fatalf("slug = %q", infos.Slug)
	}

	if len(infos.AltTitles) != 0 {
		t.Fatalf("altTitles = %v", infos.AltTitles)
	}

	if len(infos.Genres) != 2 || infos.Genres[0] != "mature" || infos.Genres[1] != "romance" {
		t.Fatalf("genres = %v", infos.Genres)
	}

	if infos.ChapterCount != 2 {
		t.Fatalf("chapterCount = %d", infos.ChapterCount)
	}

	if !infos.CreatedAt.Equal(wantDate) {
		t.Fatalf("createdAt = %v", infos.CreatedAt)
	}

	if !infos.UpdatedAt.Equal(wantDate) {
		t.Fatalf("updatedAt = %v", infos.UpdatedAt)
	}

	if len(page.Chapters) != 2 {
		t.Fatalf("len(chapters) = %d", len(page.Chapters))
	}

	if page.Chapters[0].ID != slug+"-chapter-115" {
		t.Fatalf("chapter0 id = %q", page.Chapters[0].ID)
	}

	if page.Chapters[0].Number != 115 || page.Chapters[0].Title != "" || page.Chapters[0].PageCount != 0 {
		t.Fatalf("chapter0 %+v", page.Chapters[0])
	}

	if !page.Chapters[0].PublishedAt.Equal(wantDate) {
		t.Fatalf("chapter0 date = %v", page.Chapters[0].PublishedAt)
	}

	if page.Chapters[1].ID != slug+"-chapter-111-5" {
		t.Fatalf("chapter1 id = %q", page.Chapters[1].ID)
	}

	if page.Chapters[1].Number != 111.5 {
		t.Fatalf("chapter1 number = %v", page.Chapters[1].Number)
	}
}

func TestParseSeriesUnparseableChapterDate(t *testing.T) {
	t.Parallel()

	html := strings.Replace(
		readSeriesFixture(t),
		`<span class="chapterdate">August 21, 2026</span>`,
		`<span class="chapterdate">2 days ago</span>`,
		1,
	)

	page, err := parse.ParseSeries(html, "tears-on-a-withered-flower")
	if err != nil {
		t.Fatal(err)
	}

	if page.Infos.Title != "Tears on a Withered Flower" {
		t.Fatalf("title = %q", page.Infos.Title)
	}

	if len(page.Chapters) != 2 {
		t.Fatalf("len(chapters) = %d, want 2", len(page.Chapters))
	}

	if !page.Chapters[0].PublishedAt.IsZero() {
		t.Fatalf("chapter0 PublishedAt = %v, want zero", page.Chapters[0].PublishedAt)
	}

	wantDate := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	if !page.Chapters[1].PublishedAt.Equal(wantDate) {
		t.Fatalf("chapter1 PublishedAt = %v, want %v", page.Chapters[1].PublishedAt, wantDate)
	}
}

func TestParseSeriesMangaType(t *testing.T) {
	t.Parallel()

	html := strings.Replace(
		readSeriesFixture(t),
		"<td>Manhwa</td>",
		"<td>Manga</td>",
		1,
	)

	page, err := parse.ParseSeries(html, "example-manga")
	if err != nil {
		t.Fatal(err)
	}

	if page.Infos.Type != sources.SeriesTypeMangatoon {
		t.Fatalf("type = %q", page.Infos.Type)
	}
}

func TestParseSeriesNovelType(t *testing.T) {
	t.Parallel()

	html := strings.Replace(
		readSeriesFixture(t),
		"<td>Manhwa</td>",
		"<td>Novel</td>",
		1,
	)

	_, err := parse.ParseSeries(html, "example-novel")
	if !errors.Is(err, parse.ErrUnsupportedType) {
		t.Fatalf("err = %v", err)
	}
}

func readSeriesFixture(t *testing.T) string {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("testdata", "series.html"))
	if err != nil {
		t.Fatal(err)
	}

	return string(raw)
}
