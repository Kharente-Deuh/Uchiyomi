// SPDX-License-Identifier: AGPL-3.0-or-later

package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	sessionshttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/setup"
	setuphttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/setup/gateway/http"
)

const (
	setupEndpoint = "/setup"
	doSetupPath   = "/setup/"
	adminUsername = "alice"
	adminPassword = "hunter2hunter2"
)

type stubSetupService struct {
	err       error
	doErr     error
	doSession *sessions.IssuedSession
	gotOpts   setup.DoSetupOpts
	calls     int
	doCalls   int
	required  bool
}

func (s *stubSetupService) IsSetupRequired(context.Context) (bool, error) {
	s.calls++

	return s.required, s.err
}

func (s *stubSetupService) DoSetup(_ context.Context, opts setup.DoSetupOpts) (*sessions.IssuedSession, error) {
	s.doCalls++
	s.gotOpts = opts

	return s.doSession, s.doErr
}

func testLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer

	return slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})), &buf
}

func newCookies(t *testing.T) *sessionshttp.CookieManager {
	t.Helper()

	m, err := sessionshttp.NewCookieManager(sessionshttp.CookieConfig{Name: "uchiyomi_session", Path: "/"})
	if err != nil {
		t.Fatalf("NewCookieManager: %v", err)
	}

	return m
}

func frozenNow() time.Time {
	return time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
}

func setupRequestBody(username, password string) string {
	return fmt.Sprintf(`{"username":%q,"password":%q}`, username, password)
}

func newTestRouter(t *testing.T, svc *stubSetupService) (chi.Router, *bytes.Buffer) {
	t.Helper()

	logger, logs := testLogger()

	c, err := setuphttp.New(setuphttp.Config{Endpoint: setupEndpoint}, setuphttp.Deps{
		SetupService: svc,
		Cookies:      newCookies(t),
		Logger:       logger,
		Now:          frozenNow,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	r := chi.NewRouter()
	c.InitRouter(r)

	return r, logs
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		endpoint string
		wantErr  string
	}{
		"endpoint valide":       {endpoint: setupEndpoint, wantErr: ""},
		"nested endpoint":     {endpoint: "/api/v1/setup", wantErr: ""},
		"racine":                {endpoint: "/", wantErr: ""},
		"empty":                  {endpoint: "", wantErr: "endpoint is required"},
		"without leading slash":    {endpoint: "setup", wantErr: `endpoint must start with '/', got "setup"`},
		"URL absolute":          {endpoint: "http://x/setup", wantErr: `endpoint must start with '/', got "http://x/setup"`},
		"espace avant le slash": {endpoint: " /setup", wantErr: `endpoint must start with '/', got " /setup"`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := setuphttp.Config{Endpoint: tc.endpoint}
			err := cfg.Validate()

			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}

				return
			}

			if err == nil {
				t.Fatalf("Validate() = nil, want %q", tc.wantErr)
			}

			if err.Error() != tc.wantErr {
				t.Errorf("Validate() = %q, want %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestDepsValidate(t *testing.T) {
	t.Parallel()

	logger, _ := testLogger()

	tests := map[string]struct {
		deps    setuphttp.Deps
		wantErr string
	}{
		"complet": {
			deps:    setuphttp.Deps{SetupService: &stubSetupService{}, Cookies: newCookies(t), Logger: logger},
			wantErr: "",
		},
		"without service": {
			deps:    setuphttp.Deps{Cookies: newCookies(t), Logger: logger},
			wantErr: "setupService is required",
		},
		"without cookies": {
			deps:    setuphttp.Deps{SetupService: &stubSetupService{}, Logger: logger},
			wantErr: "cookies is required",
		},
		"without logger": {
			deps:    setuphttp.Deps{SetupService: &stubSetupService{}, Cookies: newCookies(t)},
			wantErr: "logger is required",
		},
		"tout manquant": {deps: setuphttp.Deps{}, wantErr: "setupService is required"},
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
		cfg     setuphttp.Config
		deps    setuphttp.Deps
		wantErr string
	}{
		"invalid config": {
			cfg:     setuphttp.Config{Endpoint: "setup"},
			deps:    setuphttp.Deps{SetupService: &stubSetupService{}, Cookies: newCookies(t), Logger: logger},
			wantErr: `cfg.Validate: endpoint must start with '/', got "setup"`,
		},
		"invalid deps": {
			cfg:     setuphttp.Config{Endpoint: setupEndpoint},
			deps:    setuphttp.Deps{Logger: logger},
			wantErr: "deps.Validate: setupService is required",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c, err := setuphttp.New(tc.cfg, tc.deps)
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

func TestGetSetupStatusOK(t *testing.T) {
	t.Parallel()

	tests := map[string]bool{
		"setup required":     true,
		"setup not required": false,
	}

	for name, required := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := &stubSetupService{required: required}
			r, logs := newTestRouter(t, svc)

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/setup/status", nil))

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
			}

			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want %q", ct, "application/json")
			}

			var got setuphttp.GetStatusResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("body not decodable (%q): %v", rec.Body.String(), err)
			}

			if got.Required != required {
				t.Errorf("required = %v, want %v", got.Required, required)
			}

			if svc.calls != 1 {
				t.Errorf("service called %d times, want 1", svc.calls)
			}

			if logs.Len() != 0 {
				t.Errorf("no logs expected on happy path: %s", logs.String())
			}
		})
	}
}

func TestGetSetupStatusServiceError(t *testing.T) {
	t.Parallel()

	svc := &stubSetupService{err: errors.New("database is down")}
	r, logs := newTestRouter(t, svc)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/setup/status", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var got struct {
		Message string `json:"message"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body not decodable (%q): %v", rec.Body.String(), err)
	}

	if got.Message != "Internal Server Error" {
		t.Errorf("message = %q, want %q", got.Message, "Internal Server Error")
	}

	if strings.Contains(rec.Body.String(), "database is down") {
		t.Errorf("internal error leaked into response: %s", rec.Body.String())
	}

	if !strings.Contains(logs.String(), "database is down") {
		t.Errorf("internal error missing from logs: %s", logs.String())
	}

	if !strings.Contains(logs.String(), "failed to check setup status") {
		t.Errorf("expected log message missing: %s", logs.String())
	}
}

func TestLoggerCarriesComponentAttribute(t *testing.T) {
	t.Parallel()

	svc := &stubSetupService{err: errors.New("boom")}
	r, logs := newTestRouter(t, svc)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/setup/status", nil))

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(logs.Bytes()), &entry); err != nil {
		t.Fatalf("log not decodable (%q): %v", logs.String(), err)
	}

	if entry["component"] != "setup.gateway.http" {
		//nolint:lll
		t.Errorf("component = %v, want %q — logger was not decorated in New", entry["component"], "setup.gateway.http")
	}
}

func TestInitRouterMountsUnderEndpoint(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		endpoint   string
		path       string
		wantStatus int
	}{
		"mounted route":           {endpoint: setupEndpoint, path: "/setup/status", wantStatus: http.StatusOK},
		"nested endpoint":      {endpoint: "/api/v1/setup", path: "/api/v1/setup/status", wantStatus: http.StatusOK},
		"unknown path":           {endpoint: setupEndpoint, path: "/status", wantStatus: http.StatusNotFound},
		"missing subpath":        {endpoint: setupEndpoint, path: "/setup/status/extra", wantStatus: http.StatusNotFound},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			logger, _ := testLogger()
			c, err := setuphttp.New(setuphttp.Config{Endpoint: tc.endpoint}, setuphttp.Deps{
				SetupService: &stubSetupService{},
				Cookies:      newCookies(t),
				Logger:       logger,
			})
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			r := chi.NewRouter()
			c.InitRouter(r)

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if rec.Code != tc.wantStatus {
				t.Errorf("GET %s = %d, want %d", tc.path, rec.Code, tc.wantStatus)
			}
		})
	}
}

func TestGetRootOfEndpointNowMethodNotAllowed(t *testing.T) {
	t.Parallel()

	r, _ := newTestRouter(t, &stubSetupService{})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, setupEndpoint, nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET %s = %d, want %d", setupEndpoint, rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestStatusRejectsNonGET(t *testing.T) {
	t.Parallel()

	r, _ := newTestRouter(t, &stubSetupService{})

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		t.Run(method, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(method, "/setup/status", nil))

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s /setup/status = %d, want %d", method, rec.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestDoSetupSetsSessionCookie(t *testing.T) {
	t.Parallel()

	svc := &stubSetupService{doSession: &sessions.IssuedSession{
		Session: sessions.Session{ID: uuid.New(), UserID: uuid.New(), ExpiresAt: frozenNow().Add(time.Hour)},
		Token:   "letoken",
	}}
	r, _ := newTestRouter(t, svc)

	rec := httptest.NewRecorder()
	body := strings.NewReader(setupRequestBody(adminUsername, adminPassword))
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, doSetupPath, body))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("%d cookies set, want 1", len(cookies))
	}

	if cookies[0].Value != "letoken" {
		t.Errorf("Value = %q, want %q", cookies[0].Value, "letoken")
	}

	if !cookies[0].HttpOnly {
		t.Error("HttpOnly missing")
	}
}

func TestDoSetupSucceedsWithoutCookieWhenSessionFails(t *testing.T) {
	t.Parallel()

	svc := &stubSetupService{doErr: fmt.Errorf("%w: disk full", setup.ErrSessionNotIssued)}
	r, logs := newTestRouter(t, svc)

	rec := httptest.NewRecorder()
	body := strings.NewReader(setupRequestBody(adminUsername, adminPassword))
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, doSetupPath, body))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	if n := len(rec.Result().Cookies()); n != 0 {
		t.Errorf("%d cookies set, want 0", n)
	}

	if !strings.Contains(logs.String(), "disk full") {
		t.Errorf("issuance failure missing from logs: %s", logs.String())
	}
}

func TestDoSetupConflictWhenAlreadyDone(t *testing.T) {
	t.Parallel()

	svc := &stubSetupService{doErr: setup.ErrSetupNotNeeded}
	r, _ := newTestRouter(t, svc)

	rec := httptest.NewRecorder()
	body := strings.NewReader(setupRequestBody(adminUsername, adminPassword))
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, doSetupPath, body))

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestDoSetupRejectsBadRequests(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"corps illisible":         `{`,
		"empty name":                setupRequestBody("", adminPassword),
		"password too short": setupRequestBody(adminUsername, "court"),
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := &stubSetupService{}
			r, _ := newTestRouter(t, svc)

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, doSetupPath, strings.NewReader(body)))

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}

			if svc.doCalls != 0 {
				t.Errorf("service called %d times on invalid request", svc.doCalls)
			}
		})
	}
}

func TestDoSetupServiceError(t *testing.T) {
	t.Parallel()

	svc := &stubSetupService{doErr: errors.New("database is down")}
	r, logs := newTestRouter(t, svc)

	rec := httptest.NewRecorder()
	body := strings.NewReader(setupRequestBody(adminUsername, adminPassword))
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, doSetupPath, body))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	if strings.Contains(rec.Body.String(), "database is down") {
		t.Errorf("internal error leaked into response: %s", rec.Body.String())
	}

	if !strings.Contains(logs.String(), "database is down") {
		t.Errorf("internal error missing from logs: %s", logs.String())
	}
}
