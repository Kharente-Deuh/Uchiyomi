// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
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

func TestDiskPagesOpenPage(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	comicID := uuid.New()
	pageDir := chapterDir(dir, comicID, 1.5)
	if err := os.MkdirAll(pageDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	path := filepath.Join(pageDir, "002.png")
	if err := os.WriteFile(path, []byte("img"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	gotPath, contentType, err := DiskPages{Dir: dir}.OpenPage(comicID, 1.5, 2)
	if err != nil {
		t.Fatalf("OpenPage: %v", err)
	}

	if gotPath != path {
		t.Errorf("path = %q, want %q", gotPath, path)
	}

	if contentType != "image/png" {
		t.Errorf("contentType = %q", contentType)
	}
}

func TestDiskPagesOpenPageMissing(t *testing.T) {
	t.Parallel()

	_, _, err := DiskPages{Dir: t.TempDir()}.OpenPage(uuid.New(), 1, 1)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("OpenPage = %v, want ErrNotFound", err)
	}
}
