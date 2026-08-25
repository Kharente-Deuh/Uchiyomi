// SPDX-License-Identifier: AGPL-3.0-or-later

package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/parse"
)

func TestParsePageURLs(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("testdata", "reader.html"))
	if err != nil {
		t.Fatal(err)
	}

	got := parse.ParsePageURLs(string(raw))

	want := []string{
		"https://cdn.example/page1.jpg",
		"https://cdn.example/page2.jpg",
		"https://cdn.example/page3.jpg",
		"https://cdn.example/page6.jpg",
	}

	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %v", len(got), len(want), got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("urls[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParsePageURLsInvalidHTML(t *testing.T) {
	t.Parallel()

	got := parse.ParsePageURLs("<html><body><div id=\"readerarea\">")

	if got == nil {
		t.Fatal("fragment HTML still parses; got nil, want empty non-nil slice")
	}

	if len(got) != 0 {
		t.Fatalf("got = %v, want empty", got)
	}
}

func TestParsePageURLsEmptyReaderArea(t *testing.T) {
	t.Parallel()

	got := parse.ParsePageURLs(`<html><body><div id="readerarea"></div></body></html>`)

	if got == nil {
		t.Fatal("got nil, want empty non-nil slice")
	}

	if len(got) != 0 {
		t.Fatalf("got = %v, want empty", got)
	}
}
