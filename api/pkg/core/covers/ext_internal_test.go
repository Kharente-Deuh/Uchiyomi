// SPDX-License-Identifier: AGPL-3.0-or-later

package covers

import "testing"

func TestCacheKeyAndParse(t *testing.T) {
	t.Parallel()

	key := cacheKey("asurascans", "solo-leveling", ".webp")
	if key != "asurascans/solo-leveling.webp" {
		t.Fatalf("cacheKey = %q", key)
	}

	source, slug, ext, err := parseCacheKey(key)
	if err != nil {
		t.Fatalf("parseCacheKey: %v", err)
	}

	if source != "asurascans" || slug != "solo-leveling" || ext != ".webp" {
		t.Errorf("parsed = %q %q %q", source, slug, ext)
	}
}

func TestParseCacheKeyInvalid(t *testing.T) {
	t.Parallel()

	if _, _, _, err := parseCacheKey(""); err == nil {
		t.Error("empty key must fail")
	}

	if _, _, _, err := parseCacheKey("noslash"); err == nil {
		t.Error("key without slash must fail")
	}
}

func TestParseCacheKeyWithoutExtension(t *testing.T) {
	t.Parallel()

	source, slug, ext, err := parseCacheKey("asurascans/solo-leveling")
	if err != nil {
		t.Fatalf("parseCacheKey: %v", err)
	}

	if source != "asurascans" || slug != "solo-leveling" || ext != "" {
		t.Errorf("parsed = %q %q %q", source, slug, ext)
	}
}
