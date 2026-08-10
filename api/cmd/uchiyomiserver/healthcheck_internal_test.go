// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	healthhttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/health/gateway/http"
)

func serverReturning(t *testing.T, status int) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)

	return srv
}

func TestHealthcheckSucceedsOn200(t *testing.T) {
	srv := serverReturning(t, http.StatusOK)

	if err := healthcheck(context.Background(), srv.URL+"/readyz"); err != nil {
		t.Fatalf("healthcheck: %v", err)
	}
}

func TestHealthcheckFailsOn503(t *testing.T) {
	srv := serverReturning(t, http.StatusServiceUnavailable)

	if err := healthcheck(context.Background(), srv.URL+"/readyz"); err == nil {
		t.Fatal("healthcheck: error expected on 503")
	}
}

func TestHealthcheckFailsWhenNothingListens(t *testing.T) {
	srv := serverReturning(t, http.StatusOK)
	url := srv.URL + "/readyz"
	srv.Close()

	if err := healthcheck(context.Background(), url); err == nil {
		t.Fatal("healthcheck: error expected when nobody is listening")
	}
}

func TestHealthcheckTimeoutExceedsTheProbeBudget(t *testing.T) {
	if healthcheckTimeout <= healthhttp.DefaultProbeTimeout {
		t.Fatalf(
			"healthcheckTimeout = %s, want strictly greater than DefaultProbeTimeout (%s)",
			healthcheckTimeout, healthhttp.DefaultProbeTimeout,
		)
	}
}

func TestReadyzURLUsesPortEnv(t *testing.T) {
	t.Setenv("PORT", "8080")

	if got := readyzURL(); got != "http://localhost:8080/readyz" {
		t.Fatalf("readyzURL = %q", got)
	}
}

func TestReadyzURLFallsBackToDefaultPort(t *testing.T) {
	t.Setenv("PORT", "")

	if got := readyzURL(); got != "http://localhost:3000/readyz" {
		t.Fatalf("readyzURL = %q", got)
	}
}
