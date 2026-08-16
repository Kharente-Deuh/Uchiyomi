// SPDX-License-Identifier: AGPL-3.0-or-later

//nolint:goconst,lll
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	core "github.com/kharente-deuh/uchiyomi-server/pkg/core"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	httpcomics "github.com/kharente-deuh/uchiyomi-server/pkg/core/comics/gateway/http"
	comicrefresh "github.com/kharente-deuh/uchiyomi-server/pkg/core/comics/refresh"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	asura "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/core"
	asuradomain "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
	asuraclient "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/transport/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/database"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/fncache"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/health"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/imgcache"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/logging"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/federatedidentities"
	pgfederatedidentities "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/federatedidentities/repository/pg"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/hash/bcrypthash"
	pgpwd "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/password/repository/pg"
	httpauth "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidc"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	httpoidcproviders "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders/gateway/http"
	pgoidcproviders "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders/repository/pg"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	httpsession "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	pgsessions "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/repository/pg"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters/download"
	httpchapters "github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters/gateway/http"
	pgchapters "github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters/repository/pg"
	pgcomics "github.com/kharente-deuh/uchiyomi-server/pkg/core/comics/repository/pg"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
	httpcovers "github.com/kharente-deuh/uchiyomi-server/pkg/core/covers/gateway/http"
	coversasura "github.com/kharente-deuh/uchiyomi-server/pkg/core/covers/source/asurascans"
	httphealth "github.com/kharente-deuh/uchiyomi-server/pkg/core/health/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/library"
	pglibrary "github.com/kharente-deuh/uchiyomi-server/pkg/core/library/repository/pg"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/setup"
	httpsetup "github.com/kharente-deuh/uchiyomi-server/pkg/core/setup/gateway/http"
	httpusers "github.com/kharente-deuh/uchiyomi-server/pkg/core/users/gateway/http"
	pgusers "github.com/kharente-deuh/uchiyomi-server/pkg/core/users/repository/pg"
	httpasura "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/crypto"
)

const oidcCallbackPath = core.APIPrefix + "/auth/oidc/callback"

func setupApp(cfg *cfg) (*core.App, error) {
	logger := logging.New(logging.Config{Level: cfg.Logger.Level})

	if err := utils.PrepareDataDirs(logger, cfg.Runtime.UID, cfg.Runtime.GID, cfg.Covers.Dir, cfg.Downloads.Dir); err != nil {
		return nil, fmt.Errorf("utils.PrepareDataDirs: %w", err)
	}

	coversDir := cfg.Covers.Dir
	downloadsDir := cfg.Downloads.Dir

	dbr, err := setupDBRelated(cfg, logger)
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	services, err := setupServices(servicesDeps{
		DBr:           dbr,
		Logger:        logger,
		EncryptionKey: cfg.OIDC.EncryptionKey,
		PublicURL:     strings.TrimSuffix(cfg.OIDC.PublicURL, "/"),
	})
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	asuraApp, err := setupAsura(logger, dbr.ComicsRepository)
	if err != nil {
		return nil, fmt.Errorf("failed to init asura source: %w", err)
	}

	coversBundle, err := setupCovers(coversDeps{
		Logger:           logger,
		CoversDir:        coversDir,
		DownloadsDir:     downloadsDir,
		AsuraApp:         asuraApp,
		ComicsRepository: dbr.ComicsRepository,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init covers: %w", err)
	}

	apps, err := setupApps(appsDeps{
		Logger:                        logger,
		SessionsRepository:            dbr.SessionsRepository,
		AuthService:                   services.Auth,
		FederatedIdentitiesRepository: dbr.FederatedIdentitiesRepository,
		OIDCProvidersRepository:       dbr.OIDCProvidersRepository,
		RevalidationInterval:          cfg.OIDC.RevalidationInterval,
	})
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	chaptersSvc, downloadApp, err := setupChapters(chaptersDeps{
		Logger:             logger,
		ChaptersRepository: dbr.ChaptersRepository,
		ComicsRepository:   dbr.ComicsRepository,
		LibraryRepository:  dbr.LibraryRepository,
		Sources:            sources.SourceMap{sources.SourceAsuraScans: asuraApp},
		DownloadsDir:       downloadsDir,
		RateLimit:          cfg.Downloads.RateLimit,
		ScanInterval:       cfg.Downloads.ScanInterval,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init chapters: %w", err)
	}

	asuraApp.BindChaptersService(chaptersSvc)

	comicsSvc, err := comics.NewService(comics.Deps{
		ComicsRepository:  dbr.ComicsRepository,
		Transactor:        dbr.Txor,
		LibraryRepository: dbr.LibraryRepository,
		ChaptersService:   chaptersSvc,
		Sources:           sources.SourceMap{sources.SourceAsuraScans: asuraApp},
		Logger:            logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init comicsSvc: %w", err)
	}

	chapterListRefresh, err := comicrefresh.New(
		comicrefresh.Config{Interval: cfg.ChapterListRefreshInterval},
		comicrefresh.Deps{ComicsService: comicsSvc, Logger: logger},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init chapter list refresh: %w", err)
	}

	apps.Asura = asuraApp
	apps.Covers = coversBundle.App

	registry := core.NewHealthRegistry(dbr.PGDB)

	ctrls, err := setupCtrls(ctrlsDeps{
		Logger:               logger,
		ComicsService:        comicsSvc,
		ChaptersService:      chaptersSvc,
		SessionsService:      services.Sessions,
		SetupService:         services.Setup,
		AuthService:          services.Auth,
		OIDCProvidersService: services.OIDCProviders,
		AsuraApp:             asuraApp,
		CoversService:        coversBundle.Service,
		Registry:             registry,
	})
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	app, err := core.New(
		core.Config{
			ServerPort:     cfg.Http.Port,
			AllowedOrigins: cfg.Http.AllowedOrigins,
		},
		core.Deps{
			SetupCtrl:         ctrls.Setup,
			AsuraCtrl:         ctrls.Asura,
			CoversCtrl:        ctrls.Covers,
			HealthCtrl:        ctrls.Health,
			AuthCtrl:          ctrls.Auth,
			UsersCtrl:         ctrls.Users,
			OIDCProvidersCtrl: ctrls.OIDCProviders,
			ComicsCtrl:        ctrls.Comics,
			ChaptersCtrl:      ctrls.Chapters,

			Health:             registry,
			Logger:             logger,
			DB:                 dbr.PGDB,
			Asura:              apps.Asura,
			Covers:             apps.Covers,
			Downloads:          downloadApp,
			ChapterListRefresh: chapterListRefresh,
			Sessions:           apps.Sessions,
			OIDCRevalidation:   apps.OIDCRevalidation,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to init main app: %w", err)
	}

	return app, nil
}

type dbRelated struct {
	PGDB                          *database.PGDB
	Txor                          *pgtx.PGTransactor
	UsersRepository               *pgusers.PGUsersRepository
	SessionsRepository            *pgsessions.PGSessionsRepository
	PwdRepository                 *pgpwd.PGPasswordCredsRepository
	OIDCProvidersRepository       *pgoidcproviders.PGOIDCProvidersRepository
	FederatedIdentitiesRepository *pgfederatedidentities.PGFederatedIdentitiesRepository
	ComicsRepository              *pgcomics.PGComicsRepository
	ChaptersRepository            *pgchapters.PGChaptersRepository
	LibraryRepository             *pglibrary.PGLibraryRepository
}

func setupDBRelated(c *cfg, logger *slog.Logger) (*dbRelated, error) {
	pgdb, err := database.NewPGDatabase(
		database.PGConfig{
			Host:        c.PG.Host,
			Username:    c.PG.Username,
			Password:    c.PG.Password,
			Database:    c.PG.Database,
			Schema:      c.PG.Schema,
			SSLRequired: c.PG.SSLRequired,
			Port:        c.PG.Port,
		},
		database.PGDeps{Logger: logger},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	txor, err := pgtx.New(pgtx.Deps{DB: pgdb.DB})
	if err != nil {
		return nil, fmt.Errorf("failed to init transactor: %w", err)
	}

	usersRepository, err := pgusers.New(pgusers.Deps{DB: pgdb.DB})
	if err != nil {
		return nil, fmt.Errorf("failed to init usersRepository: %w", err)
	}

	pwdRepository, err := pgpwd.New(pgpwd.Deps{DB: pgdb.DB})
	if err != nil {
		return nil, fmt.Errorf("failed to init pwdRepository: %w", err)
	}

	sessionsRepository, err := pgsessions.New(pgsessions.Deps{DB: pgdb.DB})
	if err != nil {
		return nil, fmt.Errorf("failed to init pwdRepository: %w", err)
	}

	oidcProvidersRepository, err := pgoidcproviders.New(pgoidcproviders.Deps{DB: pgdb.DB})
	if err != nil {
		return nil, fmt.Errorf("failed to init oidcProvidersRepository: %w", err)
	}

	federatedIdentitiesRepository, err := pgfederatedidentities.New(pgfederatedidentities.Deps{DB: pgdb.DB})
	if err != nil {
		return nil, fmt.Errorf("failed to init federatedIdentitiesRepository: %w", err)
	}

	comicsRepository, err := pgcomics.New(pgcomics.Deps{DB: pgdb.DB})
	if err != nil {
		return nil, fmt.Errorf("failed to init comicsRepository: %w", err)
	}

	chaptersRepository, err := pgchapters.New(pgchapters.Deps{DB: pgdb.DB})
	if err != nil {
		return nil, fmt.Errorf("failed to init chaptersRepository: %w", err)
	}

	libraryRepository, err := pglibrary.New(pglibrary.Deps{DB: pgdb.DB})
	if err != nil {
		return nil, fmt.Errorf("failed to init libraryRepository: %w", err)
	}

	dbr := &dbRelated{
		PGDB:                          pgdb,
		Txor:                          txor,
		UsersRepository:               usersRepository,
		SessionsRepository:            sessionsRepository,
		PwdRepository:                 pwdRepository,
		OIDCProvidersRepository:       oidcProvidersRepository,
		FederatedIdentitiesRepository: federatedIdentitiesRepository,
		ComicsRepository:              comicsRepository,
		ChaptersRepository:            chaptersRepository,
		LibraryRepository:             libraryRepository,
	}

	return dbr, nil
}

type services struct {
	Auth          *auth.Service
	Setup         *setup.Service
	Sessions      *sessions.Service
	OIDCProviders *oidcproviders.Service
}

type servicesDeps struct {
	Logger        *slog.Logger
	DBr           *dbRelated
	PublicURL     string
	EncryptionKey []byte
}

func setupServices(deps servicesDeps) (*services, error) {
	sessionsSvc, err := sessions.NewService(
		sessions.ServiceConfig{
			Password: sessions.TTL{
				Idle:     30 * 24 * time.Hour,
				Absolute: 90 * 24 * time.Hour,
			},
			OIDC: sessions.TTL{
				Idle:     24 * time.Hour,
				Absolute: 30 * 24 * time.Hour,
			},
			RenewThreshold: time.Hour,
		},
		sessions.ServiceDeps{Repository: deps.DBr.SessionsRepository},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init sessionsSvc: %w", err)
	}

	hashSvc, err := bcrypthash.New(bcrypthash.Config{Cost: 8})
	if err != nil {
		return nil, fmt.Errorf("failed to init hashSvc: %w", err)
	}

	cipher, err := crypto.New(crypto.Config{Key: deps.EncryptionKey})
	if err != nil {
		return nil, fmt.Errorf("failed to init cipher: %w", err)
	}

	discoverer, err := oidc.New(
		oidc.Config{Timeout: 10 * time.Second},
		oidc.Deps{HTTPClient: &http.Client{Timeout: 10 * time.Second}},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init discoverer: %w", err)
	}

	oidcClient, err := oidc.NewClient(
		oidc.ClientConfig{Timeout: 10 * time.Second, CacheTTL: 15 * time.Minute},
		oidc.ClientDeps{HTTPClient: &http.Client{Timeout: 10 * time.Second}, Cipher: cipher},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init oidcClient: %w", err)
	}

	redirectURI := deps.PublicURL + oidcCallbackPath

	authSvc, err := auth.New(
		auth.Config{RedirectURI: redirectURI, PublicURL: deps.PublicURL, StateCookieTTL: 10 * time.Minute},
		auth.Deps{
			Transactor:                    deps.DBr.Txor,
			UsersRepository:               deps.DBr.UsersRepository,
			PwdRepository:                 deps.DBr.PwdRepository,
			HashService:                   hashSvc,
			SessionService:                sessionsSvc,
			OIDCProvidersRepository:       deps.DBr.OIDCProvidersRepository,
			FederatedIdentitiesRepository: deps.DBr.FederatedIdentitiesRepository,
			OIDCClient:                    oidcClient,
			StateCipher:                   cipher,
			Logger:                        deps.Logger,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init authSvc: %w", err)
	}

	setupSvc, err := setup.New(setup.Deps{
		AuthService:     authSvc,
		UsersRepository: deps.DBr.UsersRepository,
		Transactor:      deps.DBr.Txor,
		SessionService:  sessionsSvc,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init setupSvc: %w", err)
	}

	oidcProvidersSvc, err := oidcproviders.NewService(
		oidcproviders.ServiceConfig{RedirectURI: redirectURI},
		oidcproviders.ServiceDeps{
			Repository: deps.DBr.OIDCProvidersRepository,
			Cipher:     cipher,
			Discoverer: discoverer,
			Cache:      oidcClient,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init oidcProvidersSvc: %w", err)
	}

	svcs := &services{
		Auth:          authSvc,
		Sessions:      sessionsSvc,
		Setup:         setupSvc,
		OIDCProviders: oidcProvidersSvc,
	}

	return svcs, nil
}

type apps struct {
	Asura            *asura.App
	Covers           *covers.App
	Sessions         *sessions.App
	OIDCRevalidation *oidc.RevalidationApp
}

type appsDeps struct {
	Logger                        *slog.Logger
	SessionsRepository            sessions.SessionsRepository
	AuthService                   *auth.Service
	FederatedIdentitiesRepository federatedidentities.FederatedIdentitiesRepository
	OIDCProvidersRepository       oidcproviders.OIDCProvidersRepository
	RevalidationInterval          time.Duration
}

func setupApps(deps appsDeps) (*apps, error) {
	sessionApp, err := sessions.New(
		sessions.Config{RemoveExpiredSessionsInterval: 12 * time.Hour},
		sessions.Deps{
			Logger:             deps.Logger,
			SessionsRepository: deps.SessionsRepository,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to init sessions app: %w", err)
	}

	oidcRevalidation, err := oidc.NewRevalidationApp(
		oidc.RevalidationConfig{Interval: deps.RevalidationInterval},
		oidc.RevalidationDeps{
			Logger:                        deps.Logger,
			Revalidator:                   deps.AuthService,
			FederatedIdentitiesRepository: deps.FederatedIdentitiesRepository,
			OIDCProvidersRepository:       deps.OIDCProvidersRepository,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init oidc revalidation app: %w", err)
	}

	a := &apps{
		Sessions:         sessionApp,
		OIDCRevalidation: oidcRevalidation,
	}

	return a, nil
}

type coversBundle struct {
	App     *covers.App
	Service *covers.Service
}

type coversDeps struct {
	Logger           *slog.Logger
	AsuraApp         *asura.App
	ComicsRepository comics.ComicsRepository

	CoversDir    string
	DownloadsDir string
}

type comicsCoverFinder struct {
	repo comics.ComicsRepository
}

func (f comicsCoverFinder) FindBySourceSlug(ctx context.Context, source, slug string) (uuid.UUID, error) {
	name, err := sources.ParseSourceName(source)
	if err != nil {
		return uuid.Nil, fmt.Errorf("sources.ParseSourceName: %w", err)
	}

	comic, err := f.repo.FindBySourceSlug(ctx, comics.FindBySourceSlugOpts{Source: name, Slug: slug})
	if err != nil {
		return uuid.Nil, fmt.Errorf("repo.FindBySourceSlug: %w", err)
	}

	return comic.ID, nil
}

func setupCovers(deps coversDeps) (*coversBundle, error) {
	fetchTimeout := time.Minute

	httpClient := &http.Client{
		Timeout: fetchTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	asuraResolver, err := coversasura.New(
		coversasura.Config{CDNBaseURL: coversasura.DefaultCDNBaseURL},
		coversasura.Deps{
			Getter:     deps.AsuraApp,
			HTTPClient: httpClient,
			Logger:     deps.Logger,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("coversasura.New: %w", err)
	}

	resolvers := map[string]covers.CoverResolver{
		covers.SourceAsuraScans: asuraResolver,
	}

	cache, err := imgcache.New(imgcache.Config{
		Dir:           deps.CoversDir,
		FetchFn:       covers.NewFetchFn(resolvers),
		ErrorCacheTTL: time.Minute,
		MinInterval:   500 * time.Millisecond,
		Logger:        deps.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("imgcache.New: %w", err)
	}

	svc, err := covers.NewService(
		covers.ServiceConfig{
			ProxyPathPrefix: core.APIPrefix + "/sources/cover",
			DownloadsDir:    deps.DownloadsDir,
		},
		covers.ServiceDeps{
			Cache:      cache,
			Resolvers:  resolvers,
			HTTPClient: httpClient,
			Logger:     deps.Logger,
			Finder:     comicsCoverFinder{repo: deps.ComicsRepository},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("covers.NewService: %w", err)
	}

	app, err := covers.NewApp(covers.AppDeps{Cache: cache})
	if err != nil {
		return nil, fmt.Errorf("covers.NewApp: %w", err)
	}

	return &coversBundle{App: app, Service: svc}, nil
}

func setupAsura(logger *slog.Logger, comicsRepo comics.ComicsRepository) (*asura.App, error) {
	fetchTimeout := time.Minute

	httpClient := &http.Client{
		Timeout: fetchTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	apiClient, err := asuraclient.New(
		asuraclient.Deps{Http: httpClient, Logger: logger},
		asuraclient.Config{AsuraURL: "https://api.asurascans.com/api"},
	)

	if err != nil {
		return nil, fmt.Errorf("asuraclient.New: %w", err)
	}

	errorTTL := 30 * time.Second
	searchTTL := 5 * time.Minute
	getInfosBySlugTTL := 15 * time.Minute
	getChaptersListBySeriesTTL := 10 * time.Minute
	getImageURLsByChapterTTL := 2 * time.Hour

	searchCache, err := fncache.New(
		fncache.Config[asuradomain.SearchCacheOpts, asuradomain.SearchCacheResult]{
			Fn:            apiClient.Search,
			TTL:           searchTTL,
			ErrorTTL:      errorTTL,
			FetchTimeout:  fetchTimeout,
			CleanInterval: searchTTL,
			MaxEntries:    256,
			Name:          "asurascans.search",
			Key: func(opts asuradomain.SearchCacheOpts) string {
				return fmt.Sprintf(
					"%s %s %s %s %s %s %v %d %d %d",
					opts.Search,
					opts.Sort,
					opts.SortOrder,
					opts.Status,
					opts.Type,
					opts.Artist,
					opts.Genres,
					opts.Offset,
					opts.Limit,
					opts.MinChapters,
				)
			},
		},
		fncache.Deps{Logger: logger},
	)
	if err != nil {
		return nil, fmt.Errorf("fncache.New (asura.search): %w", err)
	}

	getInfosBySlugCache, err := fncache.New(
		fncache.Config[string, asuradomain.GetInfosBySlugResponse]{
			Fn:            apiClient.GetInfosBySlug,
			TTL:           getInfosBySlugTTL,
			ErrorTTL:      errorTTL,
			FetchTimeout:  fetchTimeout,
			CleanInterval: getInfosBySlugTTL,
			MaxEntries:    512,
			Name:          "asurascans.getInfosBySlug",
			Key: func(slug string) string {
				return slug
			},
		},
		fncache.Deps{Logger: logger},
	)
	if err != nil {
		return nil, fmt.Errorf("fncache.New (asura.getInfosBySlug): %w", err)
	}

	getChaptersListBySerieCache, err := fncache.New(
		fncache.Config[string, []asuradomain.Chapter]{
			Fn:            apiClient.GetChaptersListBySerie,
			TTL:           getChaptersListBySeriesTTL,
			ErrorTTL:      errorTTL,
			FetchTimeout:  fetchTimeout,
			CleanInterval: getChaptersListBySeriesTTL,
			MaxEntries:    512,
			Name:          "asurascans.getChaptersListBySerie",
			Key: func(slug string) string {
				return slug
			},
		},
		fncache.Deps{Logger: logger},
	)
	if err != nil {
		return nil, fmt.Errorf("fncache.New (asura.GetChaptersListBySerie): %w", err)
	}

	getImageURLsByChapterCache, err := fncache.New(
		fncache.Config[asuradomain.GetImageURLsByChapterOpts, []string]{
			Fn:            apiClient.GetImageURLsByChapter,
			TTL:           getImageURLsByChapterTTL,
			ErrorTTL:      errorTTL,
			FetchTimeout:  fetchTimeout,
			CleanInterval: getImageURLsByChapterTTL,
			MaxEntries:    1024,
			Name:          "asurascans.GetImageURLsByChapter",
			Key: func(opts asuradomain.GetImageURLsByChapterOpts) string {
				return fmt.Sprintf("%s %s", opts.SeriesSlug, opts.ChapterID)
			},
		},
		fncache.Deps{Logger: logger},
	)
	if err != nil {
		return nil, fmt.Errorf("fncache.New (asura.GetChaptersListBySerie): %w", err)
	}

	a, err := asura.New(
		asura.Config{SourceName: sources.SourceAsuraScans},
		asura.Deps{
			Logger:                       logger,
			SearchCache:                  searchCache,
			GetInfosBySlugCache:          getInfosBySlugCache,
			GetChaptersListBySeriesCache: getChaptersListBySerieCache,
			GetImageURLsByChapter:        getImageURLsByChapterCache,
			ComicsRepository:             comicsRepo,
		})
	if err != nil {
		return nil, fmt.Errorf("asura.New: %w", err)
	}

	return a, nil
}

type ctrls struct {
	Setup         *httpsetup.Controller
	Asura         *httpasura.Controller
	Covers        *httpcovers.Controller
	Health        *httphealth.Controller
	Auth          *httpauth.Controller
	Users         *httpusers.Controller
	OIDCProviders *httpoidcproviders.Controller
	Comics        *httpcomics.Controller
	Chapters      *httpchapters.Controller
}

type ctrlsDeps struct {
	AsuraApp             *asura.App
	CoversService        *covers.Service
	SetupService         *setup.Service
	SessionsService      *sessions.Service
	Logger               *slog.Logger
	Registry             *health.Registry
	AuthService          *auth.Service
	OIDCProvidersService *oidcproviders.Service
	ComicsService        *comics.Service
	ChaptersService      *chapters.Service
}

func setupCtrls(deps ctrlsDeps) (*ctrls, error) {
	cookiesMgr, err := httpsession.NewCookieManager(httpsession.CookieConfig{
		Name:   "uchiyomi_session",
		Path:   "/",
		Secure: false, //TODO: update that at release
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init cookiesManager: %w", err)
	}

	oidcStateCookiesMgr, err := httpsession.NewCookieManager(httpsession.CookieConfig{
		Name:   "uchiyomi_oidc_state",
		Path:   "/api/auth/oidc",
		Secure: false, //TODO: update that at release
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init oidcStateCookiesManager: %w", err)
	}

	authenticator, err := httpsession.NewAuthenticator(httpsession.AuthenticatorDeps{
		SessionService: deps.SessionsService,
		Cookies:        cookiesMgr,
		Logger:         deps.Logger,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to init authenticator: %w", err)
	}

	asuraCtrl, err := httpasura.New(httpasura.Config{
		Endpoint:    "/asura",
		Middlewares: chi.Middlewares{authenticator.Middleware},
	},
		httpasura.Deps{
			AsuraApp: deps.AsuraApp,
			Logger:   deps.Logger,
			CoverURLBuilder: func(source, slug string) string {
				return deps.CoversService.BuildProxyURL(source, slug)
			},
		})

	if err != nil {
		return nil, fmt.Errorf("failed to init asura controller: %w", err)
	}

	coversCtrl, err := httpcovers.New(httpcovers.Config{
		Endpoint:    "/cover",
		Middlewares: chi.Middlewares{authenticator.Middleware},
	}, httpcovers.Deps{
		Service: deps.CoversService,
		Logger:  deps.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init covers controller: %w", err)
	}

	setup, err := httpsetup.New(httpsetup.Config{Endpoint: "/setup"}, httpsetup.Deps{
		Logger:       deps.Logger,
		SetupService: deps.SetupService,
		Cookies:      cookiesMgr,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to init setup controller: %w", err)
	}

	healthCtrl, err := httphealth.New(
		httphealth.Config{ProbeTimeout: httphealth.DefaultProbeTimeout},
		httphealth.Deps{Registry: deps.Registry, Logger: deps.Logger},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init health controller: %w", err)
	}

	authCtrl, err := httpauth.New(
		httpauth.Config{
			Endpoint:          "/auth",
			LogoutMiddlewares: chi.Middlewares{authenticator.RequireSession},
		},
		httpauth.Deps{
			AuthService:      deps.AuthService,
			Cookies:          cookiesMgr,
			OIDCStateCookies: oidcStateCookiesMgr,
			ProvidersLister:  deps.OIDCProvidersService,
			Logger:           deps.Logger,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init auth controller: %w", err)
	}

	usersCtrl, err := httpusers.New(
		httpusers.Config{
			Endpoint:    "/users",
			Middlewares: chi.Middlewares{authenticator.Middleware},
		},
		httpusers.Deps{
			Logger: deps.Logger,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init users controller: %w", err)
	}

	oidcProvidersCtrl, err := httpoidcproviders.New(
		httpoidcproviders.Config{
			Endpoint:    "/oidc/providers",
			Middlewares: chi.Middlewares{authenticator.Middleware, authenticator.RequireAdmin},
		},
		httpoidcproviders.Deps{
			Logger:  deps.Logger,
			Service: deps.OIDCProvidersService,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init oidc providers controller: %w", err)
	}

	comicsCtrl, err := httpcomics.New(
		httpcomics.Config{
			Endpoint:    "/comics",
			Middlewares: chi.Middlewares{authenticator.Middleware, authenticator.RequireAdmin},
		},
		httpcomics.Deps{
			Logger:        deps.Logger,
			ComicsService: deps.ComicsService,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to init comics controller: %w", err)
	}

	chaptersCtrl, err := httpchapters.New(
		httpchapters.Config{
			Endpoint:    "/chapters",
			Middlewares: chi.Middlewares{authenticator.Middleware},
		},
		httpchapters.Deps{
			Logger:          deps.Logger,
			ChaptersService: deps.ChaptersService,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init chapters controller: %w", err)
	}

	c := &ctrls{
		Asura:         asuraCtrl,
		Covers:        coversCtrl,
		Setup:         setup,
		Health:        healthCtrl,
		Auth:          authCtrl,
		Users:         usersCtrl,
		OIDCProviders: oidcProvidersCtrl,
		Comics:        comicsCtrl,
		Chapters:      chaptersCtrl,
	}

	return c, nil
}

type chaptersDeps struct {
	Logger             *slog.Logger
	ChaptersRepository chapters.ChaptersRepository
	ComicsRepository   comics.ComicsRepository
	LibraryRepository  library.LibraryRepository
	Sources            sources.SourceMap
	DownloadsDir       string
	RateLimit          time.Duration
	ScanInterval       time.Duration
}

func setupChapters(deps chaptersDeps) (*chapters.Service, *download.App, error) {
	fetchTimeout := time.Minute

	httpClient := &http.Client{
		Timeout: fetchTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	worker, err := download.New(
		download.Config{
			Dir:       deps.DownloadsDir,
			RateLimit: deps.RateLimit,
		},
		download.Deps{
			ChaptersRepository: deps.ChaptersRepository,
			ComicsRepository:   deps.ComicsRepository,
			Sources:            deps.Sources,
			HTTPClient:         httpClient,
			Logger:             deps.Logger,
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("download.New: %w", err)
	}

	svc, err := chapters.NewService(chapters.Deps{
		Repository:        deps.ChaptersRepository,
		ChapterDownloader: worker,
		LibraryRepository: deps.LibraryRepository,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("chapters.NewService: %w", err)
	}

	app, err := download.NewApp(
		download.AppConfig{ScanInterval: deps.ScanInterval},
		download.AppDeps{
			Worker:          worker,
			ChaptersService: svc,
			Logger:          deps.Logger,
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("download.NewApp: %w", err)
	}

	return svc, app, nil
}
