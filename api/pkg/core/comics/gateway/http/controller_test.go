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
	"strings"
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
	testComicTitle = "Solo Leveling"
	testComicSlug  = "solo-leveling"
)

//nolint:govet // fieldalignment on a test fake is not worth the unreadable field order
type stubComicsService struct {
	createErr          error
	getByIDErr         error
	getManyErr         error
	deleteErr          error
	serveCoverErr      error
	refreshComicErr    error
	retryChaptersErr   error
	updateTypeErr      error
	createResult       *comics.Comic
	getByIDResult      *comics.Comic
	refreshComicResult *comics.Comic
	updateTypeResult   *comics.Comic
	getManyResult      comics.Page
	lastGetMany        comics.GetManyOpts
	lastRetryChapters  comics.RetryChaptersOpts
	serveCoverPath     string
	serveCoverType     string
	lastUpdateType     comics.UpdateTypeOpts
	lastServeCoverID   uuid.UUID
	serveCoverCalls    int
	updateTypeCalls    int
}

func (s *stubComicsService) Create(_ context.Context, _ comics.CreateOpts) (*comics.Comic, error) {
	return s.createResult, s.createErr
}

func (s *stubComicsService) GetByID(_ context.Context, _ comics.GetByIDOpts) (*comics.Comic, error) {
	return s.getByIDResult, s.getByIDErr
}

func (s *stubComicsService) GetMany(_ context.Context, opts comics.GetManyOpts) (comics.Page, error) {
	s.lastGetMany = opts

	return s.getManyResult, s.getManyErr
}

func (s *stubComicsService) Delete(_ context.Context, _ comics.DeleteOpts) error {
	return s.deleteErr
}

func (s *stubComicsService) RefreshChapterLists(context.Context) error {
	return nil
}

func (s *stubComicsService) RefreshComic(_ context.Context, _ comics.RefreshComicOpts) (*comics.Comic, error) {
	return s.refreshComicResult, s.refreshComicErr
}

func (s *stubComicsService) RetryChapters(_ context.Context, opts comics.RetryChaptersOpts) error {
	s.lastRetryChapters = opts

	return s.retryChaptersErr
}

func (s *stubComicsService) ServeCover(_ context.Context, opts comics.GetByIDOpts) (string, string, error) {
	s.serveCoverCalls++
	s.lastServeCoverID = opts.ID

	return s.serveCoverPath, s.serveCoverType, s.serveCoverErr
}

func (s *stubComicsService) UpdateType(_ context.Context, opts comics.UpdateTypeOpts) (*comics.Comic, error) {
	s.updateTypeCalls++
	s.lastUpdateType = opts

	return s.updateTypeResult, s.updateTypeErr
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
		bytes.NewReader([]byte(`{"source":"asurascans","slug":"`+testComicSlug+`"}`)),
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
		Title:        testComicTitle,
		Slug:         testComicSlug,
		Source:       sources.SourceAsuraScans,
		Status:       sources.SeriesStatusCompleted,
		ChapterCount: 200,
	}

	r := newTestRouter(t, &stubComicsService{createResult: comic}, authenticatorFor(t, user))

	req := httptest.NewRequest(
		http.MethodPost,
		comicsEndpoint+"/",
		bytes.NewReader([]byte(`{"source":"asurascans","slug":"`+testComicSlug+`"}`)),
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
	Type         sources.SeriesType   `json:"type"`
	ChapterCount int                  `json:"chapterCount"`
	ID           uuid.UUID            `json:"id"`
}

type comicshttpPage struct {
	Items []comicshttpLightComic `json:"items"`
	Total int64                  `json:"total"`
}

type comicshttpDetail struct {
	Type         sources.SeriesType `json:"type"`
	Cover        string             `json:"cover"`
	AltTitles    []string           `json:"altTitles"`
	ChapterCount int                `json:"chapterCount"`
	ID           uuid.UUID          `json:"id"`
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
	comicID := uuid.New()
	getManyResult := comics.Page{
		Items: []comics.Comic{{
			ID:           comicID,
			Slug:         testComicSlug,
			Source:       sources.SourceAsuraScans,
			Title:        testComicTitle,
			Status:       sources.SeriesStatusCompleted,
			Type:         sources.SeriesTypeManhwa,
			ChapterCount: 200,
		}},
		Total: 42,
	}

	r := newTestRouter(t, &stubComicsService{getManyResult: getManyResult}, authenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, comicsEndpoint+"/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got comicshttpPage
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Total != 42 {
		t.Errorf("total = %d, want 42", got.Total)
	}

	if len(got.Items) != 1 || got.Items[0].Slug != testComicSlug {
		t.Errorf("response = %+v, want %+v", got, getManyResult.Items)
	}

	if got.Items[0].Type != sources.SeriesTypeManhwa {
		t.Errorf("type = %q, want %q", got.Items[0].Type, sources.SeriesTypeManhwa)
	}

	if got.Items[0].ChapterCount != 200 {
		t.Errorf("chapterCount = %d, want 200", got.Items[0].ChapterCount)
	}

	wantCover := "/api/comics/" + comicID.String() + "/cover"
	if got.Items[0].Cover != wantCover {
		t.Errorf("cover = %q, want %q", got.Items[0].Cover, wantCover)
	}
}

func TestGetManyUnauthorized(t *testing.T) {
	t.Parallel()

	r := newTestRouter(t, &stubComicsService{}, nil)
	req := httptest.NewRequest(http.MethodGet, comicsEndpoint+"/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGetManyInvalidQuery(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	r := newTestRouter(t, &stubComicsService{}, authenticatorFor(t, user))

	queries := []string{
		"?sort=created_at",
		"?order=ASC",
		"?source=not-a-source",
		"?type=webtoon",
		"?status=unknown",
		"?limit=x",
		"?offset=x",
	}

	for _, q := range queries {
		t.Run(q, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, comicsEndpoint+"/?"+strings.TrimPrefix(q, "?"), nil)
			req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
		})
	}
}

func TestGetManyPassesSearchSortAndUserID(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	svc := &stubComicsService{getManyResult: comics.Page{Items: []comics.Comic{}, Total: 0}}
	r := newTestRouter(t, svc, authenticatorFor(t, user))

	req := httptest.NewRequest(
		http.MethodGet,
		comicsEndpoint+"/?search=%20solo%20&sort=addedAt&order=desc&limit=5&offset=2",
		nil,
	)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	if svc.lastGetMany.Search != "solo" {
		t.Errorf("Search = %q, want solo", svc.lastGetMany.Search)
	}

	if svc.lastGetMany.Sort != comics.ListSortAddedAt || svc.lastGetMany.Order != comics.ListOrderDesc {
		t.Errorf("sort/order = %s %s", svc.lastGetMany.Sort, svc.lastGetMany.Order)
	}

	if svc.lastGetMany.UserID == nil || *svc.lastGetMany.UserID != user.ID {
		t.Errorf("UserID = %v, want %s", svc.lastGetMany.UserID, user.ID)
	}

	if svc.lastGetMany.Limit != 5 || svc.lastGetMany.Offset != 2 {
		t.Errorf("limit/offset = %d %d", svc.lastGetMany.Limit, svc.lastGetMany.Offset)
	}
}

func TestGetByIDCamelCaseJSON(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comic := &comics.Comic{
		ID:           uuid.New(),
		Title:        testComicTitle,
		Slug:         testComicSlug,
		AltTitles:    []string{"Na Honjaman Level-Up"},
		ChapterCount: 200,
		Source:       sources.SourceAsuraScans,
		Status:       sources.SeriesStatusCompleted,
		Type:         sources.SeriesTypeManhwa,
	}

	r := newTestRouter(t, &stubComicsService{getByIDResult: comic}, authenticatorFor(t, user))
	req := httptest.NewRequest(http.MethodGet, comicsEndpoint+"/"+comic.ID.String(), nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	var got comicshttpDetail
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if got.ChapterCount != 200 || len(got.AltTitles) != 1 {
		t.Errorf("detail = %+v", got)
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

func TestRefreshRequiresAuthentication(t *testing.T) {
	t.Parallel()

	r := newTestRouter(t, &stubComicsService{}, nil)
	req := httptest.NewRequest(http.MethodPost, comicsEndpoint+"/"+uuid.New().String()+"/refresh", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRefreshInvalidID(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	r := newTestRouter(t, &stubComicsService{}, authenticatorFor(t, user))
	req := httptest.NewRequest(http.MethodPost, comicsEndpoint+"/not-a-uuid/refresh", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRefreshNotFound(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comicID := uuid.New()
	r := newTestRouter(t, &stubComicsService{
		refreshComicErr: fmt.Errorf("s.deps.ComicsRepository.FindByID: %w", domain.ErrNotFound),
	}, authenticatorFor(t, user))
	req := httptest.NewRequest(http.MethodPost, comicsEndpoint+"/"+comicID.String()+"/refresh", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestRefreshForbidden(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comicID := uuid.New()
	r := newTestRouter(t, &stubComicsService{refreshComicErr: domain.ErrForbidden}, authenticatorFor(t, user))
	req := httptest.NewRequest(http.MethodPost, comicsEndpoint+"/"+comicID.String()+"/refresh", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestRefreshConflict(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comicID := uuid.New()
	r := newTestRouter(t, &stubComicsService{refreshComicErr: domain.ErrConflict}, authenticatorFor(t, user))
	req := httptest.NewRequest(http.MethodPost, comicsEndpoint+"/"+comicID.String()+"/refresh", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestRefreshSourceUnavailable(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comicID := uuid.New()
	r := newTestRouter(t, &stubComicsService{
		refreshComicErr: fmt.Errorf("src.GetInfosBySlug: %w", comics.ErrSourceUnavailable),
	}, authenticatorFor(t, user))
	req := httptest.NewRequest(http.MethodPost, comicsEndpoint+"/"+comicID.String()+"/refresh", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadGateway)
	}
}

func TestRefreshOK(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comic := &comics.Comic{
		ID:           uuid.New(),
		Title:        testComicTitle,
		Slug:         testComicSlug,
		Source:       sources.SourceAsuraScans,
		Status:       sources.SeriesStatusOngoing,
		Type:         sources.SeriesTypeManhwa,
		ChapterCount: 42,
	}

	r := newTestRouter(t, &stubComicsService{refreshComicResult: comic}, authenticatorFor(t, user))
	req := httptest.NewRequest(http.MethodPost, comicsEndpoint+"/"+comic.ID.String()+"/refresh", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got comicshttpDetail
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if got.ID != comic.ID || got.ChapterCount != 42 {
		t.Errorf("detail = %+v, want id=%s chapterCount=42", got, comic.ID)
	}
}

func TestRefreshInternalError(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comicID := uuid.New()
	sentinel := errors.New("db down")
	r := newTestRouter(t, &stubComicsService{refreshComicErr: sentinel}, authenticatorFor(t, user))
	req := httptest.NewRequest(http.MethodPost, comicsEndpoint+"/"+comicID.String()+"/refresh", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestRetryChaptersRequiresAuthentication(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	r := newTestRouter(t, &stubComicsService{}, nil)

	body := strings.NewReader(`{"chapterIds":["` + uuid.New().String() + `"]}`)
	req := httptest.NewRequest(http.MethodPost, comicsEndpoint+"/"+comicID.String()+"/retry", body)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRetryChaptersInvalidComicID(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	r := newTestRouter(t, &stubComicsService{}, authenticatorFor(t, user))

	body := strings.NewReader(`{"chapterIds":["` + uuid.New().String() + `"]}`)
	req := httptest.NewRequest(http.MethodPost, comicsEndpoint+"/invalid-uuid/retry", body)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRetryChaptersEmptyChapterIDs(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	r := newTestRouter(t, &stubComicsService{}, authenticatorFor(t, user))

	req := httptest.NewRequest(
		http.MethodPost,
		comicsEndpoint+"/"+uuid.New().String()+"/retry",
		strings.NewReader(`{"chapterIds":[]}`),
	)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRetryChaptersNotFound(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comicID := uuid.New()
	r := newTestRouter(t, &stubComicsService{
		retryChaptersErr: fmt.Errorf("s.deps.ComicsRepository.FindByID: %w", domain.ErrNotFound),
	}, authenticatorFor(t, user))

	body := strings.NewReader(`{"chapterIds":["` + uuid.New().String() + `"]}`)
	req := httptest.NewRequest(http.MethodPost, comicsEndpoint+"/"+comicID.String()+"/retry", body)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestRetryChaptersForbidden(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comicID := uuid.New()
	r := newTestRouter(t, &stubComicsService{
		retryChaptersErr: domain.ErrForbidden,
	}, authenticatorFor(t, user))

	body := strings.NewReader(`{"chapterIds":["` + uuid.New().String() + `"]}`)
	req := httptest.NewRequest(http.MethodPost, comicsEndpoint+"/"+comicID.String()+"/retry", body)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestRetryChaptersAccepted(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	comicID := uuid.New()
	chapterID := uuid.New()
	svc := &stubComicsService{}
	r := newTestRouter(t, svc, authenticatorFor(t, user))

	body := strings.NewReader(`{"chapterIds":["` + chapterID.String() + `"]}`)
	req := httptest.NewRequest(http.MethodPost, comicsEndpoint+"/"+comicID.String()+"/retry", body)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}

	last := svc.lastRetryChapters
	if last.ComicID != comicID || last.UserID != user.ID || len(last.ChapterIDs) != 1 || last.ChapterIDs[0] != chapterID {
		t.Errorf("unexpected lastRetryChapters: %+v", last)
	}
}

func TestUpdateTypeSuccess(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	userID := uuid.New()
	user := &users.User{ID: userID, Name: testUsername}
	stub := &stubComicsService{
		updateTypeResult: &comics.Comic{
			ID:           id,
			Title:        testComicTitle,
			Slug:         testComicSlug,
			Source:       sources.SourceKingOfShojo,
			Status:       sources.SeriesStatusOngoing,
			Type:         sources.SeriesTypeManga,
			ChapterCount: 10,
		},
	}
	r := newTestRouter(t, stub, authenticatorFor(t, user))

	body := []byte(`{"type":"manga"}`)
	req := httptest.NewRequest(http.MethodPatch, comicsEndpoint+"/"+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var res comicshttpDetail
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("json decode: %v", err)
	}

	if res.Type != sources.SeriesTypeManga {
		t.Errorf("res.Type = %q, want %q", res.Type, sources.SeriesTypeManga)
	}
	if stub.lastUpdateType.ID != id {
		t.Errorf("stub.lastUpdateType.ID = %v, want %v", stub.lastUpdateType.ID, id)
	}
	if stub.lastUpdateType.UserID != userID {
		t.Errorf("stub.lastUpdateType.UserID = %v, want %v", stub.lastUpdateType.UserID, userID)
	}
}

func TestUpdateTypeInvalidUUID(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: testUsername}
	r := newTestRouter(t, &stubComicsService{}, authenticatorFor(t, user))

	body := []byte(`{"type":"manga"}`)
	req := httptest.NewRequest(http.MethodPatch, comicsEndpoint+"/invalid-uuid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateTypeInvalidType(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	user := &users.User{ID: uuid.New(), Name: testUsername}
	r := newTestRouter(t, &stubComicsService{}, authenticatorFor(t, user))

	body := []byte(`{"type":"invalid_type"}`)
	req := httptest.NewRequest(http.MethodPatch, comicsEndpoint+"/"+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateTypeNotFound(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	user := &users.User{ID: uuid.New(), Name: testUsername}
	stub := &stubComicsService{
		updateTypeErr: domain.ErrNotFound,
	}
	r := newTestRouter(t, stub, authenticatorFor(t, user))

	body := []byte(`{"type":"manga"}`)
	req := httptest.NewRequest(http.MethodPatch, comicsEndpoint+"/"+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestUpdateTypeForbidden(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	user := &users.User{ID: uuid.New(), Name: testUsername}
	stub := &stubComicsService{
		updateTypeErr: domain.ErrForbidden,
	}
	r := newTestRouter(t, stub, authenticatorFor(t, user))

	body := []byte(`{"type":"manga"}`)
	req := httptest.NewRequest(http.MethodPatch, comicsEndpoint+"/"+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: testToken})
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
