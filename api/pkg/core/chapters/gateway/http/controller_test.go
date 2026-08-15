// SPDX-License-Identifier: AGPL-3.0-or-later

package http_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	sessionshttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	chaptershttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
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
	retryErr error
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

func newChaptersTestRouter(t *testing.T, svc chapters.ChaptersService, mws chi.Middlewares) chi.Router {
	t.Helper()

	c, err := chaptershttp.New(
		chaptershttp.Config{Endpoint: chaptersEndpoint, Middlewares: mws},
		chaptershttp.Deps{Logger: chaptersTestLogger(), ChaptersService: svc},
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
			},
		},
		"missing service": {
			deps:    chaptershttp.Deps{Logger: logger},
			wantErr: "chapters service is required",
		},
		"missing logger": {
			deps:    chaptershttp.Deps{ChaptersService: &stubChaptersService{}},
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
		chaptershttp.Deps{Logger: chaptersTestLogger(), ChaptersService: &stubChaptersService{}},
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
