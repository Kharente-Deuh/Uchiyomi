// SPDX-License-Identifier: AGPL-3.0-or-later

package http_test

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readersettings"
	httpreadersettings "github.com/kharente-deuh/uchiyomi-server/pkg/core/readersettings/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

const (
	endpoint     = "/me"
	listPath     = "/me/reader-settings"
	cookieName   = "uchiyomi_session"
	testToken    = "letoken"
	testUsername = "alice"
)

type stubService struct {
	listErr     error
	replaceErr  error
	list        []readersettings.Profile
	lastReplace readersettings.ReplaceOpts
	lastList    uuid.UUID
}

func (s *stubService) ListForUser(_ context.Context, userID uuid.UUID) ([]readersettings.Profile, error) {
	s.lastList = userID

	return s.list, s.listErr
}

func (s *stubService) Replace(_ context.Context, opts readersettings.ReplaceOpts) (readersettings.Profile, error) {
	s.lastReplace = opts

	return readersettings.Profile{
		Type:        opts.Type,
		ReadingMode: opts.ReadingMode,
		PageScale:   opts.PageScale,
		DoublePage:  opts.DoublePage,
	}, s.replaceErr
}

type stubSessionService struct {
	result *sessions.AuthenticatedSession
}

func (s *stubSessionService) Authenticate(_ context.Context, _ string) (*sessions.AuthenticatedSession, error) {
	if s.result == nil {
		return nil, sessions.ErrInvalidSession
	}

	return s.result, nil
}

type listResponse struct {
	Items []struct {
		Type        string `json:"type"`
		ReadingMode string `json:"readingMode"`
		PageScale   string `json:"pageScale"`
		DoublePage  bool   `json:"doublePage"`
	} `json:"items"`
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(bytes.NewBuffer(nil), &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func frozenNow() time.Time {
	return time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
}

func authenticatorFor(t *testing.T, user *users.User) chi.Middlewares {
	t.Helper()

	logger := testLogger()

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

func newTestRouter(t *testing.T, svc *stubService, mws chi.Middlewares) chi.Router {
	t.Helper()

	logger := testLogger()

	c, err := httpreadersettings.New(
		httpreadersettings.Config{Endpoint: endpoint, Middlewares: mws},
		httpreadersettings.Deps{Logger: logger, Service: svc},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	r := chi.NewRouter()
	c.InitRouter(r)

	return r
}

func defaultProfiles() []readersettings.Profile {
	types := readersettings.AllTypes()
	out := make([]readersettings.Profile, 0, len(types))
	for _, typ := range types {
		out = append(out, readersettings.DefaultProfile(typ))
	}

	return out
}

func testUser() *users.User {
	return &users.User{ID: uuid.New(), Name: testUsername}
}

func getList(r chi.Router, withCookie bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, listPath, nil)
	if withCookie {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec
}

func putReplace(r chi.Router, comicType, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, listPath+"/"+comicType, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec
}

func TestNewValidatesConfigAndDeps(t *testing.T) {
	t.Parallel()

	logger := testLogger()
	svc := &stubService{}
	passthrough := func(next http.Handler) http.Handler { return next }

	tests := map[string]struct {
		deps    httpreadersettings.Deps
		wantErr string
		cfg     httpreadersettings.Config
	}{
		"empty endpoint": {
			cfg:     httpreadersettings.Config{},
			deps:    httpreadersettings.Deps{Logger: logger, Service: svc},
			wantErr: "cfg.Validate: endpoint is required",
		},
		"nil middleware": {
			cfg: httpreadersettings.Config{
				Endpoint:    endpoint,
				Middlewares: chi.Middlewares{passthrough, nil},
			},
			deps:    httpreadersettings.Deps{Logger: logger, Service: svc},
			wantErr: "cfg.Validate: all middlewares must not be nil",
		},
		"nil logger": {
			cfg:     httpreadersettings.Config{Endpoint: endpoint},
			deps:    httpreadersettings.Deps{Service: svc},
			wantErr: "deps.Validate: logger is required",
		},
		"nil service": {
			cfg:     httpreadersettings.Config{Endpoint: endpoint},
			deps:    httpreadersettings.Deps{Logger: logger},
			wantErr: "deps.Validate: service is required",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c, err := httpreadersettings.New(tc.cfg, tc.deps)
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

func TestListUnauthorized(t *testing.T) {
	t.Parallel()

	user := testUser()
	r := newTestRouter(t, &stubService{list: defaultProfiles()}, authenticatorFor(t, user))

	rec := getList(r, false)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestListOK(t *testing.T) {
	t.Parallel()

	user := testUser()
	svc := &stubService{list: defaultProfiles()}
	r := newTestRouter(t, svc, authenticatorFor(t, user))

	rec := getList(r, true)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got listResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body not decodable (%q): %v", rec.Body.String(), err)
	}

	if len(got.Items) != 4 {
		t.Fatalf("items len = %d, want 4 (body: %s)", len(got.Items), rec.Body.String())
	}

	first := got.Items[0]
	if first.Type != "manga" {
		t.Errorf("items[0].type = %q, want %q", first.Type, "manga")
	}

	if first.ReadingMode != "paged-rtl" {
		t.Errorf("items[0].readingMode = %q, want %q", first.ReadingMode, "paged-rtl")
	}

	if first.PageScale != "fit-screen" {
		t.Errorf("items[0].pageScale = %q, want %q", first.PageScale, "fit-screen")
	}

	if first.DoublePage {
		t.Errorf("items[0].doublePage = true, want false")
	}

	if svc.lastList != user.ID {
		t.Errorf("lastList = %s, want %s", svc.lastList, user.ID)
	}
}

func TestReplaceOK(t *testing.T) {
	t.Parallel()

	user := testUser()
	svc := &stubService{}
	r := newTestRouter(t, svc, authenticatorFor(t, user))

	body := `{"readingMode":"paged-rtl","pageScale":"fit-height","doublePage":true}`
	rec := putReplace(r, "manga", body)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got struct {
		Type       string `json:"type"`
		DoublePage bool   `json:"doublePage"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body not decodable (%q): %v", rec.Body.String(), err)
	}

	if got.Type != "manga" {
		t.Errorf("type = %q, want %q", got.Type, "manga")
	}

	if !got.DoublePage {
		t.Errorf("doublePage = false, want true")
	}

	if svc.lastReplace.UserID != user.ID {
		t.Errorf("lastReplace.UserID = %s, want %s", svc.lastReplace.UserID, user.ID)
	}

	if svc.lastReplace.Type != sources.SeriesTypeManga {
		t.Errorf("lastReplace.Type = %q, want %q", svc.lastReplace.Type, sources.SeriesTypeManga)
	}
}

func TestReplaceUnknownType(t *testing.T) {
	t.Parallel()

	user := testUser()
	svc := &stubService{}
	r := newTestRouter(t, svc, authenticatorFor(t, user))

	rec := putReplace(r, "webcomic", `{"readingMode":"paged-rtl","pageScale":"fit-height","doublePage":true}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	if svc.lastReplace.UserID != uuid.Nil {
		t.Errorf("Replace was called with UserID %s, want uuid.Nil", svc.lastReplace.UserID)
	}
}

func TestReplaceInvalidJSON(t *testing.T) {
	t.Parallel()

	user := testUser()
	r := newTestRouter(t, &stubService{}, authenticatorFor(t, user))

	rec := putReplace(r, "manga", `{`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestReplaceMissingField(t *testing.T) {
	t.Parallel()

	user := testUser()
	r := newTestRouter(t, &stubService{}, authenticatorFor(t, user))

	rec := putReplace(r, "manga", `{"readingMode":"paged-rtl","pageScale":"fit-height"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestReplaceWebtoonDoublePage(t *testing.T) {
	t.Parallel()

	user := testUser()
	svc := &stubService{replaceErr: readersettings.ErrInvalid}
	r := newTestRouter(t, svc, authenticatorFor(t, user))

	body := `{"readingMode":"webtoon","pageScale":"fit-width","doublePage":true}`
	rec := putReplace(r, "manhwa", body)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestReplaceUsesSessionUserNotBody(t *testing.T) {
	t.Parallel()

	user := testUser()
	svc := &stubService{}
	r := newTestRouter(t, svc, authenticatorFor(t, user))

	body := `{"readingMode":"paged-rtl","pageScale":"fit-height","doublePage":false}`
	rec := putReplace(r, "manga", body)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	if svc.lastReplace.UserID != user.ID {
		t.Errorf("lastReplace.UserID = %s, want session user %s", svc.lastReplace.UserID, user.ID)
	}
}
