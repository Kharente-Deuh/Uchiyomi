// SPDX-License-Identifier: AGPL-3.0-or-later

package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/feed"
	feedhttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/feed/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

const (
	feedEndpoint = "/feed"
	cookieName   = "uchiyomi_session"
	testToken    = "letoken"
	testUsername = "alice"
)

type stubFeedService struct {
	getErr    error
	getResult feed.Page
	lastGet   feed.GetOpts
}

func (s *stubFeedService) Get(_ context.Context, opts feed.GetOpts) (feed.Page, error) {
	s.lastGet = opts

	return s.getResult, s.getErr
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

func newTestRouter(t *testing.T, svc feed.FeedService, mws chi.Middlewares) chi.Router {
	t.Helper()

	logger := slog.New(slog.DiscardHandler)

	c, err := feedhttp.New(
		feedhttp.Config{Endpoint: feedEndpoint, Middlewares: mws},
		feedhttp.Deps{Logger: logger, FeedService: svc},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	r := chi.NewRouter()
	c.InitRouter(r)

	return r
}

type feedItemResponse struct {
	Title          string                `json:"title"`
	Slug           string                `json:"slug"`
	Cover          string                `json:"cover"`
	Source         sources.SourceName    `json:"source"`
	Status         sources.SeriesStatus  `json:"status"`
	Type           sources.SeriesType    `json:"type"`
	LatestChapters []feedChapterResponse `json:"latestChapters"`
	ID             uuid.UUID             `json:"id"`
}

type feedChapterResponse struct {
	PublishedAt      time.Time  `json:"publishedAt"`
	EarlyAccessUntil *time.Time `json:"earlyAccessUntil"`
	Title            string     `json:"title"`
	Number           float64    `json:"number"`
	HasProgress      bool       `json:"hasProgress"`
	Download         int        `json:"download"`
	ID               uuid.UUID  `json:"id"`
}

type feedPageResponse struct {
	Items []feedItemResponse `json:"items"`
	Total int64              `json:"total"`
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	passthrough := func(next http.Handler) http.Handler { return next }

	tests := map[string]struct {
		wantErr string
		cfg     feedhttp.Config
	}{
		"valid": {cfg: feedhttp.Config{Endpoint: feedEndpoint}},
		"empty": {cfg: feedhttp.Config{}, wantErr: "endpoint is required"},
		"without leading slash": {
			cfg:     feedhttp.Config{Endpoint: "feed"},
			wantErr: `endpoint must start with '/', got "feed"`,
		},
		"nil middleware": {
			cfg:     feedhttp.Config{Endpoint: feedEndpoint, Middlewares: chi.Middlewares{passthrough, nil}},
			wantErr: "all middlewares must not be nil",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := tc.cfg.Validate()

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

	logger := slog.New(slog.DiscardHandler)

	tests := map[string]struct {
		deps    feedhttp.Deps
		wantErr string
	}{
		"valid": {
			deps: feedhttp.Deps{
				Logger:      logger,
				FeedService: &stubFeedService{},
			},
		},
		"missing service": {
			deps:    feedhttp.Deps{Logger: logger},
			wantErr: "feed service is required",
		},
		"missing logger": {
			deps:    feedhttp.Deps{FeedService: &stubFeedService{}},
			wantErr: "logger is required",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := tc.deps.Validate()

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

func TestGetUnauthorized(t *testing.T) {
	t.Parallel()

	r := newTestRouter(t, &stubFeedService{}, nil)

	req := httptest.NewRequest(http.MethodGet, feedEndpoint+"/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGetInvalidQuery(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	r := newTestRouter(t, &stubFeedService{}, authenticatorFor(t, user))

	queries := []string{
		"?source=not-a-source",
		"?type=webtoon",
		"?limit=x",
		"?offset=x",
	}

	for _, q := range queries {
		t.Run(q, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, feedEndpoint+"/?"+strings.TrimPrefix(q, "?"), nil)
			req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
		})
	}
}

func TestGetDefaultsLimitAndOffset(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	svc := &stubFeedService{getResult: feed.Page{Items: []feed.Item{}, Total: 0}}
	r := newTestRouter(t, svc, authenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, feedEndpoint+"/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	if svc.lastGet.Limit != 10 {
		t.Errorf("Limit = %d, want 10", svc.lastGet.Limit)
	}

	if svc.lastGet.Offset != 0 {
		t.Errorf("Offset = %d, want 0", svc.lastGet.Offset)
	}

	if svc.lastGet.UserID != user.ID {
		t.Errorf("UserID = %v, want %s", svc.lastGet.UserID, user.ID)
	}
}

func TestGetClampsLimit(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}

	tests := map[string]struct {
		query string
		want  int
	}{
		"limit zero": {query: "?limit=0", want: 10},
		"limit huge": {query: "?limit=500", want: 100},
		"offset neg": {query: "?offset=-1", want: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := &stubFeedService{getResult: feed.Page{Items: []feed.Item{}, Total: 0}}
			r := newTestRouter(t, svc, authenticatorFor(t, user))

			req := httptest.NewRequest(http.MethodGet, feedEndpoint+"/"+tc.query, nil)
			req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}

			got := svc.lastGet.Limit
			if strings.Contains(tc.query, "offset") {
				got = svc.lastGet.Offset
			}

			if got != tc.want {
				t.Errorf("got = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestGetJSON(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comicID := uuid.New()
	chapterID := uuid.New()
	publishedAt := time.Date(2026, time.January, 15, 8, 0, 0, 0, time.UTC)
	earlyAccessUntil := time.Date(2026, time.January, 20, 8, 0, 0, 0, time.UTC)

	getResult := feed.Page{
		Items: []feed.Item{{
			ID:     comicID,
			Title:  "",
			Slug:   "solo-leveling",
			Source: sources.SourceAsuraScans,
			Status: sources.SeriesStatusCompleted,
			Type:   sources.SeriesTypeManhwa,
			LatestChapters: []feed.LatestChapter{{
				ID:               chapterID,
				ComicID:          comicID,
				Title:            "Chapter 1",
				Number:           1,
				PublishedAt:      publishedAt,
				EarlyAccessUntil: &earlyAccessUntil,
				Download:         2,
			}},
		}},
		Total: 1,
	}

	r := newTestRouter(t, &stubFeedService{getResult: getResult}, authenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, feedEndpoint+"/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	var got feedPageResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Total != 1 {
		t.Errorf("total = %d, want 1", got.Total)
	}

	if len(got.Items) != 1 {
		t.Fatalf("items len = %d, want 1", len(got.Items))
	}

	item := got.Items[0]
	wantCover := "/api/comics/" + comicID.String() + "/cover"
	if item.Cover != wantCover {
		t.Errorf("cover = %q, want %q", item.Cover, wantCover)
	}

	if item.Title != "" {
		t.Errorf("title = %q, want empty string", item.Title)
	}

	if len(item.LatestChapters) != 1 {
		t.Fatalf("latestChapters len = %d, want 1", len(item.LatestChapters))
	}

	if item.LatestChapters[0].Download != 2 {
		t.Errorf("download = %d, want 2", item.LatestChapters[0].Download)
	}

	if item.LatestChapters[0].HasProgress {
		t.Error("hasProgress = true, want false")
	}

	if item.LatestChapters[0].EarlyAccessUntil == nil || !item.LatestChapters[0].EarlyAccessUntil.Equal(earlyAccessUntil) {
		t.Errorf("earlyAccessUntil = %v, want %v", item.LatestChapters[0].EarlyAccessUntil, earlyAccessUntil)
	}
}

func TestGetJSONOmitsZeroEarlyAccessUntil(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comicID := uuid.New()
	chapterID := uuid.New()
	publishedAt := time.Date(2026, time.March, 19, 1, 28, 20, 0, time.UTC)

	r := newTestRouter(t, &stubFeedService{getResult: feed.Page{
		Items: []feed.Item{{
			ID:     comicID,
			Slug:   "the-greatest-estate-developer",
			Source: sources.SourceAsuraScans,
			Status: sources.SeriesStatusCompleted,
			Type:   sources.SeriesTypeManhwa,
			LatestChapters: []feed.LatestChapter{{
				ID:          chapterID,
				ComicID:     comicID,
				Title:       "Special Chapter [END]",
				Number:      223,
				PublishedAt: publishedAt,
				Download:    100,
			}},
		}},
		Total: 1,
	}}, authenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, feedEndpoint+"/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if strings.Contains(body, `"earlyAccessUntil"`) {
		t.Errorf("body = %s, want earlyAccessUntil omitted", body)
	}

	var got feedPageResponse
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Items[0].LatestChapters[0].EarlyAccessUntil != nil {
		t.Errorf("earlyAccessUntil = %v, want nil", got.Items[0].LatestChapters[0].EarlyAccessUntil)
	}

	if !strings.Contains(body, `"hasProgress":false`) {
		t.Errorf("body = %s, want hasProgress:false present", body)
	}
}

func TestGetEmptyItemsIsArray(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	r := newTestRouter(t, &stubFeedService{
		getResult: feed.Page{Items: nil, Total: 0},
	}, authenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, feedEndpoint+"/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"items":[]`) {
		t.Errorf("body = %s, want items:[]", body)
	}
}

func TestGetServiceError(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	sentinel := errors.New("db down")

	r := newTestRouter(t, &stubFeedService{getErr: sentinel}, authenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, feedEndpoint+"/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
