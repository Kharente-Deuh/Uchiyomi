// SPDX-License-Identifier: AGPL-3.0-or-later

package http_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	sessionshttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
)

type stubSessionService struct {
	err      error
	result   *sessions.AuthenticatedSession
	gotToken string
	calls    int
}

func (s *stubSessionService) Authenticate(
	_ context.Context,
	token string,
) (*sessions.AuthenticatedSession, error) {
	s.calls++
	s.gotToken = token

	if s.err != nil {
		return nil, s.err
	}

	return s.result, nil
}

func testLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer

	return slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})), &buf
}

func protected(t *testing.T) http.Handler {
	t.Helper()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := sessionshttp.UserFrom(r.Context())
		if !ok {
			t.Error("le handler protégé n'a pas trouvé d'utilisateur dans le contexte")
			w.WriteHeader(http.StatusTeapot)

			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(user.Name))
	})
}

func newAuth(t *testing.T, svc *stubSessionService, now time.Time) (http.Handler, *bytes.Buffer) {
	t.Helper()

	logger, logs := testLogger()

	a, err := sessionshttp.NewAuthenticator(sessionshttp.AuthenticatorDeps{
		SessionService: svc,
		Cookies:        newCookies(t, true),
		Logger:         logger,
		Now:            func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("NewAuthenticator: %v", err)
	}

	return a.Middleware(protected(t)), logs
}

func requestWithToken(token string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.AddCookie(&http.Cookie{Name: cookieName, Value: token})

	return r
}

func TestMiddlewareRejectsMissingCookieWithoutCallingService(t *testing.T) {
	t.Parallel()

	svc := &stubSessionService{}
	h, _ := newAuth(t, svc, time.Now())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/protected", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	if svc.calls != 0 {
		t.Errorf("service appelé %d fois sans cookie", svc.calls)
	}
}

func TestMiddlewareClearsCookieOnInvalidSession(t *testing.T) {
	t.Parallel()

	svc := &stubSessionService{err: sessions.ErrInvalidSession}
	h, _ := newAuth(t, svc, time.Now())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithToken("perime"))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("%d cookies posés, want 1 pour effacer le token", len(cookies))
	}

	if cookies[0].MaxAge != -1 {
		t.Errorf("MaxAge = %d, want -1", cookies[0].MaxAge)
	}
}

func TestMiddlewareReturns500OnInfrastructureError(t *testing.T) {
	t.Parallel()

	svc := &stubSessionService{err: errors.New("connection refused")}
	h, logs := newAuth(t, svc, time.Now())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithToken(testToken))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	if strings.Contains(rec.Body.String(), "connection refused") {
		t.Errorf("l'erreur interne a fuité dans la réponse: %s", rec.Body.String())
	}

	if !strings.Contains(logs.String(), "connection refused") {
		t.Errorf("l'erreur interne est absente des logs: %s", logs.String())
	}

	if !strings.Contains(logs.String(), `"level":"ERROR"`) {
		t.Errorf("la panne d'infrastructure doit être loguée en ERROR: %s", logs.String())
	}
}

func TestMiddlewarePassesUserToHandler(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
	user := &users.User{ID: uuid.New(), Name: testUser}
	svc := &stubSessionService{result: &sessions.AuthenticatedSession{
		User:    user,
		Session: sessions.Session{ID: uuid.New(), UserID: user.ID, ExpiresAt: now.Add(time.Hour)},
	}}
	h, logs := newAuth(t, svc, now)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithToken(testToken))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Body.String() != testUser {
		t.Errorf("body = %q, want %q", rec.Body.String(), testUser)
	}

	if svc.gotToken != testToken {
		t.Errorf("le service a reçu %q, want %q", svc.gotToken, testToken)
	}

	if logs.Len() != 0 {
		t.Errorf("aucun log attendu sur le chemin nominal: %s", logs.String())
	}
}

func TestMiddlewareReissuesCookieAfterRenewal(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
	user := &users.User{ID: uuid.New(), Name: testUser}
	svc := &stubSessionService{result: &sessions.AuthenticatedSession{
		User:    user,
		Session: sessions.Session{ID: uuid.New(), UserID: user.ID, ExpiresAt: now.Add(3 * time.Hour)},
	}}
	h, _ := newAuth(t, svc, now)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithToken(testToken))

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("%d cookies posés, want 1", len(cookies))
	}

	if cookies[0].MaxAge != 10800 {
		t.Errorf("MaxAge = %d, want 10800", cookies[0].MaxAge)
	}

	if cookies[0].Value != testToken {
		t.Errorf("Value = %q, want le token inchangé", cookies[0].Value)
	}
}

func TestMiddlewareLogsRenewalFailureButServesRequest(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
	user := &users.User{ID: uuid.New(), Name: testUser}
	svc := &stubSessionService{result: &sessions.AuthenticatedSession{
		User:     user,
		Session:  sessions.Session{ID: uuid.New(), UserID: user.ID, ExpiresAt: now.Add(time.Hour)},
		RenewErr: errors.New("deadlock detected"),
	}}
	h, logs := newAuth(t, svc, now)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithToken(testToken))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d — une prolongation ratée ne déconnecte personne", rec.Code, http.StatusOK)
	}

	if !strings.Contains(logs.String(), "deadlock detected") {
		t.Errorf("l'échec de prolongation est absent des logs: %s", logs.String())
	}

	if !strings.Contains(logs.String(), `"level":"WARN"`) {
		t.Errorf("la prolongation ratée doit être loguée en WARN, pas au-dessus: %s", logs.String())
	}
}

func TestUserFromReturnsFalseWithoutMiddleware(t *testing.T) {
	t.Parallel()

	if _, ok := sessionshttp.UserFrom(context.Background()); ok {
		t.Error("UserFrom a trouvé un utilisateur dans un contexte vierge")
	}
}

func TestSessionFromReturnsFalseWithoutMiddleware(t *testing.T) {
	t.Parallel()

	if _, ok := sessionshttp.SessionFrom(context.Background()); ok {
		t.Error("SessionFrom a trouvé une session dans un contexte vierge")
	}
}

func TestRequireSessionRejectsMissingCookieAndClearsIt(t *testing.T) {
	t.Parallel()

	svc := &stubSessionService{}
	logger, _ := testLogger()

	a, err := sessionshttp.NewAuthenticator(sessionshttp.AuthenticatorDeps{
		SessionService: svc,
		Cookies:        newCookies(t, true),
		Logger:         logger,
		Now:            func() time.Time { return time.Now() },
	})
	if err != nil {
		t.Fatalf("NewAuthenticator: %v", err)
	}

	handler := a.RequireSession(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/logout", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	if svc.calls != 0 {
		t.Errorf("service appelé %d fois sans cookie", svc.calls)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("%d cookies posés, want 1", len(cookies))
	}

	if cookies[0].MaxAge != -1 {
		t.Errorf("MaxAge = %d, want -1", cookies[0].MaxAge)
	}
}

func TestRequireSessionPopulatesSessionFrom(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
	user := &users.User{ID: uuid.New(), Name: testUser}
	wantSession := sessions.Session{ID: uuid.New(), UserID: user.ID, ExpiresAt: now.Add(time.Hour)}
	svc := &stubSessionService{result: &sessions.AuthenticatedSession{
		User:    user,
		Session: wantSession,
	}}
	logger, _ := testLogger()

	a, err := sessionshttp.NewAuthenticator(sessionshttp.AuthenticatorDeps{
		SessionService: svc,
		Cookies:        newCookies(t, true),
		Logger:         logger,
		Now:            func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("NewAuthenticator: %v", err)
	}

	var gotSession sessions.Session

	handler := a.RequireSession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionshttp.SessionFrom(r.Context())
		if !ok {
			t.Error("SessionFrom n'a pas trouvé de session dans le contexte")
			w.WriteHeader(http.StatusTeapot)

			return
		}

		gotSession = session
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, requestWithToken(testToken))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if gotSession.ID != wantSession.ID {
		t.Errorf("session = %v, want %v", gotSession.ID, wantSession.ID)
	}
}

func newAuthenticatorForUser(t *testing.T, user *users.User) *sessionshttp.Authenticator {
	t.Helper()

	logger, _ := testLogger()

	a, err := sessionshttp.NewAuthenticator(sessionshttp.AuthenticatorDeps{
		SessionService: &stubSessionService{result: &sessions.AuthenticatedSession{
			User: user,
			Session: sessions.Session{
				ID:        uuid.New(),
				UserID:    user.ID,
				ExpiresAt: time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC).Add(time.Hour),
			},
		}},
		Cookies: newCookies(t, true),
		Logger:  logger,
	})
	if err != nil {
		t.Fatalf("NewAuthenticator: %v", err)
	}

	return a
}

func serveThroughAdminChain(t *testing.T, a *sessionshttp.Authenticator, withCookie bool) *httptest.ResponseRecorder {
	t.Helper()

	handler := a.Middleware(a.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if withCookie {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

func TestRequireAdminAllowsAnAdmin(t *testing.T) {
	t.Parallel()

	a := newAuthenticatorForUser(t, &users.User{ID: uuid.New(), Name: "root", IsAdmin: true})

	rec := serveThroughAdminChain(t, a, true)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestRequireAdminRejectsANonAdmin(t *testing.T) {
	t.Parallel()

	a := newAuthenticatorForUser(t, &users.User{ID: uuid.New(), Name: "alice", IsAdmin: false})

	rec := serveThroughAdminChain(t, a, true)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d (body: %s)", rec.Code, http.StatusForbidden, rec.Body.String())
	}

	if strings.Contains(rec.Body.String(), "alice") {
		t.Errorf("the identity leaked in the 403 body: %s", rec.Body.String())
	}
}

func TestRequireAdminRejectsAnEmptyContext(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	a := newAuthenticatorForUser(t, &users.User{ID: uuid.New(), Name: "alice", IsAdmin: true})
	handler := a.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
