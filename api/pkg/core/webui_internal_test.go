// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"net/http"
	"testing"
)

const uiBody = "je suis le front"

func newRouterWithUI(t *testing.T, db Database) http.Handler {
	t.Helper()

	app, registry := newTestApp(t, db, gatePort)
	registry.Set(componentMigrations, nil)

	ui := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(uiBody)); err != nil {
			t.Errorf("écriture du front factice: %v", err)
		}
	})

	return app.newRouter(ui)
}

func TestRouterServesTheUIOnUnknownRootPaths(t *testing.T) {
	h := newRouterWithUI(t, &fakeDB{})

	for _, path := range []string{"/", "/library", "/settings/sources"} {
		code, raw := serve(t, h, path)
		if code != http.StatusOK {
			t.Errorf("%s: code = %d, attendu %d", path, code, http.StatusOK)

			continue
		}

		if string(raw) != uiBody {
			t.Errorf("%s: corps = %q, attendu le front", path, raw)
		}
	}
}

func TestRouterNeverServesTheUIUnderTheAPIPrefix(t *testing.T) {
	h := newRouterWithUI(t, &fakeDB{})

	for _, path := range []string{"/api", "/api/inconnu", "/api/setup/inconnu", "/api/sources/inconnu"} {
		code, raw := serve(t, h, path)
		if code != http.StatusNotFound {
			t.Errorf("%s: code = %d, attendu %d", path, code, http.StatusNotFound)
		}

		if string(raw) == uiBody {
			t.Errorf("%s: le fallback SPA a avalé une route API", path)
		}
	}
}

func TestRouterServesHealthOnBothPrefixes(t *testing.T) {
	h := newRouterWithUI(t, &fakeDB{})

	for _, path := range []string{"/healthz", "/api/healthz"} {
		if code, _ := serve(t, h, path); code != http.StatusOK {
			t.Errorf("%s: code = %d, attendu %d", path, code, http.StatusOK)
		}
	}
}

func TestRouterServesTheAPIUnderThePrefix(t *testing.T) {
	h := newRouterWithUI(t, &fakeDB{})

	code, raw := serve(t, h, "/api/setup/status")
	if code != http.StatusOK {
		t.Fatalf("code = %d, attendu %d (corps: %s)", code, http.StatusOK, raw)
	}
}

func TestRouterWithoutUIReturns404AtRoot(t *testing.T) {
	app, _ := newTestApp(t, &fakeDB{}, gatePort)

	if code, _ := serve(t, app.newRouter(nil), "/library"); code != http.StatusNotFound {
		t.Errorf("code = %d, attendu %d", code, http.StatusNotFound)
	}
}
