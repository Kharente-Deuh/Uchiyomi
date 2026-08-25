// SPDX-License-Identifier: AGPL-3.0-or-later

package sources_test

import (
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

func TestParseSourceName(t *testing.T) {
	t.Parallel()

	got, err := sources.ParseSourceName("asurascans")
	if err != nil || got != sources.SourceAsuraScans {
		t.Fatalf("asurascans: got %q err %v", got, err)
	}

	if _, err = sources.ParseSourceName("AsuraScans"); err == nil {
		t.Fatal("AsuraScans must be invalid")
	}

	got, err = sources.ParseSourceName("kingofshojo")
	if err != nil || got != sources.SourceKingOfShojo {
		t.Fatalf("kingofshojo: got %q err %v", got, err)
	}

	if _, err = sources.ParseSourceName("KingOfShojo"); err == nil {
		t.Fatal("KingOfShojo must be invalid")
	}

	if _, err = sources.ParseSourceName(""); err == nil {
		t.Fatal("empty source name must be invalid")
	}
}

func TestParseSeriesType(t *testing.T) {
	t.Parallel()

	tests := map[string]sources.SeriesType{
		"manga":     sources.SeriesTypeManga,
		"mangatoon": sources.SeriesTypeMangatoon,
		"manhua":    sources.SeriesTypeManhua,
		"manhwa":    sources.SeriesTypeManhwa,
	}

	for raw, want := range tests {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			got, err := sources.ParseSeriesType(raw)
			if err != nil || got != want {
				t.Fatalf("ParseSeriesType(%q) = %q err %v, want %q", raw, got, err, want)
			}
		})
	}

	if _, err := sources.ParseSeriesType("Manga"); err == nil {
		t.Fatal("Manga must be invalid")
	}
}

func TestParseSeriesStatus(t *testing.T) {
	t.Parallel()

	tests := map[string]sources.SeriesStatus{
		"ongoing":   sources.SeriesStatusOngoing,
		"completed": sources.SeriesStatusCompleted,
		"hiatus":    sources.SeriesStatusHiatus,
		"cancelled": sources.SeriesStatusCancelled,
		"dropped":   sources.SeriesStatusDropped,
	}

	for raw, want := range tests {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			got, err := sources.ParseSeriesStatus(raw)
			if err != nil || got != want {
				t.Fatalf("ParseSeriesStatus(%q) = %q err %v, want %q", raw, got, err, want)
			}
		})
	}

	if _, err := sources.ParseSeriesStatus("ONGOING"); err == nil {
		t.Fatal("ONGOING must be invalid")
	}
}
