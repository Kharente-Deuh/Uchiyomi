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
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readingprogress"
	httpreadingprogress "github.com/kharente-deuh/uchiyomi-server/pkg/core/readingprogress/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
)

const (
	endpoint         = "/comics"
	chaptersEndpoint = "/chapters"
	cookieName       = "uchiyomi_session"
	testToken        = "letoken"
	testUsername     = "alice"
)

type stubService struct {
	listErr    error
	saveErr    error
	markErr    error
	listResult readingprogress.ListResult
	markResult readingprogress.ListResult
	saveResult readingprogress.Progress
	lastMark   readingprogress.MarkReadOpts
	lastSave   readingprogress.SaveOpts
	lastList   readingprogress.ListOpts
}

func (s *stubService) List(_ context.Context, opts readingprogress.ListOpts) (readingprogress.ListResult, error) {
	s.lastList = opts

	return s.listResult, s.listErr
}

func (s *stubService) Save(_ context.Context, opts readingprogress.SaveOpts) (readingprogress.Progress, error) {
	s.lastSave = opts

	return s.saveResult, s.saveErr
}

func (s *stubService) MarkRead(
	_ context.Context, opts readingprogress.MarkReadOpts,
) (readingprogress.ListResult, error) {
	s.lastMark = opts

	return s.markResult, s.markErr
}

func (s *stubService) MapByChapterIDs(
	_ context.Context, _ readingprogress.MapOpts,
) (map[uuid.UUID]readingprogress.Progress, error) {
	return map[uuid.UUID]readingprogress.Progress{}, nil
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

type progressJSON struct {
	ChapterID string `json:"chapterId"`
	UpdatedAt string `json:"updatedAt"`
	Page      int    `json:"page"`
}

type continueJSON struct {
	Continue *struct {
		ChapterID string `json:"chapterId"`
		Page      int    `json:"page"`
	} `json:"continue"`
}

func testLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
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

	c, err := httpreadingprogress.New(
		httpreadingprogress.Config{
			Endpoint:         endpoint,
			ChaptersEndpoint: chaptersEndpoint,
			Middlewares:      mws,
		},
		httpreadingprogress.Deps{Logger: testLogger(), Service: svc},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	r := chi.NewRouter()
	c.InitRouter(r)

	return r
}

func testUser() *users.User {
	return &users.User{ID: uuid.New(), Name: testUsername}
}

func getProgress(t *testing.T, r chi.Router, comicID string, withCookie bool) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, endpoint+"/"+comicID+"/progress", nil)
	if withCookie {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec
}

func putProgress(t *testing.T, r chi.Router, chapterID, body string, withCookie bool) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPut, chaptersEndpoint+"/"+chapterID+"/progress", strings.NewReader(body))
	if withCookie {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	}

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
		deps    httpreadingprogress.Deps
		wantErr string
		cfg     httpreadingprogress.Config
	}{
		"empty endpoint": {
			cfg:     httpreadingprogress.Config{ChaptersEndpoint: chaptersEndpoint},
			deps:    httpreadingprogress.Deps{Logger: logger, Service: svc},
			wantErr: "cfg.Validate: endpoint is required",
		},
		"empty chapters endpoint": {
			cfg:     httpreadingprogress.Config{Endpoint: endpoint},
			deps:    httpreadingprogress.Deps{Logger: logger, Service: svc},
			wantErr: "cfg.Validate: chapters endpoint is required",
		},
		"nil middleware": {
			cfg: httpreadingprogress.Config{
				Endpoint:         endpoint,
				ChaptersEndpoint: chaptersEndpoint,
				Middlewares:      chi.Middlewares{passthrough, nil},
			},
			deps:    httpreadingprogress.Deps{Logger: logger, Service: svc},
			wantErr: "cfg.Validate: all middlewares must not be nil",
		},
		"nil logger": {
			cfg:     httpreadingprogress.Config{Endpoint: endpoint, ChaptersEndpoint: chaptersEndpoint},
			deps:    httpreadingprogress.Deps{Service: svc},
			wantErr: "deps.Validate: logger is required",
		},
		"nil service": {
			cfg:     httpreadingprogress.Config{Endpoint: endpoint, ChaptersEndpoint: chaptersEndpoint},
			deps:    httpreadingprogress.Deps{Logger: logger},
			wantErr: "deps.Validate: service is required",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c, err := httpreadingprogress.New(tc.cfg, tc.deps)
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

func TestGetUnauthorized(t *testing.T) {
	t.Parallel()

	user := testUser()
	r := newTestRouter(t, &stubService{}, authenticatorFor(t, user))

	rec := getProgress(t, r, uuid.New().String(), false)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestPutUnauthorized(t *testing.T) {
	t.Parallel()

	user := testUser()
	r := newTestRouter(t, &stubService{}, authenticatorFor(t, user))

	rec := putProgress(t, r, uuid.New().String(), `{"page":1}`, false)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestGetOK(t *testing.T) {
	t.Parallel()

	user := testUser()
	comicID := uuid.New()
	chapterNewer := uuid.New()

	svc := &stubService{
		listResult: readingprogress.ListResult{
			Continue: &readingprogress.Continue{ChapterID: chapterNewer, Page: 5},
		},
	}
	r := newTestRouter(t, svc, authenticatorFor(t, user))

	rec := getProgress(t, r, comicID.String(), true)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	if svc.lastList.UserID != user.ID {
		t.Errorf("lastList.UserID = %s, want %s", svc.lastList.UserID, user.ID)
	}

	if svc.lastList.ComicID != comicID {
		t.Errorf("lastList.ComicID = %s, want %s", svc.lastList.ComicID, comicID)
	}

	var got continueJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body not decodable (%q): %v", rec.Body.String(), err)
	}

	if got.Continue == nil {
		t.Fatal("continue is nil, want object")
	}

	if got.Continue.ChapterID != chapterNewer.String() || got.Continue.Page != 5 {
		t.Errorf("continue = %+v", got.Continue)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode raw body: %v", err)
	}

	if _, ok := raw["chapters"]; ok {
		t.Error("body must not include chapters")
	}

	var continueRaw map[string]json.RawMessage
	if err := json.Unmarshal(raw["continue"], &continueRaw); err != nil {
		t.Fatalf("decode continue: %v", err)
	}

	if _, ok := continueRaw["updatedAt"]; ok {
		t.Error("continue must not include updatedAt")
	}
}

func TestGetEmptyContinueNull(t *testing.T) {
	t.Parallel()

	user := testUser()
	comicID := uuid.New()
	svc := &stubService{
		listResult: readingprogress.ListResult{},
	}
	r := newTestRouter(t, svc, authenticatorFor(t, user))

	rec := getProgress(t, r, comicID.String(), true)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got continueJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body not decodable (%q): %v", rec.Body.String(), err)
	}

	if got.Continue != nil {
		t.Errorf("continue = %+v, want null", got.Continue)
	}

	if !bytes.Contains(rec.Body.Bytes(), []byte(`"continue":null`)) {
		t.Errorf("body = %s, want continue:null", rec.Body.String())
	}

	if bytes.Contains(rec.Body.Bytes(), []byte(`"chapters"`)) {
		t.Errorf("body = %s, must not include chapters", rec.Body.String())
	}
}

func TestPutOK(t *testing.T) {
	t.Parallel()

	user := testUser()
	chapterID := uuid.New()
	savedAt := time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC)

	svc := &stubService{
		saveResult: readingprogress.Progress{UpdatedAt: savedAt, ChapterID: chapterID, Page: 20},
	}
	r := newTestRouter(t, svc, authenticatorFor(t, user))

	rec := putProgress(t, r, chapterID.String(), `{"page":8}`, true)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	if svc.lastSave.UserID != user.ID {
		t.Errorf("lastSave.UserID = %s, want %s", svc.lastSave.UserID, user.ID)
	}

	if svc.lastSave.ChapterID != chapterID {
		t.Errorf("lastSave.ChapterID = %s, want %s", svc.lastSave.ChapterID, chapterID)
	}

	if svc.lastSave.Page != 8 {
		t.Errorf("lastSave.Page = %d, want 8", svc.lastSave.Page)
	}

	var got progressJSON
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body not decodable (%q): %v", rec.Body.String(), err)
	}

	if got.ChapterID != chapterID.String() {
		t.Errorf("chapterId = %q, want %q", got.ChapterID, chapterID.String())
	}

	if got.Page != 20 {
		t.Errorf("page = %d, want 20", got.Page)
	}
}

func TestPutInvalidChapterID(t *testing.T) {
	t.Parallel()

	user := testUser()
	r := newTestRouter(t, &stubService{}, authenticatorFor(t, user))

	rec := putProgress(t, r, "not-a-uuid", `{"page":1}`, true)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestPutInvalidJSON(t *testing.T) {
	t.Parallel()

	user := testUser()
	r := newTestRouter(t, &stubService{}, authenticatorFor(t, user))

	rec := putProgress(t, r, uuid.New().String(), `{`, true)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestPutPageZero(t *testing.T) {
	t.Parallel()

	user := testUser()
	r := newTestRouter(t, &stubService{}, authenticatorFor(t, user))

	rec := putProgress(t, r, uuid.New().String(), `{"page":0}`, true)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestPutForbidden(t *testing.T) {
	t.Parallel()

	user := testUser()
	svc := &stubService{saveErr: domain.ErrForbidden}
	r := newTestRouter(t, svc, authenticatorFor(t, user))

	rec := putProgress(t, r, uuid.New().String(), `{"page":1}`, true)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusForbidden, rec.Body.String())
	}

	var got struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode error body: %v", err)
	}

	if got.Message != "comic not in library" {
		t.Errorf("message = %q, want %q", got.Message, "comic not in library")
	}
}

func TestGetNotFound(t *testing.T) {
	t.Parallel()

	user := testUser()
	svc := &stubService{listErr: domain.ErrNotFound}
	r := newTestRouter(t, svc, authenticatorFor(t, user))

	rec := getProgress(t, r, uuid.New().String(), true)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestPutUsesSessionUser(t *testing.T) {
	t.Parallel()

	user := testUser()
	chapterID := uuid.New()

	svc := &stubService{
		saveResult: readingprogress.Progress{UpdatedAt: time.Now().UTC(), ChapterID: chapterID, Page: 3},
	}
	r := newTestRouter(t, svc, authenticatorFor(t, user))

	rec := putProgress(t, r, chapterID.String(), `{"page":3}`, true)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	if svc.lastSave.UserID != user.ID {
		t.Errorf("lastSave.UserID = %s, want session user %s", svc.lastSave.UserID, user.ID)
	}
}
