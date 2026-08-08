// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/health"
)

func TestRunComponentMarksOKBeforeEnteringTheLoop(t *testing.T) {
	reg := health.NewRegistry(componentAsura)
	a := &App{deps: Deps{Health: reg}}

	var seen health.Status

	run := func(ctx context.Context) error {
		seen = reg.Snapshot(ctx).Components[componentAsura].Status

		return nil
	}

	if err := a.runComponent(context.Background(), componentAsura, run)(); err != nil {
		t.Fatalf("runComponent: %v", err)
	}

	if seen != health.StatusOK {
		t.Fatalf("statut vu depuis la boucle = %q, attendu %q", seen, health.StatusOK)
	}
}

func TestRunComponentMarksFailedWhenLoopReturnsError(t *testing.T) {
	wantErr := errors.New("boucle morte")

	reg := health.NewRegistry(componentSessions)
	a := &App{deps: Deps{Health: reg}}

	run := func(context.Context) error { return wantErr }

	err := a.runComponent(context.Background(), componentSessions, run)()
	if err == nil {
		t.Fatal("runComponent: erreur attendue")
	}

	if !errors.Is(err, wantErr) {
		t.Errorf("erreur %v, attendu qu'elle encapsule %v", err, wantErr)
	}

	c := reg.Snapshot(context.Background()).Components[componentSessions]
	if c.Status != health.StatusFailed {
		t.Errorf("statut = %q, attendu %q", c.Status, health.StatusFailed)
	}

	if c.Reason != wantErr.Error() {
		t.Errorf("raison = %q, attendu %q", c.Reason, wantErr.Error())
	}
}

func TestRunComponentStaysOKOnCleanReturn(t *testing.T) {
	reg := health.NewRegistry(componentAsura)
	a := &App{deps: Deps{Health: reg}}

	run := func(context.Context) error { return nil }

	if err := a.runComponent(context.Background(), componentAsura, run)(); err != nil {
		t.Fatalf("runComponent: %v", err)
	}

	if got := reg.Snapshot(context.Background()).Components[componentAsura].Status; got != health.StatusOK {
		t.Fatalf("statut = %q, attendu %q", got, health.StatusOK)
	}
}

func TestRunServesReadyzWhileMigrationIsRunning(t *testing.T) {
	app, _ := startBlockedApp(t)
	app.waitListening(t)

	code, raw := app.get(t, "/readyz")
	if code != http.StatusServiceUnavailable {
		t.Fatalf("code = %d, attendu %d", code, http.StatusServiceUnavailable)
	}

	body := decodeReadyz(t, raw)
	if got := body.Components[componentMigrations].Status; got != string(health.StatusStarting) {
		t.Errorf("migrations = %q, attendu %q", got, health.StatusStarting)
	}
}

func TestRunReportsEverythingOKOnceMigrationCompletes(t *testing.T) {
	app, unblock := startBlockedApp(t)
	app.waitListening(t)
	unblock()

	var raw []byte

	app.waitFor(t, "/readyz n'est jamais passé à 200 après la migration", func() bool {
		code, body := app.get(t, "/readyz")
		raw = body

		return code == http.StatusOK
	})

	body := decodeReadyz(t, raw)
	for _, name := range []string{componentMigrations, componentAsura, componentSessions, componentDB} {
		if got := body.Components[name].Status; got != string(health.StatusOK) {
			t.Errorf("%s = %q, attendu %q", name, got, health.StatusOK)
		}
	}
}

func TestRunReturnsMigrationErrorAndNotContextCanceled(t *testing.T) {
	boom := errors.New("colonne inconnue")
	db := &fakeDB{migrate: func() error { return boom }}

	err := startApp(t, db).wait(t)
	if !errors.Is(err, boom) {
		t.Fatalf("erreur = %v, attendu qu'elle encapsule %v", err, boom)
	}

	if errors.Is(err, context.Canceled) {
		t.Errorf("erreur = %v : main la prendrait pour un arrêt propre et sortirait en 0", err)
	}
}
