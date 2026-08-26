// SPDX-License-Identifier: AGPL-3.0-or-later

package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/parse"
)

const fixtureSeriesTitle = "Tears on a Withered Flower"

func TestParseSearch(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("testdata", "search.html"))
	if err != nil {
		t.Fatal(err)
	}

	page, err := parse.ParseSearch(string(raw))
	if err != nil {
		t.Fatal(err)
	}

	if !page.HasNext {
		t.Fatal("HasNext = false, want true")
	}

	if len(page.Items) != 2 {
		t.Fatalf("len = %d", len(page.Items))
	}

	if page.Items[0].Slug != "tears-on-a-withered-flower" || page.Items[0].Title != fixtureSeriesTitle {
		t.Fatalf("item0 %+v", page.Items[0])
	}

	if page.Items[0].LastChapter != 115 || page.Items[0].Cover != "https://cdn.example/cover1.jpg" {
		t.Fatalf("item0 meta %+v", page.Items[0])
	}

	if page.Items[1].Slug != "example-manga" || page.Items[1].LastChapter != 111.5 {
		t.Fatalf("item1 %+v", page.Items[1])
	}

	if page.Items[1].Cover != "https://cdn.example/cover2.jpg" {
		t.Fatalf("cover2 %q", page.Items[1].Cover)
	}
}

func TestParseSearchWithoutNext(t *testing.T) {
	t.Parallel()

	page, err := parse.ParseSearch(`<!DOCTYPE html><html><body>
<div class="listupd"><div class="bsx">
  <a href="/manga/last-one/"><div class="tt">Last</div></a>
</div></div>
<div class="hpage"></div>
</body></html>`)
	if err != nil {
		t.Fatal(err)
	}

	if page.HasNext {
		t.Fatal("HasNext = true, want false")
	}

	if len(page.Items) != 1 {
		t.Fatalf("len = %d", len(page.Items))
	}
}
