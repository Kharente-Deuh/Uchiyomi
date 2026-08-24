// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
	coredomain "github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/feed"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readersettings"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readingprogress"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/health"
)

type stubComicsRepository struct{}

func (stubComicsRepository) GetByID(context.Context, comics.GetByIDOpts) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (stubComicsRepository) FindByID(context.Context, uuid.UUID) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (stubComicsRepository) GetBySourceSlug(
	context.Context, comics.GetBySourceSlugOpts,
) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (stubComicsRepository) FindBySourceSlug(
	context.Context, comics.FindBySourceSlugOpts,
) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (stubComicsRepository) Create(context.Context, comics.CreateComicOpts) (*comics.Comic, error) {
	return nil, coredomain.ErrNotFound
}

func (stubComicsRepository) GetBySlugsAndSource(
	context.Context, comics.GetBySlugsAndSource,
) ([]comics.Comic, error) {
	return nil, nil
}

func (stubComicsRepository) Delete(context.Context, uuid.UUID) error {
	return coredomain.ErrNotFound
}

func (stubComicsRepository) GetMany(context.Context, comics.GetManyOpts) (comics.Page, error) {
	return comics.Page{}, nil
}

func (stubComicsRepository) ListByStatuses(context.Context, comics.ListByStatusesOpts) ([]comics.Comic, error) {
	return nil, nil
}

func (stubComicsRepository) UpdateStatusAndChapterCount(context.Context, comics.UpdateStatusAndChapterCountOpts) error {
	return nil
}

type stubFeedService struct{}

func (stubFeedService) Get(context.Context, feed.GetOpts) (feed.Page, error) {
	return feed.Page{Items: []feed.Item{}, Total: 0}, nil
}

type stubReaderSettingsService struct{}

func (stubReaderSettingsService) ListForUser(context.Context, uuid.UUID) ([]readersettings.Profile, error) {
	return []readersettings.Profile{}, nil
}

func (stubReaderSettingsService) Replace(context.Context, readersettings.ReplaceOpts) (readersettings.Profile, error) {
	return readersettings.Profile{}, nil
}

type stubReadingProgressService struct{}

func (stubReadingProgressService) List(context.Context, readingprogress.ListOpts) (readingprogress.ListResult, error) {
	return readingprogress.ListResult{}, nil
}

func (stubReadingProgressService) MapByChapterIDs(
	context.Context, readingprogress.MapOpts,
) (map[uuid.UUID]readingprogress.Progress, error) {
	return map[uuid.UUID]readingprogress.Progress{}, nil
}

func (stubReadingProgressService) Save(context.Context, readingprogress.SaveOpts) (readingprogress.Progress, error) {
	return readingprogress.Progress{}, nil
}

func (stubReadingProgressService) SetRead(
	context.Context, readingprogress.SetReadOpts,
) (readingprogress.ListResult, error) {
	return readingprogress.ListResult{}, nil
}

func (stubReadingProgressService) Delete(
	context.Context, readingprogress.DeleteOpts,
) error {
	return nil
}

type emptyOIDCProvidersRepository struct{}

func (emptyOIDCProvidersRepository) GetByID(
	context.Context, uuid.UUID,
) (*oidcproviders.OIDCProvider, error) {
	return nil, errors.New("not implemented")
}

func (emptyOIDCProvidersRepository) GetByIssuerURL(
	context.Context, string,
) (*oidcproviders.OIDCProvider, error) {
	return nil, errors.New("not implemented")
}

func (emptyOIDCProvidersRepository) Create(
	context.Context, oidcproviders.CreateOIDCProviderOpts,
) (*oidcproviders.OIDCProvider, error) {
	return nil, errors.New("not implemented")
}

func (emptyOIDCProvidersRepository) Update(
	context.Context, uuid.UUID, oidcproviders.UpdateOIDCProviderOpts,
) (*oidcproviders.OIDCProvider, error) {
	return nil, errors.New("not implemented")
}

func (emptyOIDCProvidersRepository) DeleteByID(context.Context, uuid.UUID) error {
	return errors.New("not implemented")
}

func (emptyOIDCProvidersRepository) GetAll(context.Context) ([]oidcproviders.LightOIDCProvider, error) {
	return nil, nil
}

func (emptyOIDCProvidersRepository) GetUsers(
	context.Context, uuid.UUID,
) ([]oidcproviders.OIDCProviderUser, error) {
	return nil, nil
}

type noopCipher struct{}

func (noopCipher) Seal(plaintext []byte) ([]byte, error) { return plaintext, nil }

type noopDiscoverer struct{}

func (noopDiscoverer) Discover(context.Context, string) (*oidcproviders.Discovery, error) {
	return nil, errors.New("not implemented")
}

type noopCacheEvictor struct{}

func (noopCacheEvictor) Evict(uuid.UUID) {}

func newTestOIDCProvidersService(t *testing.T) *oidcproviders.Service {
	t.Helper()

	svc, err := oidcproviders.NewService(
		oidcproviders.ServiceConfig{RedirectURI: "https://manga.example.com/api/auth/oidc/callback"},
		oidcproviders.ServiceDeps{
			Repository: emptyOIDCProvidersRepository{},
			Cipher:     noopCipher{},
			Discoverer: noopDiscoverer{},
			Cache:      noopCacheEvictor{},
		},
	)
	if err != nil {
		t.Fatalf("oidcproviders.NewService: %v", err)
	}

	return svc
}

type fixedUserSessionsRepository struct {
	user *users.User
}

func (f *fixedUserSessionsRepository) Insert(
	context.Context, sessions.InsertSessionOpts,
) (*sessions.Session, error) {
	return nil, errors.New("not implemented")
}

func (f *fixedUserSessionsRepository) GetByTokenHash(
	context.Context, []byte,
) (*sessions.Session, *users.User, error) {
	return &sessions.Session{
		ID:         uuid.New(),
		UserID:     f.user.ID,
		AuthMethod: sessions.AuthMethodPassword,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(48 * time.Hour),
	}, f.user, nil
}

func (f *fixedUserSessionsRepository) UpdateExpiry(context.Context, uuid.UUID, time.Time) error {
	return errors.New("not implemented")
}

func (f *fixedUserSessionsRepository) DeleteByTokenHash(context.Context, []byte) error {
	return errors.New("not implemented")
}

func (f *fixedUserSessionsRepository) DeleteByUserID(context.Context, uuid.UUID) error {
	return errors.New("not implemented")
}

func (f *fixedUserSessionsRepository) DeleteByUserAndProvider(context.Context, uuid.UUID, uuid.UUID) error {
	return errors.New("not implemented")
}

func (f *fixedUserSessionsRepository) DeleteByProviderAndSID(context.Context, uuid.UUID, string) error {
	return errors.New("not implemented")
}

func (f *fixedUserSessionsRepository) DeleteExpired(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func newTestSessionsService(t *testing.T, user *users.User) *sessions.Service {
	t.Helper()

	svc, err := sessions.NewService(
		sessions.ServiceConfig{
			Password: sessions.TTL{
				Idle:     24 * time.Hour,
				Absolute: 30 * 24 * time.Hour,
			},
			OIDC: sessions.TTL{
				Idle:     24 * time.Hour,
				Absolute: 30 * 24 * time.Hour,
			},
			RenewThreshold: time.Hour,
		},
		sessions.ServiceDeps{Repository: &fixedUserSessionsRepository{user: user}},
	)
	if err != nil {
		t.Fatalf("sessions.NewService: %v", err)
	}

	return svc
}

func newTestCtrlsForUser(t *testing.T, user *users.User) *ctrls {
	t.Helper()

	logger := slog.New(slog.DiscardHandler)

	asuraScansApp, err := setupAsuraScans(logger, stubComicsRepository{})
	if err != nil {
		t.Fatalf("setupAsuraScans: %v", err)
	}

	coversBundle, err := setupCovers(coversDeps{
		Logger:           logger,
		CoversDir:        t.TempDir(),
		DownloadsDir:     t.TempDir(),
		AsuraScansApp:    asuraScansApp,
		ComicsRepository: stubComicsRepository{},
	})
	if err != nil {
		t.Fatalf("setupCovers: %v", err)
	}

	c, err := setupCtrls(ctrlsDeps{
		AsuraScansApp:          asuraScansApp,
		CoversService:          coversBundle.Service,
		SessionsService:        newTestSessionsService(t, user),
		Logger:                 logger,
		Registry:               health.NewRegistry(),
		OIDCProvidersService:   newTestOIDCProvidersService(t),
		FeedService:            stubFeedService{},
		ReaderSettingsService:  stubReaderSettingsService{},
		ReadingProgressService: stubReadingProgressService{},
	})
	if err != nil {
		t.Fatalf("setupCtrls: %v", err)
	}

	return c
}

func TestComicsCoverFinderUnknownSource(t *testing.T) {
	t.Parallel()

	finder := comicsCoverFinder{repo: stubComicsRepository{}}

	_, err := finder.FindBySourceSlug(context.Background(), "unknown", "solo-leveling")
	if !errors.Is(err, covers.ErrUnknownSource) {
		t.Errorf("FindBySourceSlug = %v, want ErrUnknownSource", err)
	}
}

func TestSetupCoversServeUnknownSource(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.DiscardHandler)

	asuraScansApp, err := setupAsuraScans(logger, stubComicsRepository{})
	if err != nil {
		t.Fatalf("setupAsuraScans: %v", err)
	}

	bundle, err := setupCovers(coversDeps{
		Logger:           logger,
		CoversDir:        t.TempDir(),
		DownloadsDir:     t.TempDir(),
		AsuraScansApp:    asuraScansApp,
		ComicsRepository: stubComicsRepository{},
	})
	if err != nil {
		t.Fatalf("setupCovers: %v", err)
	}

	_, _, err = bundle.Service.Serve(context.Background(), "unknown", "solo-leveling")
	if !errors.Is(err, covers.ErrUnknownSource) {
		t.Errorf("Serve = %v, want ErrUnknownSource", err)
	}
}

func TestOIDCProvidersController(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		user       *users.User
		wantStatus int
	}{
		"non-admin is rejected": {
			user:       &users.User{ID: uuid.New(), Name: "alice", IsAdmin: false},
			wantStatus: http.StatusForbidden,
		},
		"admin is let through": {
			user:       &users.User{ID: uuid.New(), Name: "root", IsAdmin: true},
			wantStatus: http.StatusOK,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c := newTestCtrlsForUser(t, tc.user)

			r := chi.NewRouter()
			c.OIDCProviders.InitRouter(r)

			req := httptest.NewRequest(http.MethodGet, "/oidc/providers", nil)
			req.AddCookie(&http.Cookie{Name: "uchiyomi_session", Value: "any-token"})

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("GET /oidc/providers = %d, want %d (body: %s)", rec.Code, tc.wantStatus, rec.Body.String())
			}
		})
	}
}
