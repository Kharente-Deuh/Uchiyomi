// SPDX-License-Identifier: AGPL-3.0-or-later

package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
)

func TestIsSeriesSortType(t *testing.T) {
	t.Parallel()

	for _, sort := range []string{"popular", "latest", "rating", "title", "newest", ""} {
		if !domain.IsSeriesSortType(sort) {
			t.Errorf("IsSeriesSortType(%q) = false, want true", sort)
		}
	}

	if domain.IsSeriesSortType("Popular") || domain.IsSeriesSortType("views") {
		t.Fatal("unknown sort types must be rejected")
	}
}

func TestIsSortOrder(t *testing.T) {
	t.Parallel()

	for _, order := range []string{"asc", "desc", ""} {
		if !domain.IsSortOrder(order) {
			t.Errorf("IsSortOrder(%q) = false, want true", order)
		}
	}

	if domain.IsSortOrder("ASC") || domain.IsSortOrder("ascending") {
		t.Fatal("unknown sort orders must be rejected")
	}
}

func TestSearchCacheResultItemDomain(t *testing.T) {
	t.Parallel()

	internalID := uuid.New()
	lastChapterAt := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	item := domain.SearchCacheResultItem{
		LastChapterAt: &lastChapterAt,
		Title:         "Solo Leveling",
		Slug:          "solo-leveling",
		Status:        sources.SeriesStatusOngoing,
		Type:          sources.SeriesTypeManhwa,
		ChapterCount:  200,
		Rating:        9.8,
	}

	got := item.Domain(&internalID)
	if got.Title != item.Title || got.Slug != item.Slug || got.InternalID == nil || *got.InternalID != internalID {
		t.Errorf("Domain() = %+v", got)
	}

	if got.LastChapterAt == nil || !got.LastChapterAt.Equal(lastChapterAt) {
		t.Errorf("LastChapterAt = %v, want %v", got.LastChapterAt, lastChapterAt)
	}
}

func TestGetInfosBySlugResponseSource(t *testing.T) {
	t.Parallel()

	internalID := uuid.New()
	res := domain.GetInfosBySlugResponse{
		Title:        "Nano Machine",
		Slug:         "nano-machine",
		Status:       sources.SeriesStatusOngoing,
		Type:         sources.SeriesTypeManhwa,
		ChapterCount: 150,
		Rating:       8.5,
	}

	got := res.Source(&internalID)
	if got.Title != res.Title || got.Slug != res.Slug || got.InternalID == nil || *got.InternalID != internalID {
		t.Errorf("Source() = %+v", got)
	}

	got = res.Source(nil)
	if got.InternalID != nil {
		t.Errorf("Source(nil).InternalID = %v, want nil", got.InternalID)
	}
}
