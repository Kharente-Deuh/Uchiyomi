// SPDX-License-Identifier: AGPL-3.0-or-later

package covers_test

import (
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
)

func TestExtensionFromURL(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"https://cdn.example/cover.webp":     ".webp",
		"https://cdn.example/path/cover.jpg": ".jpg",
		"https://cdn.example/noext":          "",
	}

	for raw, want := range tests {
		if got := covers.ExtensionFromURL(raw); got != want {
			t.Errorf("ExtensionFromURL(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestResolveAbsoluteURL(t *testing.T) {
	t.Parallel()

	base := "https://gg.asuracomic.net"

	if got := covers.ResolveAbsoluteURL(base, "cover.jpg"); got != base+"/cover.jpg" {
		t.Errorf("relative = %q", got)
	}

	if got := covers.ResolveAbsoluteURL(base, "/storage/cover.webp"); got != base+"/storage/cover.webp" {
		t.Errorf("absolute path = %q", got)
	}

	external := "https://cdn.example/cover.webp"
	if got := covers.ResolveAbsoluteURL("https://gg.asuracomic.net", external); got != external {
		t.Errorf("absolute url = %q", got)
	}
}
