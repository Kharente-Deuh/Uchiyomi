// SPDX-License-Identifier: AGPL-3.0-or-later

package comics_test

import (
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
)

func TestParseListSort(t *testing.T) {
	t.Parallel()

	got, err := comics.ParseListSort("title")
	if err != nil || got != comics.ListSortTitle {
		t.Fatalf("title: got %q err %v", got, err)
	}

	got, err = comics.ParseListSort("addedAt")
	if err != nil || got != comics.ListSortAddedAt {
		t.Fatalf("addedAt: got %q err %v", got, err)
	}

	if _, err = comics.ParseListSort("created_at"); err == nil {
		t.Fatal("created_at must be invalid")
	}
}

func TestParseListOrder(t *testing.T) {
	t.Parallel()

	got, err := comics.ParseListOrder("asc")
	if err != nil || got != comics.ListOrderAsc {
		t.Fatalf("asc: got %q err %v", got, err)
	}

	got, err = comics.ParseListOrder("desc")
	if err != nil || got != comics.ListOrderDesc {
		t.Fatalf("desc: got %q err %v", got, err)
	}

	if _, err = comics.ParseListOrder("ASC"); err == nil {
		t.Fatal("ASC must be invalid")
	}
}
