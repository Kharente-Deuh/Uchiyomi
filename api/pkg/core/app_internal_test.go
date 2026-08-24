// SPDX-License-Identifier: AGPL-3.0-or-later

//nolint:goconst,lll
package core

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/health"
)

func TestRouterMountsTheOIDCProvidersRoutes(t *testing.T) {
	t.Parallel()

	app, _ := newTestApp(t, &fakeDB{}, gatePort)

	req := httptest.NewRequest(http.MethodGet, "/api/oidc/providers", nil)
	rec := httptest.NewRecorder()

	app.newRouter(nil).ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Errorf("GET /api/oidc/providers = 404, want the route to be mounted")
	}
}

func TestRouterMountsTheFeedRoute(t *testing.T) {
	t.Parallel()

	app, _ := newTestApp(t, &fakeDB{}, gatePort)

	req := httptest.NewRequest(http.MethodGet, "/api/feed/", nil)
	rec := httptest.NewRecorder()

	app.newRouter(nil).ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Errorf("GET /api/feed/ = 404, want the route to be mounted")
	}
}

func TestRouterMountsTheReaderSettingsRoute(t *testing.T) {
	t.Parallel()

	app, _ := newTestApp(t, &fakeDB{}, gatePort)

	req := httptest.NewRequest(http.MethodGet, "/api/me/reader-settings", nil)
	rec := httptest.NewRecorder()

	app.newRouter(nil).ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Errorf("GET /api/me/reader-settings = 404, want the route to be mounted")
	}
}

func TestRouterMountsTheChapterProgressRoute(t *testing.T) {
	t.Parallel()

	app, _ := newTestApp(t, &fakeDB{}, gatePort)

	req := httptest.NewRequest(http.MethodPut, "/api/chapters/"+uuid.Nil.String()+"/progress", nil)
	rec := httptest.NewRecorder()

	app.newRouter(nil).ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Errorf("PUT /api/chapters/{id}/progress = 404, want the route to be mounted")
	}
}

func TestRouterMountsTheChapterByIDRoute(t *testing.T) {
	t.Parallel()

	app, _ := newTestApp(t, &fakeDB{}, gatePort)

	req := httptest.NewRequest(http.MethodGet, "/api/chapters/"+uuid.Nil.String(), nil)
	rec := httptest.NewRecorder()

	app.newRouter(nil).ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Errorf("GET /api/chapters/{id} = 404, want the route to be mounted")
	}
}

func TestRouterMountsTheOIDCCallbackAdvertisedAsRedirectURI(t *testing.T) {
	t.Parallel()

	app, registry := newTestApp(t, &fakeDB{}, gatePort)
	registry.Set(componentMigrations, nil)

	req := httptest.NewRequest(http.MethodGet, APIPrefix+"/auth/oidc/callback", nil)
	rec := httptest.NewRecorder()

	app.newRouter(nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("GET %s/auth/oidc/callback = %d, want %d", APIPrefix, rec.Code, http.StatusFound)
	}

	if loc := rec.Header().Get("Location"); loc != "/login?error=oidcUnavailable" {
		t.Errorf("Location = %q, want the callback handler's redirect", loc)
	}
}

func TestRunComponentMarksOKBeforeEnteringTheLoop(t *testing.T) {
	reg := health.NewRegistry(componentAsuraScans)
	a := &App{deps: Deps{Health: reg}}

	var seen health.Status

	run := func(ctx context.Context) error {
		seen = reg.Snapshot(ctx).Components[componentAsuraScans].Status

		return nil
	}

	if err := a.runComponent(context.Background(), componentAsuraScans, run)(); err != nil {
		t.Fatalf("runComponent: %v", err)
	}

	if seen != health.StatusOK {
		t.Fatalf("status seen from loop = %q, want %q", seen, health.StatusOK)
	}
}

func TestRunComponentMarksFailedWhenLoopReturnsError(t *testing.T) {
	wantErr := errors.New("boucle morte")

	reg := health.NewRegistry(componentSessions)
	a := &App{deps: Deps{Health: reg}}

	run := func(context.Context) error { return wantErr }

	err := a.runComponent(context.Background(), componentSessions, run)()
	if err == nil {
		t.Fatal("runComponent: error expected")
	}

	if !errors.Is(err, wantErr) {
		t.Errorf("error %v, want it to wrap %v", err, wantErr)
	}

	c := reg.Snapshot(context.Background()).Components[componentSessions]
	if c.Status != health.StatusFailed {
		t.Errorf("status = %q, want %q", c.Status, health.StatusFailed)
	}

	if c.Reason != wantErr.Error() {
		t.Errorf("reason = %q, want %q", c.Reason, wantErr.Error())
	}
}

func TestRunComponentStaysOKOnCleanReturn(t *testing.T) {
	reg := health.NewRegistry(componentAsuraScans)
	a := &App{deps: Deps{Health: reg}}

	run := func(context.Context) error { return nil }

	if err := a.runComponent(context.Background(), componentAsuraScans, run)(); err != nil {
		t.Fatalf("runComponent: %v", err)
	}

	if got := reg.Snapshot(context.Background()).Components[componentAsuraScans].Status; got != health.StatusOK {
		t.Fatalf("status = %q, want %q", got, health.StatusOK)
	}
}

func TestRunServesReadyzWhileMigrationIsRunning(t *testing.T) {
	app, _ := startBlockedApp(t)
	app.waitListening(t)

	code, raw := app.get(t, "/readyz")
	if code != http.StatusServiceUnavailable {
		t.Fatalf("code = %d, want %d", code, http.StatusServiceUnavailable)
	}

	body := decodeReadyz(t, raw)
	if got := body.Components[componentMigrations].Status; got != string(health.StatusStarting) {
		t.Errorf("migrations = %q, want %q", got, health.StatusStarting)
	}
}

func TestRunReportsEverythingOKOnceMigrationCompletes(t *testing.T) {
	app, unblock := startBlockedApp(t)
	app.waitListening(t)
	unblock()

	var raw []byte

	app.waitFor(t, "/readyz never reached 200 after migration", func() bool {
		code, body := app.get(t, "/readyz")
		raw = body

		return code == http.StatusOK
	})

	body := decodeReadyz(t, raw)
	components := []string{
		componentMigrations, componentAsuraScans, componentCovers, componentDownloads, componentChapterListRefresh, componentSessions, componentOIDCRevalidation, componentDB,
	}
	for _, name := range components {
		if got := body.Components[name].Status; got != string(health.StatusOK) {
			t.Errorf("%s = %q, want %q", name, got, health.StatusOK)
		}
	}
}

func TestRunReturnsMigrationErrorAndNotContextCanceled(t *testing.T) {
	boom := errors.New("unknown column")
	db := &fakeDB{migrate: func() error { return boom }}

	err := startApp(t, db).wait(t)
	if !errors.Is(err, boom) {
		t.Fatalf("error = %v, want it to wrap %v", err, boom)
	}

	if errors.Is(err, context.Canceled) {
		t.Errorf("error = %v: main would treat it as clean shutdown and exit 0", err)
	}
}
