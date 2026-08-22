// SPDX-License-Identifier: AGPL-3.0-or-later

//nolint:goconst,lll
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
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	chaptershttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readingprogress"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

const (
	chaptersEndpoint = "/chapters"
	chaptersCookie   = "uchiyomi_session"
	chaptersToken    = "letoken"
	chaptersUsername = "alice"
)

type stubChaptersService struct {
	retryErr                    error
	getByIdsErr                 error
	listForLibraryErr           error
	getForLibraryErr            error
	getDetailForLibraryErr      error
	servePageErr                error
	getByIdsResult              []chapters.Chapter
	listForLibraryResult        []chapters.Chapter
	getForLibraryResult         *chapters.Chapter
	getDetailForLibraryResult   *chapters.ChapterDetail
	servePagePath               string
	servePageType               string
	lastGetByIdsOpts            chapters.GetByIdsOpts
	lastListForLibraryOpts      chapters.ListForLibraryOpts
	lastGetForLibraryOpts       chapters.GetForLibraryOpts
	lastGetDetailForLibraryOpts chapters.GetForLibraryOpts
	lastServePageOpts           chapters.ServePageOpts
}

func (s *stubChaptersService) CreateAll(
	context.Context, uuid.UUID, []sources.SourceChapter,
) ([]chapters.Chapter, error) {
	return nil, errors.New("not implemented")
}

func (s *stubChaptersService) ListByComicID(context.Context, uuid.UUID) ([]chapters.Chapter, error) {
	return nil, errors.New("not implemented")
}

func (s *stubChaptersService) EnqueueDownloadable(context.Context, []chapters.Chapter) error {
	return errors.New("not implemented")
}

func (s *stubChaptersService) EnqueueResumable(context.Context) error {
	return errors.New("not implemented")
}

func (s *stubChaptersService) ScanEarlyAccess(context.Context) error {
	return errors.New("not implemented")
}

func (s *stubChaptersService) CleanupComic(context.Context, uuid.UUID, []chapters.Chapter) error {
	return errors.New("not implemented")
}

func (s *stubChaptersService) RetryDownload(_ context.Context, _ chapters.RetryDownloadOpts) error {
	return s.retryErr
}

func (s *stubChaptersService) GetByIds(_ context.Context, opts chapters.GetByIdsOpts) ([]chapters.Chapter, error) {
	s.lastGetByIdsOpts = opts

	return s.getByIdsResult, s.getByIdsErr
}

func (s *stubChaptersService) ListForLibrary(
	_ context.Context, opts chapters.ListForLibraryOpts,
) ([]chapters.Chapter, error) {
	s.lastListForLibraryOpts = opts

	return s.listForLibraryResult, s.listForLibraryErr
}

func (s *stubChaptersService) GetForLibrary(
	_ context.Context, opts chapters.GetForLibraryOpts,
) (*chapters.Chapter, error) {
	s.lastGetForLibraryOpts = opts

	return s.getForLibraryResult, s.getForLibraryErr
}

func (s *stubChaptersService) GetDetailForLibrary(
	_ context.Context, opts chapters.GetForLibraryOpts,
) (*chapters.ChapterDetail, error) {
	s.lastGetDetailForLibraryOpts = opts

	return s.getDetailForLibraryResult, s.getDetailForLibraryErr
}

func (s *stubChaptersService) ServePage(_ context.Context, opts chapters.ServePageOpts) (string, string, error) {
	s.lastServePageOpts = opts

	return s.servePagePath, s.servePageType, s.servePageErr
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

func chaptersTestLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func chaptersFrozenNow() time.Time {
	return time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
}

func chaptersAuthenticatorFor(t *testing.T, user *users.User) chi.Middlewares {
	t.Helper()

	logger := chaptersTestLogger()

	cookies, err := sessionshttp.NewCookieManager(sessionshttp.CookieConfig{Name: chaptersCookie, Path: "/"})
	if err != nil {
		t.Fatalf("NewCookieManager: %v", err)
	}

	a, err := sessionshttp.NewAuthenticator(sessionshttp.AuthenticatorDeps{
		SessionService: &stubSessionService{result: &sessions.AuthenticatedSession{
			User:    user,
			Session: sessions.Session{ID: uuid.New(), UserID: user.ID, ExpiresAt: chaptersFrozenNow().Add(time.Hour)},
		}},
		Cookies: cookies,
		Logger:  logger,
		Now:     chaptersFrozenNow,
	})
	if err != nil {
		t.Fatalf("NewAuthenticator: %v", err)
	}

	return chi.Middlewares{a.Middleware}
}

type stubProgress struct {
	err  error
	byID map[uuid.UUID]readingprogress.Progress
	last readingprogress.MapOpts
}

func (s *stubProgress) MapByChapterIDs(
	_ context.Context, opts readingprogress.MapOpts,
) (map[uuid.UUID]readingprogress.Progress, error) {
	s.last = opts
	if s.err != nil {
		return nil, s.err
	}

	if s.byID != nil {
		return s.byID, nil
	}

	return map[uuid.UUID]readingprogress.Progress{}, nil
}

func newChaptersTestRouter(t *testing.T, svc chapters.ChaptersService, mws chi.Middlewares) chi.Router {
	t.Helper()

	return newChaptersTestRouterWithProgress(t, svc, &stubProgress{}, mws)
}

func newChaptersTestRouterWithProgress(
	t *testing.T, svc chapters.ChaptersService, progress chaptershttp.ProgressReader, mws chi.Middlewares,
) chi.Router {
	t.Helper()

	c, err := chaptershttp.New(
		chaptershttp.Config{Endpoint: chaptersEndpoint, Middlewares: mws},
		chaptershttp.Deps{
			Logger:          chaptersTestLogger(),
			ChaptersService: svc,
			Progress:        progress,
		},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	r := chi.NewRouter()
	c.InitRouter(r)

	return r
}

func TestRetryDownloadRequiresAuthentication(t *testing.T) {
	t.Parallel()

	chapterID := uuid.New()
	r := newChaptersTestRouter(t, &stubChaptersService{}, nil)

	req := httptest.NewRequest(http.MethodPost, chaptersEndpoint+"/"+chapterID.String()+"/retry", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRetryDownloadInvalidID(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	r := newChaptersTestRouter(t, &stubChaptersService{}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodPost, chaptersEndpoint+"/not-a-uuid/retry", nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRetryDownloadNotFound(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	chapterID := uuid.New()

	r := newChaptersTestRouter(t, &stubChaptersService{
		retryErr: fmt.Errorf("s.deps.Repository.GetByID: %w", domain.ErrNotFound),
	}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodPost, chaptersEndpoint+"/"+chapterID.String()+"/retry", nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestRetryDownloadForbidden(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	chapterID := uuid.New()

	r := newChaptersTestRouter(t, &stubChaptersService{retryErr: domain.ErrForbidden}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodPost, chaptersEndpoint+"/"+chapterID.String()+"/retry", nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestRetryDownloadConflict(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	chapterID := uuid.New()

	r := newChaptersTestRouter(t, &stubChaptersService{retryErr: domain.ErrConflict}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodPost, chaptersEndpoint+"/"+chapterID.String()+"/retry", nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestRetryDownloadAccepted(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	chapterID := uuid.New()

	r := newChaptersTestRouter(t, &stubChaptersService{}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodPost, chaptersEndpoint+"/"+chapterID.String()+"/retry", nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusAccepted, rec.Body.String())
	}
}

func TestRetryDownloadInternalError(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	chapterID := uuid.New()
	sentinel := errors.New("db down")

	r := newChaptersTestRouter(t, &stubChaptersService{retryErr: sentinel}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodPost, chaptersEndpoint+"/"+chapterID.String()+"/retry", nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

//nolint:govet // JSON test DTO; field order matches the API payload, not alignment
type postListHTTPChapter struct {
	PublishedAt      time.Time  `json:"publishedAt"`
	EarlyAccessUntil *time.Time `json:"earlyAccessUntil"`
	Progress         *struct {
		UpdatedAt time.Time `json:"updatedAt"`
		Page      int       `json:"page"`
	} `json:"progress"`
	Next *struct {
		Title  string    `json:"title"`
		Number float64   `json:"number"`
		ID     uuid.UUID `json:"id"`
	} `json:"next"`
	Previous *struct {
		Title  string    `json:"title"`
		Number float64   `json:"number"`
		ID     uuid.UUID `json:"id"`
	} `json:"previous"`
	PageURLs          []string  `json:"pageUrls"`
	SourceChapterSlug string    `json:"sourceChapterSlug"`
	Title             string    `json:"title"`
	Number            float64   `json:"number"`
	PagesNb           int       `json:"pagesNb"`
	Download          int       `json:"download"`
	ID                uuid.UUID `json:"id"`
	ComicID           uuid.UUID `json:"comicId"`
}

func TestPostListRequiresAuthentication(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	r := newChaptersTestRouter(t, &stubChaptersService{}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(
		http.MethodPost,
		chaptersEndpoint+"/list",
		strings.NewReader(`{"ids":["`+uuid.New().String()+`"]}`),
	)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestPostListInvalidBody(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	r := newChaptersTestRouter(t, &stubChaptersService{}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodPost, chaptersEndpoint+"/list", strings.NewReader(`{`))
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestPostListEmptyIDs(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	r := newChaptersTestRouter(t, &stubChaptersService{}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodPost, chaptersEndpoint+"/list", strings.NewReader(`{"ids":[]}`))
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestPostListServiceError(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	r := newChaptersTestRouter(t, &stubChaptersService{
		getByIdsErr: errors.New("db down"),
	}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(
		http.MethodPost,
		chaptersEndpoint+"/list",
		strings.NewReader(`{"ids":["`+uuid.New().String()+`"]}`),
	)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

func TestPostListReturnsChapters(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	chapterID := uuid.New()
	comicID := uuid.New()
	publishedAt := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	earlyAccessUntil := time.Date(2024, 2, 3, 4, 5, 6, 0, time.UTC)
	svc := &stubChaptersService{
		getByIdsResult: []chapters.Chapter{
			{
				PublishedAt:       publishedAt,
				EarlyAccessUntil:  &earlyAccessUntil,
				SourceChapterSlug: "chapter-1",
				Title:             "Chapter 1",
				Number:            1.5,
				PagesNb:           42,
				Download:          80,
				ID:                chapterID,
				ComicID:           comicID,
			},
			{
				PublishedAt:       publishedAt,
				SourceChapterSlug: "chapter-2",
				Title:             "Chapter 2",
				Number:            2,
				ID:                uuid.New(),
				ComicID:           comicID,
			},
		},
	}
	r := newChaptersTestRouter(t, svc, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(
		http.MethodPost,
		chaptersEndpoint+"/list",
		strings.NewReader(`{"ids":["`+chapterID.String()+`"]}`),
	)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	if svc.lastGetByIdsOpts.UserID != user.ID {
		t.Errorf("GetByIds userID = %s, want %s", svc.lastGetByIdsOpts.UserID, user.ID)
	}

	if len(svc.lastGetByIdsOpts.IDs) != 1 || svc.lastGetByIdsOpts.IDs[0] != chapterID {
		t.Errorf("GetByIds ids = %v, want [%s]", svc.lastGetByIdsOpts.IDs, chapterID)
	}

	var got []postListHTTPChapter
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len(chapters) = %d, want 2", len(got))
	}

	first := got[0]
	if first.ID != chapterID || first.ComicID != comicID || first.Title != "Chapter 1" {
		t.Errorf("chapters[0] = %+v", first)
	}

	if first.Number != 1.5 || first.PagesNb != 42 || first.Download != 80 || first.SourceChapterSlug != "chapter-1" {
		t.Errorf("chapters[0] fields = %+v", first)
	}

	if !first.PublishedAt.Equal(publishedAt) {
		t.Errorf("publishedAt = %v, want %v", first.PublishedAt, publishedAt)
	}

	if first.EarlyAccessUntil == nil || !first.EarlyAccessUntil.Equal(earlyAccessUntil) {
		t.Errorf("earlyAccessUntil = %v, want %v", first.EarlyAccessUntil, earlyAccessUntil)
	}

	if got[1].EarlyAccessUntil != nil {
		t.Errorf("chapters[1].earlyAccessUntil = %v, want nil", got[1].EarlyAccessUntil)
	}

	if first.Progress != nil || got[1].Progress != nil {
		t.Errorf("progress = %+v / %+v, want null", first.Progress, got[1].Progress)
	}
}

func TestListForLibraryRequiresAuthentication(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	r := newChaptersTestRouter(t, &stubChaptersService{}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint+"?comicId="+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestListForLibraryMissingComicID(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	r := newChaptersTestRouter(t, &stubChaptersService{}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint, nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestListForLibraryInvalidComicID(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	r := newChaptersTestRouter(t, &stubChaptersService{}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint+"?comicId=not-a-uuid", nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestListForLibraryNotFoundWhenComicMissing(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	comicID := uuid.New()
	r := newChaptersTestRouter(t, &stubChaptersService{
		listForLibraryErr: fmt.Errorf("s.deps.ComicLookup.Exists: %w", domain.ErrNotFound),
	}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint+"?comicId="+comicID.String(), nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d body=%s (404 when comic is missing)", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestListForLibraryForbiddenWhenNotInLibrary(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	comicID := uuid.New()
	r := newChaptersTestRouter(t, &stubChaptersService{
		listForLibraryErr: domain.ErrForbidden,
	}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint+"?comicId="+comicID.String(), nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d body=%s (403 when comic exists but is not in the user's library)", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestListForLibraryInternalError(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	r := newChaptersTestRouter(t, &stubChaptersService{
		listForLibraryErr: errors.New("db down"),
	}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint+"?comicId="+uuid.New().String(), nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestListForLibraryReturnsChapters(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	chapterID := uuid.New()
	comicID := uuid.New()
	publishedAt := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	earlyAccessUntil := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	svc := &stubChaptersService{
		listForLibraryResult: []chapters.Chapter{
			{
				PublishedAt:       publishedAt,
				EarlyAccessUntil:  &earlyAccessUntil,
				SourceChapterSlug: "chapter-3",
				Title:             "Locked",
				Number:            3,
				PagesNb:           12,
				Download:          0,
				ID:                chapterID,
				ComicID:           comicID,
			},
			{
				PublishedAt:       publishedAt,
				SourceChapterSlug: "chapter-1",
				Title:             "Chapter 1",
				Number:            1,
				PagesNb:           20,
				Download:          100,
				ID:                uuid.New(),
				ComicID:           comicID,
			},
		},
	}
	r := newChaptersTestRouter(t, svc, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint+"?comicId="+comicID.String(), nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	if svc.lastListForLibraryOpts.UserID != user.ID || svc.lastListForLibraryOpts.ComicID != comicID {
		t.Errorf("ListForLibrary opts = %+v, want user=%s comic=%s", svc.lastListForLibraryOpts, user.ID, comicID)
	}

	var got []postListHTTPChapter
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len(chapters) = %d, want 2", len(got))
	}

	if got[0].ID != chapterID || got[0].Number != 3 || got[0].Download != 0 {
		t.Errorf("chapters[0] = %+v", got[0])
	}

	if got[0].EarlyAccessUntil == nil || !got[0].EarlyAccessUntil.Equal(earlyAccessUntil) {
		t.Errorf("locked chapter earlyAccessUntil = %v, want %v", got[0].EarlyAccessUntil, earlyAccessUntil)
	}

	if got[1].EarlyAccessUntil != nil || got[1].Download != 100 {
		t.Errorf("chapters[1] = %+v", got[1])
	}

	if got[0].Progress != nil || got[1].Progress != nil {
		t.Errorf("progress = %+v / %+v, want null", got[0].Progress, got[1].Progress)
	}
}

func TestPostListIncludesProgress(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	chapterID := uuid.New()
	updatedAt := time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC)
	svc := &stubChaptersService{
		getByIdsResult: []chapters.Chapter{{
			ID:      chapterID,
			ComicID: uuid.New(),
			PagesNb: 10,
		}},
	}
	progress := &stubProgress{
		byID: map[uuid.UUID]readingprogress.Progress{
			chapterID: {ChapterID: chapterID, Page: 18, UpdatedAt: updatedAt},
		},
	}
	r := newChaptersTestRouterWithProgress(t, svc, progress, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(
		http.MethodPost,
		chaptersEndpoint+"/list",
		strings.NewReader(`{"ids":["`+chapterID.String()+`"]}`),
	)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var got []postListHTTPChapter
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(got) != 1 || got[0].Progress == nil {
		t.Fatalf("chapters = %+v, want progress", got)
	}

	if got[0].Progress.Page != 10 {
		t.Errorf("progress.page = %d, want clamped 10", got[0].Progress.Page)
	}

	if !got[0].Progress.UpdatedAt.Equal(updatedAt) {
		t.Errorf("progress.updatedAt = %v, want %v", got[0].Progress.UpdatedAt, updatedAt)
	}

	if progress.last.UserID != user.ID || len(progress.last.IDs) != 1 || progress.last.IDs[0] != chapterID {
		t.Errorf("MapByChapterIDs opts = %+v", progress.last)
	}
}

func TestGetByIDRequiresAuthentication(t *testing.T) {
	t.Parallel()

	r := newChaptersTestRouter(t, &stubChaptersService{}, nil)

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint+"/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGetByIDInvalidID(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	r := newChaptersTestRouter(t, &stubChaptersService{}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint+"/not-a-uuid", nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	r := newChaptersTestRouter(t, &stubChaptersService{
		getDetailForLibraryErr: domain.ErrNotFound,
	}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint+"/"+uuid.New().String(), nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetByIDForbidden(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	r := newChaptersTestRouter(t, &stubChaptersService{
		getDetailForLibraryErr: domain.ErrForbidden,
	}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint+"/"+uuid.New().String(), nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestGetByIDReturnsChapterWithProgress(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	chapterID := uuid.New()
	comicID := uuid.New()
	publishedAt := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	updatedAt := time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC)
	svc := &stubChaptersService{
		getDetailForLibraryResult: &chapters.ChapterDetail{
			Chapter: chapters.Chapter{
				PublishedAt: publishedAt,
				Title:       "Chapter 1",
				Number:      1,
				PagesNb:     20,
				Download:    40,
				ID:          chapterID,
				ComicID:     comicID,
			},
		},
	}
	progress := &stubProgress{
		byID: map[uuid.UUID]readingprogress.Progress{
			chapterID: {ChapterID: chapterID, Page: 4, UpdatedAt: updatedAt},
		},
	}
	r := newChaptersTestRouterWithProgress(t, svc, progress, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint+"/"+chapterID.String(), nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	if svc.lastGetDetailForLibraryOpts.UserID != user.ID || svc.lastGetDetailForLibraryOpts.ChapterID != chapterID {
		t.Errorf("GetDetailForLibrary opts = %+v", svc.lastGetDetailForLibraryOpts)
	}

	var got postListHTTPChapter
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.ID != chapterID || got.ComicID != comicID || got.Download != 40 {
		t.Errorf("chapter = %+v", got)
	}

	if got.Progress == nil || got.Progress.Page != 4 || !got.Progress.UpdatedAt.Equal(updatedAt) {
		t.Errorf("progress = %+v", got.Progress)
	}

	if got.PageURLs == nil || len(got.PageURLs) != 0 {
		t.Errorf("pageUrls = %#v, want empty slice", got.PageURLs)
	}

	if got.Next != nil || got.Previous != nil {
		t.Errorf("neighbors = next %#v prev %#v, want omitted", got.Next, got.Previous)
	}
}

func TestGetByIDPageURLsAndNeighbors(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	chapterID := uuid.New()
	prevID := uuid.New()
	nextID := uuid.New()
	svc := &stubChaptersService{
		getDetailForLibraryResult: &chapters.ChapterDetail{
			Chapter: chapters.Chapter{
				ID:       chapterID,
				ComicID:  uuid.New(),
				Number:   2,
				PagesNb:  2,
				Download: 100,
			},
			Previous: &chapters.ChapterNeighbor{ID: prevID, Title: "One", Number: 1},
			Next:     &chapters.ChapterNeighbor{ID: nextID, Title: "Three", Number: 3},
		},
	}
	r := newChaptersTestRouter(t, svc, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint+"/"+chapterID.String(), nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	var got postListHTTPChapter
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}

	want0 := "/api/chapters/" + chapterID.String() + "/pages/1"
	want1 := "/api/chapters/" + chapterID.String() + "/pages/2"
	if len(got.PageURLs) != 2 || got.PageURLs[0] != want0 || got.PageURLs[1] != want1 {
		t.Errorf("pageUrls = %#v", got.PageURLs)
	}

	if got.Previous == nil || got.Previous.ID != prevID || got.Previous.Title != "One" || got.Previous.Number != 1 {
		t.Errorf("previous = %+v, want id=%s title=One number=1", got.Previous, prevID)
	}

	if got.Next == nil || got.Next.ID != nextID || got.Next.Title != "Three" || got.Next.Number != 3 {
		t.Errorf("next = %+v, want id=%s title=Three number=3", got.Next, nextID)
	}
}

func TestServePageOK(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	chapterID := uuid.New()
	dir := t.TempDir()
	path := filepath.Join(dir, "001.webp")
	if err := os.WriteFile(path, []byte("page-bytes"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	r := newChaptersTestRouter(t, &stubChaptersService{
		servePagePath: path,
		servePageType: "image/webp",
	}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint+"/"+chapterID.String()+"/pages/1", nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	if rec.Header().Get("Content-Type") != "image/webp" {
		t.Errorf("Content-Type = %q", rec.Header().Get("Content-Type"))
	}

	if rec.Body.String() != "page-bytes" {
		t.Errorf("body = %q", rec.Body.String())
	}
}

func TestServePageUnauthorized(t *testing.T) {
	t.Parallel()

	r := newChaptersTestRouter(t, &stubChaptersService{}, chaptersAuthenticatorFor(t, &users.User{
		ID: uuid.New(), Name: chaptersUsername,
	}))

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint+"/"+uuid.New().String()+"/pages/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestServePageNotFoundWhenIncomplete(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	r := newChaptersTestRouter(t, &stubChaptersService{
		servePageErr: domain.ErrNotFound,
	}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint+"/"+uuid.New().String()+"/pages/1", nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestServePageForbidden(t *testing.T) {
	t.Parallel()

	user := &users.User{ID: uuid.New(), Name: chaptersUsername}
	r := newChaptersTestRouter(t, &stubChaptersService{
		servePageErr: domain.ErrForbidden,
	}, chaptersAuthenticatorFor(t, user))

	req := httptest.NewRequest(http.MethodGet, chaptersEndpoint+"/"+uuid.New().String()+"/pages/1", nil)
	req.AddCookie(&http.Cookie{Name: chaptersCookie, Value: chaptersToken})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	passthrough := func(next http.Handler) http.Handler { return next }

	tests := map[string]struct {
		wantErr string
		cfg     chaptershttp.Config
	}{
		"valid": {cfg: chaptershttp.Config{Endpoint: chaptersEndpoint}},
		"empty": {cfg: chaptershttp.Config{}, wantErr: "endpoint is required"},
		"without leading slash": {
			cfg:     chaptershttp.Config{Endpoint: "chapters"},
			wantErr: `endpoint must start with '/', got "chapters"`,
		},
		"nil middleware": {
			cfg:     chaptershttp.Config{Endpoint: chaptersEndpoint, Middlewares: chi.Middlewares{passthrough, nil}},
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

	logger := chaptersTestLogger()

	tests := map[string]struct {
		deps    chaptershttp.Deps
		wantErr string
	}{
		"valid": {
			deps: chaptershttp.Deps{
				Logger:          logger,
				ChaptersService: &stubChaptersService{},
				Progress:        &stubProgress{},
			},
		},
		"missing service": {
			deps:    chaptershttp.Deps{Logger: logger, Progress: &stubProgress{}},
			wantErr: "chapters service is required",
		},
		"missing progress": {
			deps:    chaptershttp.Deps{Logger: logger, ChaptersService: &stubChaptersService{}},
			wantErr: "progress is required",
		},
		"missing logger": {
			deps:    chaptershttp.Deps{ChaptersService: &stubChaptersService{}, Progress: &stubProgress{}},
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

func TestNewRequiresValidConfig(t *testing.T) {
	t.Parallel()

	_, err := chaptershttp.New(
		chaptershttp.Config{},
		chaptershttp.Deps{
			Logger:          chaptersTestLogger(),
			ChaptersService: &stubChaptersService{},
			Progress:        &stubProgress{},
		},
	)
	if err == nil {
		t.Fatal("New with invalid config must fail")
	}

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	_, err = chaptershttp.New(
		chaptershttp.Config{Endpoint: chaptersEndpoint},
		chaptershttp.Deps{Logger: logger},
	)
	if err == nil {
		t.Fatal("New with invalid deps must fail")
	}
}
