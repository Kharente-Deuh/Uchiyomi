// SPDX-License-Identifier: AGPL-3.0-or-later

package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	healthhttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/health/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/health"
)

type componentBody struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type readyzBody struct {
	Components map[string]componentBody `json:"components"`
	Status     string                   `json:"status"`
}

func newRouter(t *testing.T, reg *health.Registry) chi.Router {
	t.Helper()

	r, _ := newRouterWithLogs(t, reg)

	return r
}

func newRouterWithLogs(t *testing.T, reg *health.Registry) (chi.Router, *bytes.Buffer) {
	t.Helper()

	var logs bytes.Buffer

	ctrl, err := healthhttp.New(
		healthhttp.Config{ProbeTimeout: healthhttp.DefaultProbeTimeout},
		healthhttp.Deps{
			Registry: reg,
			Logger:   slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
		},
	)
	if err != nil {
		t.Fatalf("healthhttp.New: %v", err)
	}

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	return r, &logs
}

func get(t *testing.T, r chi.Router, path string) (int, readyzBody) {
	t.Helper()

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

	var body readyzBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("décodage de la réponse: %v", err)
	}

	return rec.Code, body
}

func TestNewRejectsMissingRegistry(t *testing.T) {
	_, err := healthhttp.New(
		healthhttp.Config{ProbeTimeout: time.Second},
		healthhttp.Deps{Logger: slog.New(slog.DiscardHandler)},
	)
	if err == nil {
		t.Fatal("New: erreur attendue quand Registry est nil")
	}
}

func TestNewRejectsNonPositiveProbeTimeout(t *testing.T) {
	_, err := healthhttp.New(
		healthhttp.Config{},
		healthhttp.Deps{Registry: health.NewRegistry(), Logger: slog.New(slog.DiscardHandler)},
	)
	if err == nil {
		t.Fatal("New: erreur attendue quand ProbeTimeout est nul")
	}
}

func TestHealthzIs200EvenWhenNotReady(t *testing.T) {
	reg := health.NewRegistry("migrations")
	reg.Set("migrations", errors.New("boom"))

	code, body := get(t, newRouter(t, reg), "/healthz")

	if code != http.StatusOK {
		t.Fatalf("code = %d, attendu %d", code, http.StatusOK)
	}

	if body.Status != "ok" {
		t.Errorf("status = %q, attendu %q", body.Status, "ok")
	}
}

func TestReadyzIs200WhenEverythingIsOK(t *testing.T) {
	reg := health.NewRegistry("migrations")
	reg.AddProbe("db", func(context.Context) error { return nil })
	reg.Set("migrations", nil)

	code, body := get(t, newRouter(t, reg), "/readyz")

	if code != http.StatusOK {
		t.Fatalf("code = %d, attendu %d", code, http.StatusOK)
	}

	if body.Status != "ok" {
		t.Errorf("status = %q, attendu %q", body.Status, "ok")
	}

	if got := body.Components["db"].Status; got != "ok" {
		t.Errorf("db = %q, attendu %q", got, "ok")
	}
}

func TestReadyzIs503WhileStartingAndNamesWhatIsMissing(t *testing.T) {
	reg := health.NewRegistry("migrations", "asura")
	reg.Set("asura", nil)

	code, body := get(t, newRouter(t, reg), "/readyz")

	if code != http.StatusServiceUnavailable {
		t.Fatalf("code = %d, attendu %d", code, http.StatusServiceUnavailable)
	}

	if body.Status != "starting" {
		t.Errorf("status = %q, attendu %q", body.Status, "starting")
	}

	if got := body.Components["migrations"].Status; got != "starting" {
		t.Errorf("migrations = %q, attendu %q", got, "starting")
	}

	if got := body.Components["asura"].Status; got != "ok" {
		t.Errorf("asura = %q, attendu %q", got, "ok")
	}
}

func TestReadyzIs503WithLatchReasonVerbatim(t *testing.T) {
	reg := health.NewRegistry("migrations")
	reg.Set("migrations", errors.New("colonne inconnue"))

	code, body := get(t, newRouter(t, reg), "/readyz")

	if code != http.StatusServiceUnavailable {
		t.Fatalf("code = %d, attendu %d", code, http.StatusServiceUnavailable)
	}

	if body.Status != "failed" {
		t.Errorf("status = %q, attendu %q", body.Status, "failed")
	}

	if got := body.Components["migrations"].Reason; got != "colonne inconnue" {
		t.Errorf("raison = %q, attendu %q", got, "colonne inconnue")
	}
}

func TestReadyzReplacesProbeReasonWithStableLabel(t *testing.T) {
	driverMsg := "db.PingContext: failed to connect to `user=uchiyomi database=uchiyomi`: 172.18.0.2:5432"

	reg := health.NewRegistry()
	reg.AddProbe("db", func(context.Context) error { return errors.New(driverMsg) })

	rec := httptest.NewRecorder()
	newRouter(t, reg).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("code = %d, attendu %d", rec.Code, http.StatusServiceUnavailable)
	}

	var body readyzBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("décodage: %v", err)
	}

	if got := body.Components["db"].Reason; got != "unreachable" {
		t.Errorf("raison = %q, attendu %q", got, "unreachable")
	}

	for _, secret := range []string{"uchiyomi", "172.18.0.2", "PingContext"} {
		if strings.Contains(rec.Body.String(), secret) {
			t.Errorf("le corps divulgue %q: %s", secret, rec.Body.String())
		}
	}
}

func TestReadyzLogsFullProbeReason(t *testing.T) {
	driverMsg := "db.PingContext: failed to connect to `user=uchiyomi database=uchiyomi`"

	reg := health.NewRegistry()
	reg.AddProbe("db", func(context.Context) error { return errors.New(driverMsg) })

	r, logs := newRouterWithLogs(t, reg)
	get(t, r, "/readyz")

	if !strings.Contains(logs.String(), driverMsg) {
		t.Errorf("raison complète absente des logs: %s", logs.String())
	}

	if !strings.Contains(logs.String(), `"probe":"db"`) {
		t.Errorf("nom de la sonde absent des logs: %s", logs.String())
	}
}

func TestReadyzDoesNotLogLatchFailures(t *testing.T) {
	reg := health.NewRegistry("migrations")
	reg.Set("migrations", errors.New("colonne inconnue"))

	r, logs := newRouterWithLogs(t, reg)
	get(t, r, "/readyz")

	if logs.Len() != 0 {
		t.Errorf("aucun log attendu, obtenu: %s", logs.String())
	}
}

func TestReadyzAppliesProbeTimeout(t *testing.T) {
	reg := health.NewRegistry()
	reg.AddProbe("db", func(ctx context.Context) error {
		<-ctx.Done()

		//nolint:wrapcheck
		return ctx.Err()
	})

	ctrl, err := healthhttp.New(
		healthhttp.Config{ProbeTimeout: 10 * time.Millisecond},
		healthhttp.Deps{Registry: reg, Logger: slog.New(slog.DiscardHandler)},
	)
	if err != nil {
		t.Fatalf("healthhttp.New: %v", err)
	}

	r := chi.NewRouter()
	ctrl.InitRouter(r)

	code, body := get(t, r, "/readyz")

	if code != http.StatusServiceUnavailable {
		t.Fatalf("code = %d, attendu %d", code, http.StatusServiceUnavailable)
	}

	if got := body.Components["db"].Status; got != "failed" {
		t.Errorf("db = %q, attendu %q", got, "failed")
	}
}

func TestReadyzOmitsEmptyReason(t *testing.T) {
	reg := health.NewRegistry("migrations")
	reg.Set("migrations", nil)

	rec := httptest.NewRecorder()
	newRouter(t, reg).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	var raw struct {
		Components map[string]map[string]any `json:"components"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatalf("décodage: %v", err)
	}

	if _, ok := raw.Components["migrations"]["reason"]; ok {
		t.Error("reason présent sur un composant sain")
	}
}
