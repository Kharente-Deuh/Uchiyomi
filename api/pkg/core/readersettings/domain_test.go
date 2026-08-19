// SPDX-License-Identifier: AGPL-3.0-or-later

package readersettings_test

import (
	"errors"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readersettings"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

func TestAllTypesOrder(t *testing.T) {
	t.Parallel()

	got := readersettings.AllTypes()
	want := []sources.SeriesType{
		sources.SeriesTypeManga,
		sources.SeriesTypeManhua,
		sources.SeriesTypeManhwa,
		sources.SeriesTypeMangatoon,
	}

	if len(got) != len(want) {
		t.Fatalf("AllTypes() len = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("AllTypes()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDefaultProfile(t *testing.T) {
	t.Parallel()

	tests := map[sources.SeriesType]readersettings.Profile{
		sources.SeriesTypeManga: {
			Type:        sources.SeriesTypeManga,
			ReadingMode: readersettings.ReadingModePagedRTL,
			PageScale:   readersettings.PageScaleFitScreen,
			DoublePage:  false,
		},
		sources.SeriesTypeManhua: {
			Type:        sources.SeriesTypeManhua,
			ReadingMode: readersettings.ReadingModePagedLTR,
			PageScale:   readersettings.PageScaleFitScreen,
			DoublePage:  false,
		},
		sources.SeriesTypeManhwa: {
			Type:        sources.SeriesTypeManhwa,
			ReadingMode: readersettings.ReadingModeWebtoon,
			PageScale:   readersettings.PageScaleFitWidth,
			DoublePage:  false,
		},
		sources.SeriesTypeMangatoon: {
			Type:        sources.SeriesTypeMangatoon,
			ReadingMode: readersettings.ReadingModeWebtoon,
			PageScale:   readersettings.PageScaleFitWidth,
			DoublePage:  false,
		},
	}

	for typ, want := range tests {
		t.Run(string(typ), func(t *testing.T) {
			t.Parallel()

			got := readersettings.DefaultProfile(typ)
			if got != want {
				t.Errorf("DefaultProfile(%q) = %+v, want %+v", typ, got, want)
			}
		})
	}
}

func TestParseReadingMode(t *testing.T) {
	t.Parallel()

	ok := map[string]readersettings.ReadingMode{
		"paged-rtl": readersettings.ReadingModePagedRTL,
		"paged-ltr": readersettings.ReadingModePagedLTR,
		"webtoon":   readersettings.ReadingModeWebtoon,
	}

	for in, want := range ok {
		got, err := readersettings.ParseReadingMode(in)
		if err != nil {
			t.Fatalf("ParseReadingMode(%q): %v", in, err)
		}

		if got != want {
			t.Errorf("ParseReadingMode(%q) = %q, want %q", in, got, want)
		}
	}

	_, err := readersettings.ParseReadingMode("scroll")
	if !errors.Is(err, readersettings.ErrInvalid) {
		t.Errorf("ParseReadingMode(scroll) err = %v, want ErrInvalid", err)
	}
}

func TestParsePageScale(t *testing.T) {
	t.Parallel()

	ok := map[string]readersettings.PageScale{
		"fit-width":  readersettings.PageScaleFitWidth,
		"fit-height": readersettings.PageScaleFitHeight,
		"fit-screen": readersettings.PageScaleFitScreen,
	}

	for in, want := range ok {
		got, err := readersettings.ParsePageScale(in)
		if err != nil {
			t.Fatalf("ParsePageScale(%q): %v", in, err)
		}

		if got != want {
			t.Errorf("ParsePageScale(%q) = %q, want %q", in, got, want)
		}
	}

	_, err := readersettings.ParsePageScale("original")
	if !errors.Is(err, readersettings.ErrInvalid) {
		t.Errorf("ParsePageScale(original) err = %v, want ErrInvalid", err)
	}
}
