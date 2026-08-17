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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	sessionshttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	comicshttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/comics/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

const (
	comicsEndpoint = "/comics"
	cookieName     = "uchiyomi_session"
	testToken      = "letoken"
	testUsername   = "alice"
)

type stubComicsService struct {
	createErr        error
	getByIDErr       error
	getManyErr       error
	deleteErr        error
	serveCoverErr    error
	createResult     *comics.Comic
	getByIDResult    *comics.Comic
	serveCoverPath   string
	serveCoverType   string
	getManyResult    comics.Page
	serveCoverCalls  int
	lastServeCoverID uuid.UUID
}

func (s *stubComicsService) Create(_ context.Context, _ comics.CreateOpts) (*comics.Comic, error) {
	return s.createResult, s.createErr
}

func (s *stubComicsService) GetByID(_ context.Context, _ comics.GetByIDOpts) (*comics.Comic, error) {
	return s.getByIDResult, s.getByIDErr
}

func (s *stubComicsService) GetMany(_ context.Context, _ comics.GetManyOpts) (comics.Page, error) {
	return s.getManyResult, s.getManyErr
}

func (s *stubComicsService) Delete(_ context.Context, _ comics.DeleteOpts) error {
	return s.deleteErr
}

func (s *stubComicsService) RefreshChapterLists(context.Context) error {
	return nil
}

func (s *stubComicsService) ServeCover(_ context.Context, opts comics.GetByIDOpts) (string, string, error) {
	s.serveCoverCalls++
	s.lastServeCoverID = opts.ID

	return s.serveCoverPath, s.serveCoverType, s.serveCoverErr
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

func newTestRouter(t *testing.T, svc comics.ComicsService, mws chi.Middlewares) chi.Router {
	t.Helper()

	logger := slog.New(slog.DiscardHandler)

	c, err := comicshttp.New(
		comicshttp.Config{Endpoint: comicsEndpoint, Middlewares: mws},
		comicshttp.Deps{Logger: logger, ComicsService: svc},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	r := chi.NewRouter()
	c.InitRouter(r)

	return r
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	passthrough := func(next http.Handler) http.Handler { return next }

	tests := map[string]struct {
		wantErr string
		cfg     comicshttp.Config
	}{
		"valid": {cfg: comicshttp.Config{Endpoint: comicsEndpoint}},
		"empty": {cfg: comicshttp.Config{}, wantErr: "endpoint is required"},
		"without leading slash": {
			cfg:     comicshttp.Config{Endpoint: "comics"},
			wantErr: `endpoint must start with '/', got "comics"`,
		},
		"nil middleware": {
			cfg:     comicshttp.Config{Endpoint: comicsEndpoint, Middlewares: chi.Middlewares{passthrough, nil}},
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

func TestCreateRequiresAuthentication(t *testing.T) {
	t.Parallel()

	r := newTestRouter(t, &stubComicsService{}, nil)

	req := httptest.NewRequest(
		http.MethodPost,
		comicsEndpoint+"/",
		bytes.NewReader([]byte(`{"source":"asurascans","slug":"solo-leveling"}`)),
	)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestCreateReturnsComic(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comic := &comics.Comic{
		ID:           uuid.New(),
		Title:        "Solo Leveling",
		Slug:         "solo-leveling",
		Source:       sources.SourceAsuraScans,
		Status:       sources.SeriesStatusCompleted,
		ChapterCount: 200,
	}

	r := newTestRouter(t, &stubComicsService{createResult: comic}, authenticatorFor(t, user))

	req := httptest.NewRequest(
		http.MethodPost,
		comicsEndpoint+"/",
		bytes.NewReader([]byte(`{"source":"asurascans","slug":"solo-leveling"}`)),
	)

	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var got comicshttpLightComic
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.ID != comic.ID || got.Slug != comic.Slug {
		t.Errorf("response = %+v, want comic %+v", got, comic)
	}

	wantCover := "/api/comics/" + comic.ID.String() + "/cover"
	if got.Cover != wantCover {
		t.Errorf("cover = %q, want %q", got.Cover, wantCover)
	}
}

type comicshttpLightComic struct {
	Title        string               `json:"title"`
	Slug         string               `json:"slug"`
	Cover        string               `json:"cover"`
	Source       sources.SourceName   `json:"source"`
	Status       sources.SeriesStatus `json:"status"`
	ChapterCount int                  `json:"chapter_count"`
	ID           uuid.UUID            `json:"id"`
}

func TestGetByIDNotFound(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comicID := uuid.New()

	r := newTestRouter(t, &stubComicsService{
		getByIDErr: fmt.Errorf("s.deps.ComicsRepository.GetByID: %w", domain.ErrNotFound),
	}, authenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, comicsEndpoint+"/"+comicID.String(), nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetManyReturnsComics(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	items := []comics.Comic{{
		ID:     uuid.New(),
		Slug:   "solo-leveling",
		Source: sources.SourceAsuraScans,
		Title:  "Solo Leveling",
		Status: sources.SeriesStatusCompleted,
	}}

	r := newTestRouter(t, &stubComicsService{getManyResult: comics.Page{Items: items}}, authenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, comicsEndpoint+"/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got []comicshttpLightComic
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(got) != 1 || got[0].Slug != items[0].Slug {
		t.Errorf("response = %+v, want %+v", got, items)
	}

	wantCover := "/api/comics/" + items[0].ID.String() + "/cover"
	if got[0].Cover != wantCover {
		t.Errorf("cover = %q, want %q", got[0].Cover, wantCover)
	}
}

func TestServeCoverOK(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comicID := uuid.New()
	dir := t.TempDir()
	path := filepath.Join(dir, "cover.webp")
	if err := os.WriteFile(path, []byte("webp-bytes"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	r := newTestRouter(t, &stubComicsService{
		serveCoverPath: path,
		serveCoverType: "image/webp",
	}, authenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, comicsEndpoint+"/"+comicID.String()+"/cover", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	if rec.Header().Get("Content-Type") != "image/webp" {
		t.Errorf("Content-Type = %q", rec.Header().Get("Content-Type"))
	}

	if rec.Body.String() != "webp-bytes" {
		t.Errorf("body = %q", rec.Body.String())
	}
}

func TestServeCoverNotInLibrary(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comicID := uuid.New()
	r := newTestRouter(t, &stubComicsService{
		serveCoverErr: fmt.Errorf("s.GetByID: %w", domain.ErrNotFound),
	}, authenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, comicsEndpoint+"/"+comicID.String()+"/cover", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestServeCoverUnauthorized(t *testing.T) {
	t.Parallel()

	r := newTestRouter(t, &stubComicsService{}, authenticatorFor(t, &users.User{ID: uuid.New(), Name: testUsername}))

	req := httptest.NewRequest(http.MethodGet, comicsEndpoint+"/"+uuid.New().String()+"/cover", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestDeleteByID(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comicID := uuid.New()

	r := newTestRouter(t, &stubComicsService{}, authenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodDelete, comicsEndpoint+"/"+comicID.String(), nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestDeleteByIDInternalError(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comicID := uuid.New()
	sentinel := errors.New("db down")

	r := newTestRouter(t, &stubComicsService{deleteErr: sentinel}, authenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodDelete, comicsEndpoint+"/"+comicID.String(), nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
