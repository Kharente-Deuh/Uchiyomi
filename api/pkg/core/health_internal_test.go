// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/health"
)

const gatePort = 3000

type messageBody struct {
	Message string `json:"message"`
}

func newTestHandler(t *testing.T, db Database) (http.Handler, *health.Registry) {
	t.Helper()

	app, registry := newTestApp(t, db, gatePort)

	return app.initServer().Handler, registry
}

func serve(t *testing.T, h http.Handler, path string) (int, []byte) {
	t.Helper()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

	return rec.Code, rec.Body.Bytes()
}

func TestNewHealthRegistryDeclaresTheLatchesAndTheDBProbe(t *testing.T) {
	reg := NewHealthRegistry(&fakeDB{})

	rep := reg.Snapshot(context.Background())

	for _, name := range []string{componentMigrations, componentAsura, componentSessions} {
		c, ok := rep.Components[name]
		if !ok {
			t.Errorf("latche %s absente du registre", name)

			continue
		}

		if c.Status != health.StatusStarting {
			t.Errorf("%s = %q, attendu %q", name, c.Status, health.StatusStarting)
		}

		if c.Probe {
			t.Errorf("%s est déclaré comme sonde", name)
		}
	}

	if !rep.Components[componentDB].Probe {
		t.Error("db n'est pas déclaré comme sonde")
	}
}

func TestNewHealthRegistryWiresTheProbeToTheGivenDatabase(t *testing.T) {
	boom := errors.New("connexion refusée")
	reg := NewHealthRegistry(&fakeDB{ping: func(context.Context) error { return boom }})

	c := reg.Snapshot(context.Background()).Components[componentDB]
	if c.Status != health.StatusFailed {
		t.Fatalf("db = %q, attendu %q", c.Status, health.StatusFailed)
	}

	if c.Reason != boom.Error() {
		t.Errorf("raison = %q, attendu %q", c.Reason, boom.Error())
	}
}

func TestApplicationRoutesAre503WhileMigrationsAreStarting(t *testing.T) {
	h, _ := newTestHandler(t, &fakeDB{})

	for _, path := range []string{"/api/setup/status", "/api/sources/asura/search"} {
		code, raw := serve(t, h, path)
		if code != http.StatusServiceUnavailable {
			t.Errorf("%s: code = %d, attendu %d", path, code, http.StatusServiceUnavailable)

			continue
		}

		var body messageBody
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Errorf("%s: décodage de %q: %v", path, raw, err)

			continue
		}

		if body.Message != notReadyMessage {
			t.Errorf("%s: message = %q, attendu %q", path, body.Message, notReadyMessage)
		}
	}
}

func TestApplicationRoutesRespondOnceMigrationsAreOK(t *testing.T) {
	h, reg := newTestHandler(t, &fakeDB{})

	reg.Set(componentMigrations, nil)

	code, raw := serve(t, h, "/api/setup/status")
	if code != http.StatusOK {
		t.Fatalf("code = %d, attendu %d (corps: %s)", code, http.StatusOK, raw)
	}

	var body struct {
		Required bool `json:"required"`
	}

	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("décodage de %q: %v", raw, err)
	}

	if body.Required {
		t.Error("required = true, attendu la réponse du service feint")
	}
}

func TestHealthRoutesAnswerWhileTheGateIsClosed(t *testing.T) {
	h, _ := newTestHandler(t, &fakeDB{})

	if code, _ := serve(t, h, "/healthz"); code != http.StatusOK {
		t.Errorf("/healthz: code = %d, attendu %d", code, http.StatusOK)
	}

	if code, _ := serve(t, h, "/readyz"); code != http.StatusServiceUnavailable {
		t.Errorf("/readyz: code = %d, attendu %d", code, http.StatusServiceUnavailable)
	}
}

func TestRequireMigrationsDoesNotRunTheProbe(t *testing.T) {
	var pings int

	db := &fakeDB{ping: func(context.Context) error {
		pings++

		return nil
	}}

	h, reg := newTestHandler(t, db)
	reg.Set(componentMigrations, nil)

	if code, raw := serve(t, h, "/api/setup/status"); code != http.StatusOK {
		t.Fatalf("code = %d, attendu %d (corps: %s)", code, http.StatusOK, raw)
	}

	if pings != 0 {
		t.Errorf("sonde db appelée %d fois pour une requête applicative, attendu 0", pings)
	}
}
