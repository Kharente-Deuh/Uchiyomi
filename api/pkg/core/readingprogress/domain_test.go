// SPDX-License-Identifier: AGPL-3.0-or-later

package readingprogress_test

import (
	"errors"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readingprogress"
)

func TestClampPage(t *testing.T) {
	t.Parallel()

	if got := readingprogress.ClampPage(18, 12); got != 12 {
		t.Errorf("ClampPage(18, 12) = %d, want 12", got)
	}

	if got := readingprogress.ClampPage(5, 12); got != 5 {
		t.Errorf("ClampPage(5, 12) = %d, want 5", got)
	}

	if got := readingprogress.ClampPage(3, 0); got != 3 {
		t.Errorf("ClampPage(3, 0) = %d, want 3", got)
	}
}

func TestMergePageFurthest(t *testing.T) {
	t.Parallel()

	stored := 20
	got, err := readingprogress.MergePage(&stored, 8, 20)
	if err != nil {
		t.Fatalf("MergePage: %v", err)
	}

	if got != 20 {
		t.Errorf("MergePage stored 20 incoming 8 = %d, want 20", got)
	}
}

func TestMergePageFirstSave(t *testing.T) {
	t.Parallel()

	got, err := readingprogress.MergePage(nil, 5, 10)
	if err != nil {
		t.Fatalf("MergePage: %v", err)
	}

	if got != 5 {
		t.Errorf("MergePage first save = %d, want 5", got)
	}
}

func TestMergePageRejectsAbovePagesNb(t *testing.T) {
	t.Parallel()

	_, err := readingprogress.MergePage(nil, 16, 15)
	if !errors.Is(err, readingprogress.ErrInvalid) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}
}

func TestMergePageAllowsWhenPagesNbZero(t *testing.T) {
	t.Parallel()

	got, err := readingprogress.MergePage(nil, 3, 0)
	if err != nil {
		t.Fatalf("MergePage: %v", err)
	}

	if got != 3 {
		t.Errorf("got %d, want 3", got)
	}
}

func TestMergePageRejectsPageBelowOne(t *testing.T) {
	t.Parallel()

	_, err := readingprogress.MergePage(nil, 0, 10)
	if !errors.Is(err, readingprogress.ErrInvalid) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}
}

func TestMergePagePersistsClampWhenStoredAbovePagesNb(t *testing.T) {
	t.Parallel()

	stored := 18
	got, err := readingprogress.MergePage(&stored, 10, 12)
	if err != nil {
		t.Fatalf("MergePage: %v", err)
	}

	if got != 12 {
		t.Errorf("furthest 18 then clamp to 12 = %d, want 12", got)
	}
}
