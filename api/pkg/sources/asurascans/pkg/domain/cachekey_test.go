// SPDX-License-Identifier: AGPL-3.0-or-later

package domain_test

import (
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
)

const (
	testGenreAction = "action"
	testGenreDrama  = "drama"
	testValueAutre  = "autre"
)

func TestSearchOptsCacheKeyIsInjective(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		a    domain.SearchOpts
		b    domain.SearchOpts
		same bool
	}{
		"genres découpés différemment": {
			a: domain.SearchOpts{Genres: []domain.SeriesGenre{"Slice", "of", "Life"}},
			b: domain.SearchOpts{Genres: []domain.SeriesGenre{"Slice of Life"}},
		},
		"champs texte adjacents décalés": {
			a: domain.SearchOpts{Status: "ongoing", Type: "manga"},
			b: domain.SearchOpts{Status: "ongoing manga"},
		},
		"search vide contre genre homonyme": {
			a: domain.SearchOpts{Search: testGenreAction},
			b: domain.SearchOpts{Genres: []domain.SeriesGenre{testGenreAction}},
		},
		"ordre des genres indifférent": {
			a:    domain.SearchOpts{Genres: []domain.SeriesGenre{testGenreAction, testGenreDrama}},
			b:    domain.SearchOpts{Genres: []domain.SeriesGenre{testGenreDrama, testGenreAction}},
			same: true,
		},
		"genre répété indifférent": {
			a:    domain.SearchOpts{Genres: []domain.SeriesGenre{testGenreAction, testGenreAction}},
			b:    domain.SearchOpts{Genres: []domain.SeriesGenre{testGenreAction}},
			same: true,
		},
		"nil et slice vide équivalents": {
			a:    domain.SearchOpts{Genres: nil},
			b:    domain.SearchOpts{Genres: []domain.SeriesGenre{}},
			same: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			keyA, keyB := tc.a.CacheKey(), tc.b.CacheKey()

			if tc.same && keyA != keyB {
				t.Errorf("CacheKey() = %q et %q, want identiques", keyA, keyB)
			}

			if !tc.same && keyA == keyB {
				t.Errorf("CacheKey() = %q pour deux options différentes, want distinctes", keyA)
			}
		})
	}
}

func TestSearchOptsCacheKeyDoesNotMutateGenres(t *testing.T) {
	t.Parallel()

	genres := []domain.SeriesGenre{testGenreDrama, testGenreAction}
	opts := domain.SearchOpts{Genres: genres}

	_ = opts.CacheKey()

	if genres[0] != testGenreDrama || genres[1] != testGenreAction {
		t.Errorf("Genres = %v après CacheKey(), want [drama action] (la slice appartient à l'appelant)", genres)
	}
}

func TestSearchOptsCacheKeyDistinguishesEveryField(t *testing.T) {
	t.Parallel()

	base := domain.SearchOpts{
		Offset: 1, Limit: 2, Search: "s", Sort: domain.SortTypeLatest,
		SortOrder: domain.SortOrderAsc, Status: "st", Type: "ty", Artist: "ar",
		Genres: []domain.SeriesGenre{"g"}, MinChapters: 3,
	}

	variants := map[string]domain.SearchOpts{
		"Offset":      {Offset: 9},
		"Limit":       {Limit: 9},
		"Search":      {Search: testValueAutre},
		"Sort":        {Sort: domain.SortTypeRating},
		"SortOrder":   {SortOrder: domain.SortOrderDesc},
		"Status":      {Status: testValueAutre},
		"Type":        {Type: testValueAutre},
		"Artist":      {Artist: testValueAutre},
		"MinChapters": {MinChapters: 9},
	}

	baseKey := base.CacheKey()

	for field, variant := range variants {
		t.Run(field, func(t *testing.T) {
			t.Parallel()

			opts := base

			switch field {
			case "Offset":
				opts.Offset = variant.Offset
			case "Limit":
				opts.Limit = variant.Limit
			case "Search":
				opts.Search = variant.Search
			case "Sort":
				opts.Sort = variant.Sort
			case "SortOrder":
				opts.SortOrder = variant.SortOrder
			case "Status":
				opts.Status = variant.Status
			case "Type":
				opts.Type = variant.Type
			case "Artist":
				opts.Artist = variant.Artist
			case "MinChapters":
				opts.MinChapters = variant.MinChapters
			}

			if got := opts.CacheKey(); got == baseKey {
				t.Errorf("changer %s ne change pas la clé (%q)", field, got)
			}
		})
	}
}

func TestGetImageURLsByChapterOptsCacheKeyIsInjective(t *testing.T) {
	t.Parallel()

	a := domain.GetImageURLsByChapterOpts{SeriesSlug: "one", ChapterID: "piece"}
	b := domain.GetImageURLsByChapterOpts{SeriesSlug: "one piece"}

	if a.CacheKey() == b.CacheKey() {
		t.Errorf("CacheKey() = %q pour deux options différentes, want distinctes", a.CacheKey())
	}
}
