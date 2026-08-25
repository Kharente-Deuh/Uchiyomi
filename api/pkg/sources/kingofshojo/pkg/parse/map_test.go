// SPDX-License-Identifier: AGPL-3.0-or-later

package parse_test

import (
	"errors"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/parse"
)

func TestMapSeriesType(t *testing.T) {
	t.Parallel()

	got, err := parse.MapSeriesType("Manga")
	if err != nil || got != sources.SeriesTypeMangatoon {
		t.Fatalf("manga: got %q err %v", got, err)
	}

	got, err = parse.MapSeriesType("MANHWA")
	if err != nil || got != sources.SeriesTypeManhwa {
		t.Fatalf("manhwa: got %q err %v", got, err)
	}

	if _, err = parse.MapSeriesType("novel"); !errors.Is(err, parse.ErrUnsupportedType) {
		t.Fatalf("novel: err %v", err)
	}

	if _, err = parse.MapSeriesType("comic"); !errors.Is(err, parse.ErrUnsupportedType) {
		t.Fatalf("comic: err %v", err)
	}
}

func TestMapSeriesStatus(t *testing.T) {
	t.Parallel()

	if parse.MapSeriesStatus("") != sources.SeriesStatusOngoing {
		t.Fatal("empty must be ongoing")
	}

	if parse.MapSeriesStatus("  Hiatus ") != sources.SeriesStatusHiatus {
		t.Fatal("hiatus")
	}

	if parse.MapSeriesStatus("Completed") != sources.SeriesStatusCompleted {
		t.Fatal("completed")
	}
}

func TestCleanPerson(t *testing.T) {
	t.Parallel()

	if parse.CleanPerson("n/a") != "" || parse.CleanPerson(" N/A ") != "" {
		t.Fatal("n/a must be empty")
	}

	if parse.CleanPerson("Jane") != "Jane" {
		t.Fatal("keep real names")
	}
}

func TestCleanAltTitles(t *testing.T) {
	t.Parallel()

	if len(parse.CleanAltTitles("Unknown")) != 0 || len(parse.CleanAltTitles("")) != 0 {
		t.Fatal("placeholders must be empty")
	}
}

func TestSourceChapterSlug(t *testing.T) {
	t.Parallel()

	got := parse.SourceChapterSlug("tears-on-a-withered-flower", 111)
	if got != "tears-on-a-withered-flower-chapter-111" {
		t.Fatalf("got %q", got)
	}

	got = parse.SourceChapterSlug("tears-on-a-withered-flower", 111.5)
	if got != "tears-on-a-withered-flower-chapter-111-5" {
		t.Fatalf("got %q", got)
	}
}
