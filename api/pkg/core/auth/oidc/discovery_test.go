// SPDX-License-Identifier: AGPL-3.0-or-later

package oidc_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidc"
)

func discoveryServer(t *testing.T, body func(issuer string) string) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)

	t.Cleanup(srv.Close)

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body(srv.URL))
	})

	return srv
}

func fullDocument(issuer string) string {
	return fmt.Sprintf(`{
		"issuer": %q,
		"authorization_endpoint": "%s/auth",
		"token_endpoint": "%s/token",
		"userinfo_endpoint": "%s/userinfo",
		"end_session_endpoint": "%s/logout",
		"jwks_uri": "%s/jwks"
	}`, issuer, issuer, issuer, issuer, issuer, issuer)
}

func newDiscoverer(t *testing.T) *oidc.Discoverer {
	t.Helper()

	d, err := oidc.New(oidc.Config{Timeout: 5 * time.Second}, oidc.Deps{HTTPClient: http.DefaultClient})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return d
}

func TestDiscoverReadsTheEndpoints(t *testing.T) {
	t.Parallel()

	srv := discoveryServer(t, fullDocument)

	got, err := newDiscoverer(t).Discover(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	if got.Issuer != srv.URL {
		t.Errorf("Issuer = %q, want %q", got.Issuer, srv.URL)
	}

	if got.AuthorizationEndpoint != srv.URL+"/auth" {
		t.Errorf("AuthorizationEndpoint = %q", got.AuthorizationEndpoint)
	}

	if got.TokenEndpoint != srv.URL+"/token" {
		t.Errorf("TokenEndpoint = %q", got.TokenEndpoint)
	}

	if got.UserInfoEndpoint != srv.URL+"/userinfo" {
		t.Errorf("UserInfoEndpoint = %q", got.UserInfoEndpoint)
	}

	if got.EndSessionEndpoint != srv.URL+"/logout" {
		t.Errorf("EndSessionEndpoint = %q", got.EndSessionEndpoint)
	}
}

func TestDiscoverReportsAMissingEndSessionEndpoint(t *testing.T) {
	t.Parallel()

	srv := discoveryServer(t, func(issuer string) string {
		return fmt.Sprintf(`{
			"issuer": %q,
			"authorization_endpoint": "%s/auth",
			"token_endpoint": "%s/token"
		}`, issuer, issuer, issuer)
	})

	got, err := newDiscoverer(t).Discover(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	if got.EndSessionEndpoint != "" {
		t.Errorf("EndSessionEndpoint = %q, want empty", got.EndSessionEndpoint)
	}
}

func TestDiscoverRejectsAnIncompleteDocument(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"no authorization endpoint": `{"issuer": %[1]q, "token_endpoint": "%[1]s/token"}`,
		"no token endpoint":         `{"issuer": %[1]q, "authorization_endpoint": "%[1]s/auth"}`,
	}

	for name, tpl := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			srv := discoveryServer(t, func(issuer string) string {
				return fmt.Sprintf(tpl, issuer)
			})

			_, err := newDiscoverer(t).Discover(context.Background(), srv.URL)
			if !errors.Is(err, oidc.ErrDiscoveryIncomplete) {
				t.Errorf("Discover = %v, want ErrDiscoveryIncomplete", err)
			}
		})
	}
}

func TestDiscoverRejectsAnUnreachableIssuer(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	_, err := newDiscoverer(t).Discover(context.Background(), srv.URL)
	if !errors.Is(err, oidc.ErrDiscoveryFailed) {
		t.Errorf("Discover = %v, want ErrDiscoveryFailed", err)
	}
}

func TestDiscoverRejectsAnIssuerMismatch(t *testing.T) {
	t.Parallel()

	srv := discoveryServer(t, func(issuer string) string {
		return fmt.Sprintf(`{
			"issuer": "https://elsewhere.example.com",
			"authorization_endpoint": "%s/auth",
			"token_endpoint": "%s/token"
		}`, issuer, issuer)
	})

	_, err := newDiscoverer(t).Discover(context.Background(), srv.URL)
	if !errors.Is(err, oidc.ErrDiscoveryFailed) {
		t.Errorf("Discover = %v, want ErrDiscoveryFailed", err)
	}
}

func TestNewValidatesDeps(t *testing.T) {
	t.Parallel()

	if _, err := oidc.New(oidc.Config{Timeout: time.Second}, oidc.Deps{}); err == nil {
		t.Error("New without an HTTP client = nil, want an error")
	}
}
