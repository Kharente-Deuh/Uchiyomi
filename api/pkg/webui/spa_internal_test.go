// SPDX-License-Identifier: AGPL-3.0-or-later

package webui

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func testFS() fstest.MapFS {
	return fstest.MapFS{
		indexFile:             {Data: []byte("<!doctype html>index")},
		"favicon.ico":         {Data: []byte("icon")},
		"_nuxt/app.abc123.js": {Data: []byte("bundle")},
	}
}

func serve(t *testing.T, fsys fstest.MapFS, method, target string) *http.Response {
	t.Helper()

	rec := httptest.NewRecorder()
	newSPA(fsys).ServeHTTP(rec, httptest.NewRequest(method, target, nil))

	return rec.Result()
}

func TestSPAServesIndexAtRoot(t *testing.T) {
	res := serve(t, testFS(), http.MethodGet, "/")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("code = %d, want %d", res.StatusCode, http.StatusOK)
	}
}

func TestSPAServesExistingAsset(t *testing.T) {
	res := serve(t, testFS(), http.MethodGet, "/_nuxt/app.abc123.js")
	if res.StatusCode != http.StatusOK {
		t.Fatalf("code = %d, want %d", res.StatusCode, http.StatusOK)
	}

	if got := res.Header.Get("Cache-Control"); got != immutableCache {
		t.Errorf("Cache-Control = %q, want %q", got, immutableCache)
	}
}

func TestSPADoesNotCacheIndex(t *testing.T) {
	res := serve(t, testFS(), http.MethodGet, "/")
	if got := res.Header.Get("Cache-Control"); got != "no-cache" {
		t.Errorf("Cache-Control = %q, want %q", got, "no-cache")
	}
}

func TestSPAFallsBackToIndexForClientRoutes(t *testing.T) {
	for _, target := range []string{"/library", "/settings/sources", "/status"} {
		res := serve(t, testFS(), http.MethodGet, target)
		if res.StatusCode != http.StatusOK {
			t.Errorf("%s: code = %d, want %d", target, res.StatusCode, http.StatusOK)
		}
	}
}

func TestSPAReturns404ForMissingAsset(t *testing.T) {
	for _, target := range []string{"/_nuxt/disparu.js", "/logo.png", "/manifest.webmanifest"} {
		res := serve(t, testFS(), http.MethodGet, target)
		if res.StatusCode != http.StatusNotFound {
			t.Errorf("%s: code = %d, want %d", target, res.StatusCode, http.StatusNotFound)
		}
	}
}

func TestSPAReturns404WhenIndexIsMissing(t *testing.T) {
	fsys := fstest.MapFS{"favicon.ico": {Data: []byte("icon")}}

	res := serve(t, fsys, http.MethodGet, "/library")
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("code = %d, want %d", res.StatusCode, http.StatusNotFound)
	}
}

func TestSPARejectsPathTraversal(t *testing.T) {
	fsys := testFS()
	fsys["secret.txt"] = &fstest.MapFile{Data: []byte("secret")}

	for _, target := range []string{"/_nuxt/../secret.txt", "/../../../etc/passwd"} {
		res := serve(t, fsys, http.MethodGet, target)
		if res.StatusCode == http.StatusOK {
			t.Errorf("%s: code = %d, path traversal must never succeed", target, res.StatusCode)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("%s: body read: %v", target, err)
		}

		if strings.Contains(string(body), "secret") {
			t.Errorf("%s: body serves targeted file", target)
		}
	}
}

func TestSPARejectsNonReadMethods(t *testing.T) {
	res := serve(t, testFS(), http.MethodPost, "/")
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("code = %d, want %d", res.StatusCode, http.StatusMethodNotAllowed)
	}
}
