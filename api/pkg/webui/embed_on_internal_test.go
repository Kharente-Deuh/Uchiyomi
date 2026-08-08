// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build webui

package webui

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmbeddedFSKeepsTheHashedBundles(t *testing.T) {
	sub, err := fs.Sub(assets, "dist")
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}

	entries, err := fs.ReadDir(sub, "_nuxt")
	if err != nil {
		t.Fatalf("lecture de _nuxt: %v", err)
	}

	if len(entries) == 0 {
		t.Error("_nuxt est vide")
	}
}

func TestHandlerServesTheEmbeddedIndex(t *testing.T) {
	h, ok := Handler()
	if !ok {
		t.Fatal("Handler rend false alors que le tag webui est posé")
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("code = %d, attendu %d", rec.Code, http.StatusOK)
	}
}
