// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth"
	httpauth "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	httpoidcproviders "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	sessionshttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	httpchapters "github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	httpcomics "github.com/kharente-deuh/uchiyomi-server/pkg/core/comics/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
	httpcovers "github.com/kharente-deuh/uchiyomi-server/pkg/core/covers/gateway/http"
	coversasurascans "github.com/kharente-deuh/uchiyomi-server/pkg/core/covers/source/asurascans"
	coredomain "github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/feed"
	httpfeed "github.com/kharente-deuh/uchiyomi-server/pkg/core/feed/gateway/http"
	healthhttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/health/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readersettings"
	httpreadersettings "github.com/kharente-deuh/uchiyomi-server/pkg/core/readersettings/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readingprogress"
	httpreadingprogress "github.com/kharente-deuh/uchiyomi-server/pkg/core/readingprogress/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/setup"
	httpsetup "github.com/kharente-deuh/uchiyomi-server/pkg/core/setup/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	httpusers "github.com/kharente-deuh/uchiyomi-server/pkg/core/users/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	asurascans "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/core"
	asurascansdomain "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
	httpasurascans "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/gateway/http"
	kingofshojo "github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/core"
	kingofshojodomain "github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/domain"
	httpkingofshojo "github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/gateway/http"
	kingofshojoparse "github.com/kharente-deuh/uchiyomi-server/pkg/sources/kingofshojo/pkg/parse"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/fncache"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/health"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/imgcache"
)

const (
	runDeadline = 10 * time.Second

	pollInterval = 2 * time.Millisecond

	notImplemented = "not implemented"
)

type fakeDB struct {
	migrate func() error
	ping    func(context.Context) error
}

func (f *fakeDB) Migrate() error {
	if f.migrate == nil {
		return nil
	}

	return f.migrate()
}

func (f *fakeDB) Ping(ctx context.Context) error {
	if f.ping == nil {
		return nil
	}

	return f.ping(ctx)
}

func (f *fakeDB) Close() {}

func newBlockedDB() (*fakeDB, func()) {
	release := make(chan struct{})
	once := &sync.Once{}

	db := &fakeDB{migrate: func() error {
		<-release

		return nil
	}}

	return db, func() { once.Do(func() { close(release) }) }
}

type fakeSessionsRepository struct{}

func (fakeSessionsRepository) Insert(
	context.Context, sessions.InsertSessionOpts,
) (*sessions.Session, error) {
	return nil, errors.New(notImplemented)
}

func (fakeSessionsRepository) GetByTokenHash(
	context.Context, []byte,
) (*sessions.Session, *users.User, error) {
	return nil, nil, errors.New(notImplemented)
}

func (fakeSessionsRepository) UpdateExpiry(context.Context, uuid.UUID, time.Time) error {
	return errors.New(notImplemented)
}

func (fakeSessionsRepository) DeleteByTokenHash(context.Context, []byte) error {
	return errors.New(notImplemented)
}

type stubCoverFinder struct{}

func (stubCoverFinder) FindBySourceSlug(context.Context, string, string) (uuid.UUID, error) {
	return uuid.Nil, coredomain.ErrNotFound
}

func newTestCoversBundle(t *testing.T, asuraScansApp *asurascans.App) (*covers.App, *covers.Service) {
	t.Helper()

	logger := slog.New(slog.DiscardHandler)

	httpClient := &http.Client{Timeout: time.Minute}

	asurascansResolver, err := coversasurascans.New(
		coversasurascans.Config{CDNBaseURL: coversasurascans.DefaultCDNBaseURL},
		coversasurascans.Deps{
			Getter:     asuraScansApp,
			HTTPClient: httpClient,
			Logger:     logger,
		},
	)
	if err != nil {
		t.Fatalf("coversasurascans.New: %v", err)
	}

	resolvers := map[string]covers.CoverResolver{
		covers.SourceAsuraScans: asurascansResolver,
	}

	cache, err := imgcache.New(imgcache.Config{
		Dir:           t.TempDir(),
		FetchFn:       covers.NewFetchFn(resolvers),
		ErrorCacheTTL: time.Minute,
		MinInterval:   time.Millisecond,
		Logger:        logger,
	})
	if err != nil {
		t.Fatalf("imgcache.New: %v", err)
	}

	svc, err := covers.NewService(
		covers.ServiceConfig{
			ProxyPathPrefix: APIPrefix + "/sources/cover",
			DownloadsDir:    t.TempDir(),
		},
		covers.ServiceDeps{
			Cache:      cache,
			Resolvers:  resolvers,
			HTTPClient: httpClient,
			Logger:     logger,
			Finder:     stubCoverFinder{},
		},
	)
	if err != nil {
		t.Fatalf("covers.NewService: %v", err)
	}

	app, err := covers.NewApp(covers.AppDeps{Cache: cache})
	if err != nil {
		t.Fatalf("covers.NewApp: %v", err)
	}

	return app, svc
}

type fakeOIDCRevalidationApp struct{}

func (fakeOIDCRevalidationApp) Run(ctx context.Context) error {
	<-ctx.Done()

	return fmt.Errorf("context done: %w", ctx.Err())
}

type fakeDownloadsApp struct{}

func (fakeDownloadsApp) Run(ctx context.Context) error {
	<-ctx.Done()

	return fmt.Errorf("context done: %w", ctx.Err())
}

type fakeChapterListRefreshApp struct{}

func (fakeChapterListRefreshApp) Run(ctx context.Context) error {
	<-ctx.Done()

	return fmt.Errorf("context done: %w", ctx.Err())
}

func (fakeSessionsRepository) DeleteByUserAndProvider(context.Context, uuid.UUID, uuid.UUID) error {
	return errors.New(notImplemented)
}

func (fakeSessionsRepository) DeleteByProviderAndSID(context.Context, uuid.UUID, string) error {
	return errors.New(notImplemented)
}

func (fakeSessionsRepository) DeleteByUserID(context.Context, uuid.UUID) error {
	return errors.New(notImplemented)
}

func (fakeSessionsRepository) DeleteExpired(context.Context, time.Time) (int64, error) {
	return 0, nil
}

type fakeSetupService struct{}

func (fakeSetupService) IsSetupRequired(context.Context) (bool, error) { return false, nil }

func (fakeSetupService) DoSetup(context.Context, setup.DoSetupOpts) (*sessions.IssuedSession, error) {
	return nil, errors.New(notImplemented)
}

type fakeAuthService struct{}

func (fakeAuthService) LoginWithPwd(context.Context, auth.LoginWithPwdOpts) (*auth.LoginResult, error) {
	return nil, errors.New(notImplemented)
}

func (fakeAuthService) CreateUserWithPwd(context.Context, auth.CreateUserWithPwdOpts) (*users.User, error) {
	return nil, errors.New(notImplemented)
}

func (fakeAuthService) Logout(context.Context, auth.LogoutOpts) (*auth.LogoutResult, error) {
	return nil, errors.New(notImplemented)
}

func (fakeAuthService) StartOIDCLogin(context.Context, auth.StartOIDCLoginOpts) (*auth.OIDCStart, error) {
	return nil, errors.New(notImplemented)
}

//nolint:lll
func (fakeAuthService) FinishOIDCLogin(context.Context, auth.FinishOIDCLoginOpts) (*auth.OIDCLoginResult, error) {
	return nil, errors.New(notImplemented)
}

func (fakeAuthService) BackchannelLogout(context.Context, string) error {
	return errors.New(notImplemented)
}

type fakeOIDCProvidersService struct{}

func (fakeOIDCProvidersService) List(context.Context) ([]oidcproviders.LightOIDCProvider, error) {
	return nil, nil
}

//nolint:lll
func (fakeOIDCProvidersService) GetByID(context.Context, uuid.UUID) (*oidcproviders.OIDCProviderDetails, error) {
	return nil, errors.New(notImplemented)
}

//nolint:lll
func (fakeOIDCProvidersService) Create(context.Context, oidcproviders.CreateOpts) (*oidcproviders.OIDCProvider, error) {
	return nil, errors.New(notImplemented)
}

//nolint:lll
func (fakeOIDCProvidersService) Update(context.Context, uuid.UUID, oidcproviders.UpdateOpts) (*oidcproviders.OIDCProvider, error) {
	return nil, errors.New(notImplemented)
}

func (fakeOIDCProvidersService) Delete(context.Context, uuid.UUID) error {
	return errors.New(notImplemented)
}

func (fakeOIDCProvidersService) Probe(context.Context, string) (*oidcproviders.ProbeResult, error) {
	return nil, errors.New(notImplemented)
}

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

type fakeComicsService struct{}

func (fakeComicsService) Create(context.Context, comics.CreateOpts) (*comics.Comic, error) {
	return nil, errors.New(notImplemented)
}

func (fakeComicsService) GetByID(context.Context, comics.GetByIDOpts) (*comics.Comic, error) {
	return nil, errors.New(notImplemented)
}

func (fakeComicsService) GetMany(context.Context, comics.GetManyOpts) (comics.Page, error) {
	return comics.Page{}, errors.New(notImplemented)
}

func (fakeComicsService) Delete(context.Context, comics.DeleteOpts) error {
	return errors.New(notImplemented)
}

func (fakeComicsService) RefreshChapterLists(context.Context) error {
	return errors.New(notImplemented)
}

func (fakeComicsService) RefreshComic(context.Context, comics.RefreshComicOpts) (*comics.Comic, error) {
	return nil, errors.New(notImplemented)
}

func (fakeComicsService) RetryChapters(context.Context, comics.RetryChaptersOpts) error {
	return errors.New(notImplemented)
}

func (fakeComicsService) ServeCover(context.Context, comics.GetByIDOpts) (string, string, error) {
	return "", "", errors.New(notImplemented)
}

type fakeChaptersService struct{}

func (fakeChaptersService) CreateAll(
	context.Context, uuid.UUID, []sources.SourceChapter,
) ([]chapters.Chapter, error) {
	return nil, errors.New(notImplemented)
}

func (fakeChaptersService) ListByComicID(context.Context, uuid.UUID) ([]chapters.Chapter, error) {
	return nil, errors.New(notImplemented)
}

func (fakeChaptersService) EnqueueDownloadable(context.Context, []chapters.Chapter) error {
	return errors.New(notImplemented)
}

func (fakeChaptersService) EnqueueResumable(context.Context) error {
	return errors.New(notImplemented)
}

func (fakeChaptersService) ScanEarlyAccess(context.Context) error {
	return errors.New(notImplemented)
}

func (fakeChaptersService) CleanupComic(context.Context, uuid.UUID, []chapters.Chapter) error {
	return errors.New(notImplemented)
}

func (fakeChaptersService) RetryDownload(context.Context, chapters.RetryDownloadOpts) error {
	return errors.New(notImplemented)
}

func (fakeChaptersService) GetByIds(context.Context, chapters.GetByIdsOpts) ([]chapters.Chapter, error) {
	return nil, errors.New(notImplemented)
}

func (fakeChaptersService) ListForLibrary(context.Context, chapters.ListForLibraryOpts) ([]chapters.Chapter, error) {
	return nil, errors.New(notImplemented)
}

func (fakeChaptersService) GetForLibrary(context.Context, chapters.GetForLibraryOpts) (*chapters.Chapter, error) {
	return nil, errors.New(notImplemented)
}

func (fakeChaptersService) GetDetailForLibrary(
	context.Context, chapters.GetForLibraryOpts,
) (*chapters.ChapterDetail, error) {
	return nil, errors.New(notImplemented)
}

func (fakeChaptersService) ServePage(context.Context, chapters.ServePageOpts) (string, string, error) {
	return "", "", errors.New(notImplemented)
}

type fakeFeedService struct{}

func (fakeFeedService) Get(context.Context, feed.GetOpts) (feed.Page, error) {
	return feed.Page{Items: []feed.Item{}, Total: 0}, nil
}

type fakeReaderSettingsService struct{}

func (fakeReaderSettingsService) ListForUser(context.Context, uuid.UUID) ([]readersettings.Profile, error) {
	return []readersettings.Profile{}, nil
}

func (fakeReaderSettingsService) Replace(context.Context, readersettings.ReplaceOpts) (readersettings.Profile, error) {
	return readersettings.Profile{}, nil
}

type fakeReadingProgressService struct{}

func (fakeReadingProgressService) List(context.Context, readingprogress.ListOpts) (readingprogress.ListResult, error) {
	return readingprogress.ListResult{}, nil
}

func (fakeReadingProgressService) MapByChapterIDs(
	context.Context, readingprogress.MapOpts,
) (map[uuid.UUID]readingprogress.Progress, error) {
	return map[uuid.UUID]readingprogress.Progress{}, nil
}

func (fakeReadingProgressService) Save(context.Context, readingprogress.SaveOpts) (readingprogress.Progress, error) {
	return readingprogress.Progress{}, nil
}

func (fakeReadingProgressService) SetRead(
	context.Context, readingprogress.SetReadOpts,
) (readingprogress.ListResult, error) {
	return readingprogress.ListResult{}, nil
}

func (fakeReadingProgressService) Delete(
	context.Context, readingprogress.DeleteOpts,
) error {
	return nil
}

func newTestCache[P any, T any](t *testing.T, name string, logger *slog.Logger) *fncache.Cache[P, T] {
	t.Helper()

	c, err := fncache.New(
		fncache.Config[P, T]{
			Name:          name,
			Fn:            func(context.Context, P) (*T, error) { return nil, errors.New(notImplemented) },
			Key:           func(P) string { return "" },
			TTL:           time.Minute,
			ErrorTTL:      time.Minute,
			FetchTimeout:  time.Minute,
			CleanInterval: time.Minute,
			MaxEntries:    1,
		},
		fncache.Deps{Logger: logger},
	)
	if err != nil {
		t.Fatalf("fncache.New(%s): %v", name, err)
	}

	return c
}

func newTestAsuraScans(t *testing.T, logger *slog.Logger) *asurascans.App {
	t.Helper()

	app, err := asurascans.New(
		asurascans.Config{SourceName: sources.SourceAsuraScans},
		asurascans.Deps{
			Logger: logger,
			SearchCache: newTestCache[asurascansdomain.SearchCacheOpts, asurascansdomain.SearchCacheResult](
				t, "search", logger,
			),
			GetInfosBySlugCache: newTestCache[string, asurascansdomain.GetInfosBySlugResponse](
				t, "infos", logger,
			),
			GetChaptersListBySeriesCache: newTestCache[string, []asurascansdomain.Chapter](
				t, "chapters", logger,
			),
			GetImageURLsByChapter: newTestCache[asurascansdomain.GetImageURLsByChapterOpts, []string](
				t, "images", logger,
			),
			ComicsRepository: stubComicsRepository{},
		},
	)
	if err != nil {
		t.Fatalf("asurascans.New: %v", err)
	}

	return app
}

func newTestKingOfShojo(t *testing.T, logger *slog.Logger) *kingofshojo.App {
	t.Helper()

	app, err := kingofshojo.New(
		kingofshojo.Config{SourceName: sources.SourceKingOfShojo, BaseURL: "https://kingofshojo.com"},
		kingofshojo.Deps{
			Logger: logger,
			SearchCache: newTestCache[kingofshojodomain.SearchCacheOpts, kingofshojodomain.SearchCacheResult](
				t, "kos-search", logger,
			),
			SeriesCache: newTestCache[string, kingofshojoparse.SeriesPage](
				t, "kos-series", logger,
			),
			GetImageURLsByChapter: newTestCache[kingofshojodomain.GetImageURLsByChapterOpts, []string](
				t, "kos-images", logger,
			),
			ComicsRepository: stubComicsRepository{},
		},
	)
	if err != nil {
		t.Fatalf("kingofshojo.New: %v", err)
	}

	return app
}

func newTestApp(t *testing.T, db Database, port int) (*App, *health.Registry) {
	t.Helper()

	logger := slog.New(slog.DiscardHandler)
	registry := NewHealthRegistry(db)
	asuraScansApp := newTestAsuraScans(t, logger)
	kingOfShojoApp := newTestKingOfShojo(t, logger)

	sessionsApp, err := sessions.New(
		sessions.Config{RemoveExpiredSessionsInterval: time.Hour},
		sessions.Deps{Logger: logger, SessionsRepository: fakeSessionsRepository{}},
	)
	if err != nil {
		t.Fatalf("sessions.New: %v", err)
	}

	cookies, err := sessionshttp.NewCookieManager(sessionshttp.CookieConfig{Name: "s", Path: "/"})
	if err != nil {
		t.Fatalf("NewCookieManager: %v", err)
	}

	oidcStateCookies, err := sessionshttp.NewCookieManager(sessionshttp.CookieConfig{Name: "oidc_state", Path: "/"})
	if err != nil {
		t.Fatalf("NewCookieManager: %v", err)
	}

	setupCtrl, err := httpsetup.New(
		httpsetup.Config{Endpoint: "/setup"},
		httpsetup.Deps{Logger: logger, SetupService: fakeSetupService{}, Cookies: cookies},
	)
	if err != nil {
		t.Fatalf("httpsetup.New: %v", err)
	}

	authCtrl, err := httpauth.New(
		httpauth.Config{Endpoint: "/auth"},
		httpauth.Deps{
			Logger:           logger,
			AuthService:      fakeAuthService{},
			Cookies:          cookies,
			OIDCStateCookies: oidcStateCookies,
			ProvidersLister:  fakeOIDCProvidersService{},
		},
	)
	if err != nil {
		t.Fatalf("httpauth.New: %v", err)
	}

	usersCtrl, err := httpusers.New(
		httpusers.Config{Endpoint: "/users"},
		httpusers.Deps{Logger: logger},
	)
	if err != nil {
		t.Fatalf("httpusers.New: %v", err)
	}

	coversApp, coversService := newTestCoversBundle(t, asuraScansApp)

	asuraScansCtrl, err := httpasurascans.New(
		httpasurascans.Config{Endpoint: "/" + string(sources.SourceAsuraScans)},
		httpasurascans.Deps{
			Logger:        logger,
			AsuraScansApp: asuraScansApp,
			CoverURLBuilder: func(source, slug string) string {
				return coversService.BuildProxyURL(source, slug)
			},
		},
	)
	if err != nil {
		t.Fatalf("httpasurascans.New: %v", err)
	}

	kingOfShojoCtrl, err := httpkingofshojo.New(
		httpkingofshojo.Config{Endpoint: "/" + string(sources.SourceKingOfShojo)},
		httpkingofshojo.Deps{
			Logger:         logger,
			KingOfShojoApp: kingOfShojoApp,
			CoverURLBuilder: func(source, slug string) string {
				return coversService.BuildProxyURL(source, slug)
			},
		},
	)
	if err != nil {
		t.Fatalf("httpkingofshojo.New: %v", err)
	}

	coversCtrl, err := httpcovers.New(
		httpcovers.Config{Endpoint: "/cover"},
		httpcovers.Deps{Service: coversService, Logger: logger},
	)
	if err != nil {
		t.Fatalf("httpcovers.New: %v", err)
	}

	healthCtrl, err := healthhttp.New(
		healthhttp.Config{ProbeTimeout: healthhttp.DefaultProbeTimeout},
		healthhttp.Deps{Registry: registry, Logger: logger},
	)
	if err != nil {
		t.Fatalf("healthhttp.New: %v", err)
	}

	oidcProvidersCtrl, err := httpoidcproviders.New(
		httpoidcproviders.Config{Endpoint: "/oidc/providers"},
		httpoidcproviders.Deps{Logger: logger, Service: fakeOIDCProvidersService{}},
	)
	if err != nil {
		t.Fatalf("httpoidcproviders.New: %v", err)
	}

	comicsCtrl, err := httpcomics.New(
		httpcomics.Config{Endpoint: "/comics"},
		httpcomics.Deps{Logger: logger, ComicsService: fakeComicsService{}},
	)
	if err != nil {
		t.Fatalf("httpcomics.New: %v", err)
	}

	chaptersCtrl, err := httpchapters.New(
		httpchapters.Config{Endpoint: "/chapters"},
		httpchapters.Deps{
			Logger:          logger,
			ChaptersService: fakeChaptersService{},
			Progress:        fakeReadingProgressService{},
		},
	)
	if err != nil {
		t.Fatalf("httpchapters.New: %v", err)
	}

	feedCtrl, err := httpfeed.New(
		httpfeed.Config{Endpoint: "/feed"},
		httpfeed.Deps{Logger: logger, FeedService: fakeFeedService{}},
	)
	if err != nil {
		t.Fatalf("httpfeed.New: %v", err)
	}

	readerSettingsCtrl, err := httpreadersettings.New(
		httpreadersettings.Config{Endpoint: "/me"},
		httpreadersettings.Deps{Logger: logger, Service: fakeReaderSettingsService{}},
	)
	if err != nil {
		t.Fatalf("httpreadersettings.New: %v", err)
	}

	readingProgressCtrl, err := httpreadingprogress.New(
		httpreadingprogress.Config{Endpoint: "/comics", ChaptersEndpoint: "/chapters"},
		httpreadingprogress.Deps{Logger: logger, Service: fakeReadingProgressService{}},
	)
	if err != nil {
		t.Fatalf("httpreadingprogress.New: %v", err)
	}

	app, err := New(
		Config{ServerPort: port, AllowedOrigins: []string{"*"}},
		Deps{
			DB:                  db,
			SetupCtrl:           setupCtrl,
			AuthCtrl:            authCtrl,
			UsersCtrl:           usersCtrl,
			AsuraScansCtrl:      asuraScansCtrl,
			KingOfShojoCtrl:     kingOfShojoCtrl,
			CoversCtrl:          coversCtrl,
			HealthCtrl:          healthCtrl,
			OIDCProvidersCtrl:   oidcProvidersCtrl,
			ComicsCtrl:          comicsCtrl,
			ChaptersCtrl:        chaptersCtrl,
			FeedCtrl:            feedCtrl,
			ReaderSettingsCtrl:  readerSettingsCtrl,
			ReadingProgressCtrl: readingProgressCtrl,
			Logger:              logger,
			Health:              registry,
			AsuraScans:          asuraScansApp,
			KingOfShojo:         kingOfShojoApp,
			Covers:              coversApp,
			Downloads:           fakeDownloadsApp{},
			ChapterListRefresh:  fakeChapterListRefreshApp{},
			Sessions:            sessionsApp,
			OIDCRevalidation:    fakeOIDCRevalidationApp{},
		})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return app, registry
}

type runningApp struct {
	done    chan struct{}
	cancel  context.CancelFunc
	runErr  error
	baseURL string
}

func freePort(t *testing.T) int {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}

	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("adresse inattendue: %T", l.Addr())
	}

	if err := l.Close(); err != nil {
		t.Fatalf("l.Close: %v", err)
	}

	return addr.Port
}

func startApp(t *testing.T, db Database) *runningApp {
	t.Helper()

	port := freePort(t)
	app, _ := newTestApp(t, db, port)

	ctx, cancel := context.WithCancel(context.Background())
	r := &runningApp{
		done:    make(chan struct{}),
		cancel:  cancel,
		baseURL: "http://127.0.0.1:" + strconv.Itoa(port),
	}

	go func() {
		defer close(r.done)

		r.runErr = app.Run(ctx)
	}()

	t.Cleanup(func() { _ = r.wait(t) })

	return r
}

func startBlockedApp(t *testing.T) (*runningApp, func()) {
	t.Helper()

	db, unblock := newBlockedDB()
	app := startApp(t, db)

	t.Cleanup(unblock)

	return app, unblock
}

func (r *runningApp) wait(t *testing.T) error {
	t.Helper()

	r.cancel()

	select {
	case <-r.done:
		return r.runErr
	case <-time.After(runDeadline):
		t.Fatal("Run did not return after context cancellation")

		return nil
	}
}

func (r *runningApp) waitListening(t *testing.T) {
	t.Helper()

	r.waitFor(t, "server is not listening", func() bool {
		code, _ := r.get(t, "/healthz")

		return code == http.StatusOK
	})
}

func (r *runningApp) waitFor(t *testing.T, msg string, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(runDeadline)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}

		time.Sleep(pollInterval)
	}

	t.Fatal(msg)
}

func (r *runningApp) get(t *testing.T, path string) (int, []byte) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), runDeadline)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.baseURL+path, nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll: %v", err)
	}

	return resp.StatusCode, body
}

type componentBody struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type readyzBody struct {
	Components map[string]componentBody `json:"components"`
	Status     string                   `json:"status"`
}

func decodeReadyz(t *testing.T, raw []byte) readyzBody {
	t.Helper()

	var body readyzBody
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("decode of %q: %v", raw, err)
	}

	return body
}
