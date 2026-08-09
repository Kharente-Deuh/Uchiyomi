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
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth"
	authhttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	sessionshttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
)

const (
	authEndpoint      = "/auth"
	loginPath         = "/auth/login"
	logoutPath        = "/auth/logout"
	username          = "alice"
	password          = "hunter2hunter2"
	token             = "letoken"
	sessionCookieName = "uchiyomi_session"
)

type stubAuthService struct {
	err         error
	logoutErr   error
	result      *auth.LoginResult
	gotOpts     auth.LoginWithPwdOpts
	logoutToken string
	calls       int
	logoutCalls int
}

func (s *stubAuthService) LoginWithPwd(
	_ context.Context,
	opts auth.LoginWithPwdOpts,
) (*auth.LoginResult, error) {
	s.calls++
	s.gotOpts = opts

	if s.err != nil {
		return nil, s.err
	}

	return s.result, nil
}

func (s *stubAuthService) CreateUserWithPwd(context.Context, auth.CreateUserWithPwdOpts) (*users.User, error) {
	panic("CreateUserWithPwd n'est pas exposée par ce controller")
}

func (s *stubAuthService) Logout(_ context.Context, token string) error {
	s.logoutCalls++
	s.logoutToken = token

	return s.logoutErr
}

func frozenNow() time.Time {
	return time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
}

func issuedSession() *sessions.IssuedSession {
	return &sessions.IssuedSession{
		Session: sessions.Session{ID: uuid.New(), UserID: uuid.New(), ExpiresAt: frozenNow().Add(time.Hour)},
		Token:   token,
	}
}

func loggedInUser() *users.User {
	return &users.User{ID: uuid.New(), Name: username, IsAdmin: true}
}

func loggedInStub() *stubAuthService {
	return &stubAuthService{result: &auth.LoginResult{Session: issuedSession(), User: loggedInUser()}}
}

func testLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer

	return slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})), &buf
}

func newCookies(t *testing.T) *sessionshttp.CookieManager {
	t.Helper()

	m, err := sessionshttp.NewCookieManager(sessionshttp.CookieConfig{Name: sessionCookieName, Path: "/"})
	if err != nil {
		t.Fatalf("NewCookieManager: %v", err)
	}

	return m
}

func loginBody(user, pwd string) string {
	return fmt.Sprintf(`{"username":%q,"password":%q}`, user, pwd)
}

func newTestRouter(t *testing.T, svc *stubAuthService) (chi.Router, *bytes.Buffer) {
	t.Helper()

	logger, logs := testLogger()

	c, err := authhttp.New(authhttp.Config{Endpoint: authEndpoint}, authhttp.Deps{
		AuthService: svc,
		Cookies:     newCookies(t),
		Logger:      logger,
		Now:         frozenNow,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	r := chi.NewRouter()
	c.InitRouter(r)

	return r, logs
}

func postLogin(t *testing.T, r chi.Router, body string) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, loginPath, strings.NewReader(body)))

	return rec
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		endpoint string
		wantErr  string
	}{
		"endpoint valide":    {endpoint: authEndpoint, wantErr: ""},
		"endpoint imbriqué":  {endpoint: "/api/v1/auth", wantErr: ""},
		"racine":             {endpoint: "/", wantErr: ""},
		"vide":               {endpoint: "", wantErr: "endpoint is required"},
		"sans slash initial": {endpoint: "auth", wantErr: `endpoint must start with '/', got "auth"`},
		"URL complète":       {endpoint: "http://x/auth", wantErr: `endpoint must start with '/', got "http://x/auth"`},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := authhttp.Config{Endpoint: tc.endpoint}
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
		deps    authhttp.Deps
		wantErr string
	}{
		"complet": {
			deps:    authhttp.Deps{AuthService: &stubAuthService{}, Cookies: newCookies(t), Logger: logger},
			wantErr: "",
		},
		"sans service": {
			deps:    authhttp.Deps{Cookies: newCookies(t), Logger: logger},
			wantErr: "authService is required",
		},
		"sans cookies": {
			deps:    authhttp.Deps{AuthService: &stubAuthService{}, Logger: logger},
			wantErr: "cookies is required",
		},
		"sans logger": {
			deps:    authhttp.Deps{AuthService: &stubAuthService{}, Cookies: newCookies(t)},
			wantErr: "logger is required",
		},
		"tout manquant": {deps: authhttp.Deps{}, wantErr: "cookies is required"},
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
		cfg     authhttp.Config
		deps    authhttp.Deps
		wantErr string
	}{
		"config invalide": {
			cfg:     authhttp.Config{Endpoint: "auth"},
			deps:    authhttp.Deps{AuthService: &stubAuthService{}, Cookies: newCookies(t), Logger: logger},
			wantErr: `cfg.Validate: endpoint must start with '/', got "auth"`,
		},
		"deps invalides": {
			cfg:     authhttp.Config{Endpoint: authEndpoint},
			deps:    authhttp.Deps{Logger: logger},
			wantErr: "deps.Validate: cookies is required",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c, err := authhttp.New(tc.cfg, tc.deps)
			if err == nil {
				t.Fatalf("New() = nil, want %q", tc.wantErr)
			}

			if c != nil {
				t.Error("New a renvoyé un controller en plus de l'erreur")
			}

			if err.Error() != tc.wantErr {
				t.Errorf("New() = %q, want %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestInitRouterMountsUnderEndpoint(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		endpoint   string
		path       string
		wantStatus int
	}{
		"route montée":      {endpoint: authEndpoint, path: loginPath, wantStatus: http.StatusOK},
		"endpoint imbriqué": {endpoint: "/api/v1/auth", path: "/api/v1/auth/login", wantStatus: http.StatusOK},
		"hors endpoint":     {endpoint: authEndpoint, path: "/login", wantStatus: http.StatusNotFound},
		"sous-chemin":       {endpoint: authEndpoint, path: "/auth/login/extra", wantStatus: http.StatusNotFound},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			logger, _ := testLogger()
			c, err := authhttp.New(authhttp.Config{Endpoint: tc.endpoint}, authhttp.Deps{
				AuthService: loggedInStub(),
				Cookies:     newCookies(t),
				Logger:      logger,
				Now:         frozenNow,
			})
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			r := chi.NewRouter()
			c.InitRouter(r)

			rec := httptest.NewRecorder()
			body := strings.NewReader(loginBody(username, password))
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, tc.path, body))

			if rec.Code != tc.wantStatus {
				t.Errorf("POST %s = %d, want %d (body: %s)", tc.path, rec.Code, tc.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestLoginRejectsNonPOST(t *testing.T) {
	t.Parallel()

	r, _ := newTestRouter(t, loggedInStub())

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		t.Run(method, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(method, loginPath, nil))

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s %s = %d, want %d", method, loginPath, rec.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestLoginSetsSessionCookie(t *testing.T) {
	t.Parallel()

	svc := loggedInStub()
	r, logs := newTestRouter(t, svc)

	rec := postLogin(t, r, loginBody(username, password))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("%d cookies posés, want 1", len(cookies))
	}

	if cookies[0].Value != token {
		t.Errorf("Value = %q, want %q", cookies[0].Value, token)
	}

	if !cookies[0].HttpOnly {
		t.Error("HttpOnly absent : le token deviendrait lisible en JS")
	}

	if want := int(time.Hour.Seconds()); cookies[0].MaxAge != want {
		t.Errorf("MaxAge = %d, want %d", cookies[0].MaxAge, want)
	}

	if logs.Len() != 0 {
		t.Errorf("aucun log attendu sur le chemin nominal: %s", logs.String())
	}
}

func TestLoginReturnsTheAuthenticatedUser(t *testing.T) {
	t.Parallel()

	svc := loggedInStub()
	r, _ := newTestRouter(t, svc)

	rec := postLogin(t, r, loginBody(username, password))

	var got authhttp.LoginWithPwdResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal(%s): %v", rec.Body.String(), err)
	}

	want := authhttp.LoginWithPwdResponse{
		ID:       svc.result.User.ID.String(),
		Username: username,
		IsAdmin:  true,
	}

	if got != want {
		t.Errorf("body = %+v, want %+v", got, want)
	}
}

func TestLoginPassesTrimmedCredentialsToService(t *testing.T) {
	t.Parallel()

	svc := loggedInStub()
	r, _ := newTestRouter(t, svc)

	rec := postLogin(t, r, loginBody("  "+username+" ", password))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	if svc.gotOpts.Username != username {
		t.Errorf("Username = %q, want %q", svc.gotOpts.Username, username)
	}

	if svc.gotOpts.Password != password {
		t.Errorf("Password = %q, want %q", svc.gotOpts.Password, password)
	}
}

func TestLoginRejectsBadCredentials(t *testing.T) {
	t.Parallel()

	svc := &stubAuthService{err: auth.ErrInvalidLoginPwd}
	r, logs := newTestRouter(t, svc)

	rec := postLogin(t, r, loginBody(username, "mauvais"))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	if n := len(rec.Result().Cookies()); n != 0 {
		t.Errorf("%d cookies posés sur un login refusé", n)
	}

	var got struct {
		Message string `json:"message"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("corps non décodable (%q): %v", rec.Body.String(), err)
	}

	if got.Message != "invalid login/password" {
		t.Errorf("message = %q, want %q", got.Message, "invalid login/password")
	}

	if logs.Len() != 0 {
		t.Errorf("un login refusé n'est pas une erreur serveur: %s", logs.String())
	}
}

func TestLoginRejectsBadRequests(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"corps illisible":       `{`,
		"nom vide":              loginBody("", password),
		"nom en espaces":        loginBody("   ", password),
		"mot de passe vide":     loginBody(username, ""),
		"champ inconnu":         `{"username":"alice","password":"x","admin":true}`,
		"mauvais type de champ": `{"username":1,"password":"x"}`,
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := loggedInStub()
			r, _ := newTestRouter(t, svc)

			rec := postLogin(t, r, body)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}

			if svc.calls != 0 {
				t.Errorf("service appelé %d fois sur une requête invalide", svc.calls)
			}
		})
	}
}

func TestLoginServiceError(t *testing.T) {
	t.Parallel()

	svc := &stubAuthService{err: errors.New("database is down")}
	r, logs := newTestRouter(t, svc)

	rec := postLogin(t, r, loginBody(username, password))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	if n := len(rec.Result().Cookies()); n != 0 {
		t.Errorf("%d cookies posés alors que le login a échoué", n)
	}

	if strings.Contains(rec.Body.String(), "database is down") {
		t.Errorf("l'erreur interne a fuité dans la réponse: %s", rec.Body.String())
	}

	if !strings.Contains(logs.String(), "database is down") {
		t.Errorf("l'erreur interne est absente des logs: %s", logs.String())
	}

	if !strings.Contains(logs.String(), "failed to login") {
		t.Errorf("message de log attendu absent: %s", logs.String())
	}
}

func TestLoggerCarriesComponentAttribute(t *testing.T) {
	t.Parallel()

	svc := &stubAuthService{err: errors.New("boom")}
	r, logs := newTestRouter(t, svc)

	postLogin(t, r, loginBody(username, password))

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(logs.Bytes()), &entry); err != nil {
		t.Fatalf("log non décodable (%q): %v", logs.String(), err)
	}

	if entry["component"] != "auth.gateway.http" {
		t.Errorf("component = %v, want %q", entry["component"], "auth.gateway.http")
	}
}

func postLogout(t *testing.T, r chi.Router, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, logoutPath, nil)

	if cookie != nil {
		req.AddCookie(cookie)
	}

	r.ServeHTTP(rec, req)

	return rec
}

func TestLogoutRevokesSessionAndClearsCookie(t *testing.T) {
	t.Parallel()

	svc := &stubAuthService{}
	r, _ := newTestRouter(t, svc)

	rec := postLogout(t, r, &http.Cookie{Name: sessionCookieName, Value: token})

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusNoContent, rec.Body.String())
	}

	if svc.logoutCalls != 1 {
		t.Errorf("Logout appelé %d fois, want 1", svc.logoutCalls)
	}

	if svc.logoutToken != token {
		t.Errorf("token reçu = %q, want %q", svc.logoutToken, token)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("%d cookies posés, want 1", len(cookies))
	}

	if cookies[0].MaxAge != -1 {
		t.Errorf("MaxAge = %d, want -1", cookies[0].MaxAge)
	}

	if cookies[0].Value != "" {
		t.Errorf("Value = %q, want vide", cookies[0].Value)
	}
}

func TestLogoutWithoutCookieClearsCookieAndSkipsService(t *testing.T) {
	t.Parallel()

	svc := &stubAuthService{}
	r, _ := newTestRouter(t, svc)

	rec := postLogout(t, r, nil)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusNoContent, rec.Body.String())
	}

	if svc.logoutCalls != 0 {
		t.Errorf("Logout appelé %d fois, want 0", svc.logoutCalls)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("%d cookies posés, want 1", len(cookies))
	}

	if cookies[0].MaxAge != -1 {
		t.Errorf("MaxAge = %d, want -1", cookies[0].MaxAge)
	}
}

func TestLogoutServiceErrorLeavesCookieIntact(t *testing.T) {
	t.Parallel()

	svc := &stubAuthService{logoutErr: errors.New("database is down")}
	r, logs := newTestRouter(t, svc)

	rec := postLogout(t, r, &http.Cookie{Name: sessionCookieName, Value: token})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	if n := len(rec.Result().Cookies()); n != 0 {
		t.Errorf("%d cookies posés alors que le logout a échoué", n)
	}

	if !strings.Contains(logs.String(), "failed to logout") {
		t.Errorf("message de log attendu absent: %s", logs.String())
	}
}
