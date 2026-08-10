// SPDX-License-Identifier: AGPL-3.0-or-later

package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	sessionshttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	usershttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/users/gateway/http"
)

const (
	usersEndpoint = "/users"
	mePath        = "/users/me"
	cookieName    = "uchiyomi_session"
	testToken     = "letoken"
	testUsername  = "alice"
)

type stubSessionService struct {
	result *sessions.AuthenticatedSession
}

func (s *stubSessionService) Authenticate(_ context.Context, _ string) (*sessions.AuthenticatedSession, error) {
	if s.result == nil {
		return nil, sessions.ErrInvalidSession
	}

	return s.result, nil
}

func testLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer

	return slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})), &buf
}

func frozenNow() time.Time {
	return time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
}

func authenticatorFor(t *testing.T, user *users.User) chi.Middlewares {
	t.Helper()

	logger, _ := testLogger()

	cookies, err := sessionshttp.NewCookieManager(sessionshttp.CookieConfig{Name: cookieName, Path: "/"})
	if err != nil {
		t.Fatalf("NewCookieManager: %v", err)
	}

	a, err := sessionshttp.NewAuthenticator(sessionshttp.AuthenticatorDeps{
		SessionService: &stubSessionService{result: &sessions.AuthenticatedSession{
			User:    user,
			Session: sessions.Session{ID: uuid.New(), UserID: user.ID, ExpiresAt: frozenNow().Add(time.Hour)},
		}},
		Cookies: cookies,
		Logger:  logger,
		Now:     frozenNow,
	})
	if err != nil {
		t.Fatalf("NewAuthenticator: %v", err)
	}

	return chi.Middlewares{a.Middleware}
}

func newTestRouter(t *testing.T, mws chi.Middlewares) (chi.Router, *bytes.Buffer) {
	t.Helper()

	logger, logs := testLogger()

	c, err := usershttp.New(
		usershttp.Config{Endpoint: usersEndpoint, Middlewares: mws},
		usershttp.Deps{Logger: logger},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	r := chi.NewRouter()
	c.InitRouter(r)

	return r, logs
}

func getMe(r chi.Router, withCookie bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, mePath, nil)
	if withCookie {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	passthrough := func(next http.Handler) http.Handler { return next }

	tests := map[string]struct {
		wantErr string
		cfg     usershttp.Config
	}{
		"endpoint valide":   {cfg: usershttp.Config{Endpoint: usersEndpoint}},
		"nested endpoint": {cfg: usershttp.Config{Endpoint: "/api/v1/users"}},
		"racine":            {cfg: usershttp.Config{Endpoint: "/"}},
		"middlewares set": {cfg: usershttp.Config{Endpoint: usersEndpoint, Middlewares: chi.Middlewares{passthrough}}},
		"empty":              {cfg: usershttp.Config{}, wantErr: "endpoint is required"},
		"without leading slash": {
			cfg:     usershttp.Config{Endpoint: "users"},
			wantErr: `endpoint must start with '/', got "users"`,
		},
		"URL absolute": {
			cfg:     usershttp.Config{Endpoint: "http://x/users"},
			wantErr: `endpoint must start with '/', got "http://x/users"`,
		},
		"middleware nil": {
			cfg:     usershttp.Config{Endpoint: usersEndpoint, Middlewares: chi.Middlewares{passthrough, nil}},
			wantErr: "all middlewares must not be nil",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := tc.cfg
			err := cfg.Validate()

			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}

				return
			}

			if err == nil || err.Error() != tc.wantErr {
				t.Errorf("Validate() = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestDepsValidate(t *testing.T) {
	t.Parallel()

	logger, _ := testLogger()

	tests := map[string]struct {
		deps    usershttp.Deps
		wantErr string
	}{
		"complet":     {deps: usershttp.Deps{Logger: logger}},
		"without logger": {deps: usershttp.Deps{}, wantErr: "logger is required"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			deps := tc.deps
			err := deps.Validate()

			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}

				return
			}

			if err == nil || err.Error() != tc.wantErr {
				t.Errorf("Validate() = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestNewFailsFast(t *testing.T) {
	t.Parallel()

	logger, _ := testLogger()

	tests := map[string]struct {
		deps    usershttp.Deps
		wantErr string
		cfg     usershttp.Config
	}{
		"invalid config": {
			cfg:     usershttp.Config{Endpoint: "users"},
			deps:    usershttp.Deps{Logger: logger},
			wantErr: `cfg.Validate: endpoint must start with '/', got "users"`,
		},
		"invalid deps": {
			cfg:     usershttp.Config{Endpoint: usersEndpoint},
			deps:    usershttp.Deps{},
			wantErr: "deps.Validate: logger is required",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c, err := usershttp.New(tc.cfg, tc.deps)
			if err == nil {
				t.Fatalf("New() = nil, want %q", tc.wantErr)
			}

			if c != nil {
				t.Error("New returned a controller in addition to the error")
			}

			if err.Error() != tc.wantErr {
				t.Errorf("New() = %q, want %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestGetMeReturnsAuthenticatedUser(t *testing.T) {
	t.Parallel()

	tests := map[string]bool{
		"administrateur": true,
		"user":    false,
	}

	for name, isAdmin := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			user := &users.User{ID: uuid.New(), Name: testUsername, IsAdmin: isAdmin}
			r, logs := newTestRouter(t, authenticatorFor(t, user))

			rec := getMe(r, true)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
			}

			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want %q", ct, "application/json")
			}

			var got usershttp.GetMeResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("body not decodable (%q): %v", rec.Body.String(), err)
			}

			if got.ID != user.ID.String() {
				t.Errorf("id = %q, want %q", got.ID, user.ID.String())
			}

			if got.Username != testUsername {
				t.Errorf("username = %q, want %q", got.Username, testUsername)
			}

			if got.IsAdmin != isAdmin {
				t.Errorf("isAdmin = %v, want %v", got.IsAdmin, isAdmin)
			}

			if logs.Len() != 0 {
				t.Errorf("no logs expected on happy path: %s", logs.String())
			}
		})
	}
}

func TestGetMeExposesOnlyItsDTOFields(t *testing.T) {
	t.Parallel()

	user := &users.User{
		ID:        uuid.New(),
		Name:      testUsername,
		CreatedAt: frozenNow(),
		UpdatedAt: frozenNow(),
	}
	r, _ := newTestRouter(t, authenticatorFor(t, user))

	rec := getMe(r, true)

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body not decodable (%q): %v", rec.Body.String(), err)
	}

	want := map[string]bool{"id": true, "username": true, "isAdmin": true}
	for key := range got {
		if !want[key] {
			t.Errorf("field %q exposed by /me although not in GetMeResponse", key)
		}
	}

	if len(got) != len(want) {
		t.Errorf("%d fields in response, want %d: %s", len(got), len(want), rec.Body.String())
	}
}

func TestGetMeUnauthorizedWhenContextHasNoUser(t *testing.T) {
	t.Parallel()

	r, _ := newTestRouter(t, nil)

	rec := getMe(r, false)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}

	var got struct {
		Message string `json:"message"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body not decodable (%q): %v", rec.Body.String(), err)
	}

	if got.Message != "Unauthorized" {
		t.Errorf("message = %q, want %q", got.Message, "Unauthorized")
	}
}

func TestGetMeRejectsRequestWithoutSessionCookie(t *testing.T) {
	t.Parallel()

	r, _ := newTestRouter(t, authenticatorFor(t, &users.User{ID: uuid.New(), Name: testUsername}))

	rec := getMe(r, false)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	if rec.Body.Len() > 0 && bytes.Contains(rec.Body.Bytes(), []byte(testUsername)) {
		t.Errorf("identity leaked on unauthenticated request: %s", rec.Body.String())
	}
}

func TestInitRouterMountsUnderEndpoint(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		endpoint   string
		path       string
		wantStatus int
	}{
		"mounted route":           {endpoint: usersEndpoint, path: mePath, wantStatus: http.StatusOK},
		"nested endpoint":      {endpoint: "/api/v1/users", path: "/api/v1/users/me", wantStatus: http.StatusOK},
		"unknown path":           {endpoint: usersEndpoint, path: "/me", wantStatus: http.StatusNotFound},
		"missing subpath":        {endpoint: usersEndpoint, path: "/users/me/extra", wantStatus: http.StatusNotFound},
		"racine de l'endpoint":   {endpoint: usersEndpoint, path: usersEndpoint, wantStatus: http.StatusNotFound},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			user := &users.User{ID: uuid.New(), Name: testUsername}

			logger, _ := testLogger()
			c, err := usershttp.New(
				usershttp.Config{Endpoint: tc.endpoint, Middlewares: authenticatorFor(t, user)},
				usershttp.Deps{Logger: logger},
			)
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			r := chi.NewRouter()
			c.InitRouter(r)

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("GET %s = %d, want %d", tc.path, rec.Code, tc.wantStatus)
			}
		})
	}
}

func TestMeRejectsNonGET(t *testing.T) {
	t.Parallel()

	r, _ := newTestRouter(t, authenticatorFor(t, &users.User{ID: uuid.New(), Name: testUsername}))

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		t.Run(method, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(method, mePath, nil)
			req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s %s = %d, want %d", method, mePath, rec.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}
