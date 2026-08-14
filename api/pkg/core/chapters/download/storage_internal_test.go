// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"testing"
)

func TestProgressPercent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		downloaded int
		pagesNb    int
		want       int
	}{
		{name: "zero pages", downloaded: 0, pagesNb: 0, want: 0},
		{name: "half", downloaded: 1, pagesNb: 2, want: 50},
		{name: "complete", downloaded: 3, pagesNb: 3, want: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := progressPercent(tt.downloaded, tt.pagesNb); got != tt.want {
				t.Fatalf("progressPercent(%d, %d) = %d, want %d", tt.downloaded, tt.pagesNb, got, tt.want)
			}
		})
	}
}

func TestParsePageIndex(t *testing.T) {
	t.Parallel()

	index, ok := parsePageIndex("001.webp")
	if !ok || index != 1 {
		t.Fatalf("parsePageIndex(001.webp) = (%d, %v), want (1, true)", index, ok)
	}

	if _, ok := parsePageIndex("bad-name.webp"); ok {
		t.Fatal("parsePageIndex accepted invalid file name")
	}
}
