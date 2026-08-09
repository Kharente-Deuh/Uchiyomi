// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	core "github.com/kharente-deuh/uchiyomi-server/pkg/core"
	asura "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/core"
	asuradomain "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
	asuraclient "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/transport/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/database"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/fncache"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/health"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/logging"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth"
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
	httphealth "github.com/kharente-deuh/uchiyomi-server/pkg/core/health/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/setup"
	httpsetup "github.com/kharente-deuh/uchiyomi-server/pkg/core/setup/gateway/http"
	httpusers "github.com/kharente-deuh/uchiyomi-server/pkg/core/users/gateway/http"
	pgusers "github.com/kharente-deuh/uchiyomi-server/pkg/core/users/repository/pg"
	httpasura "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/crypto"
)

const oidcCallbackPath = "/api/auth/oidc/callback"

func setupApp(cfg *cfg) (*core.App, error) {
	logger := logging.New(logging.Config{Level: cfg.Logger.Level})

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

	apps, err := setupApps(appsDeps{
		Logger:             logger,
		SessionsRepository: dbr.SessionsRepository,
	})
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	registry := core.NewHealthRegistry(dbr.PGDB)

	ctrls, err := setupCtrls(ctrlsDeps{
		Logger:               logger,
		SessionsService:      services.Sessions,
		SetupService:         services.Setup,
		AuthService:          services.Auth,
		OIDCProvidersService: services.OIDCProviders,
		AsuraApp:             apps.Asura,
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
			HealthCtrl:        ctrls.Health,
			AuthCtrl:          ctrls.Auth,
			UsersCtrl:         ctrls.Users,
			OIDCProvidersCtrl: ctrls.OIDCProviders,

			Health:   registry,
			Logger:   logger,
			DB:       dbr.PGDB,
			Asura:    apps.Asura,
			Sessions: apps.Sessions,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to init main app: %w", err)
	}

	return app, nil
}

type dbRelated struct {
	PGDB                    *database.PGDB
	Txor                    *pgtx.PGTransactor
	UsersRepository         *pgusers.PGUsersRepository
	SessionsRepository      *pgsessions.PGSessionsRepository
	PwdRepository           *pgpwd.PGPasswordCredsRepository
	OIDCProvidersRepository *pgoidcproviders.PGOIDCProvidersRepository
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

	dbr := &dbRelated{
		PGDB:                    pgdb,
		Txor:                    txor,
		UsersRepository:         usersRepository,
		SessionsRepository:      sessionsRepository,
		PwdRepository:           pwdRepository,
		OIDCProvidersRepository: oidcProvidersRepository,
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

	authSvc, err := auth.New(auth.Deps{
		Transactor:      deps.DBr.Txor,
		UsersRepository: deps.DBr.UsersRepository,
		PwdRepository:   deps.DBr.PwdRepository,
		HashService:     hashSvc,
		SessionService:  sessionsSvc,
	})
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

	oidcProvidersSvc, err := oidcproviders.NewService(
		oidcproviders.ServiceConfig{RedirectURI: deps.PublicURL + oidcCallbackPath},
		oidcproviders.ServiceDeps{
			Repository: deps.DBr.OIDCProvidersRepository,
			Cipher:     cipher,
			Discoverer: discoverer,
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
	Asura    *asura.App
	Sessions *sessions.App
}

type appsDeps struct {
	Logger             *slog.Logger
	SessionsRepository sessions.SessionsRepository
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

	asura, err := setupAsura(deps.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init asura source: %w", err)
	}

	a := &apps{
		Asura:    asura,
		Sessions: sessionApp,
	}

	return a, nil
}

func setupAsura(logger *slog.Logger) (*asura.App, error) {
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
		asuraclient.Deps{Http: httpClient},
		asuraclient.Config{AsuraURL: "https://api.asurascans.com"},
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
		fncache.Config[asuradomain.SearchOpts, asuradomain.SearchResult]{
			Fn:            apiClient.Search,
			TTL:           searchTTL,
			ErrorTTL:      errorTTL,
			FetchTimeout:  fetchTimeout,
			CleanInterval: searchTTL,
			MaxEntries:    256,
			Name:          "asurascans.search",
			Key: func(opts asuradomain.SearchOpts) string {
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

	a, err := asura.New(asura.Dependencies{
		Logger:                       logger,
		SearchCache:                  searchCache,
		GetInfosBySlugCache:          getInfosBySlugCache,
		GetChaptersListBySeriesCache: getChaptersListBySerieCache,
		GetImageURLsByChapter:        getImageURLsByChapterCache,
	})
	if err != nil {
		return nil, fmt.Errorf("asura.New: %w", err)
	}

	return a, nil
}

type ctrls struct {
	Setup         *httpsetup.Controller
	Asura         *httpasura.Controller
	Health        *httphealth.Controller
	Auth          *httpauth.Controller
	Users         *httpusers.Controller
	OIDCProviders *httpoidcproviders.Controller
}

type ctrlsDeps struct {
	AsuraApp             *asura.App
	SetupService         *setup.Service
	SessionsService      *sessions.Service
	Logger               *slog.Logger
	Registry             *health.Registry
	AuthService          *auth.Service
	OIDCProvidersService *oidcproviders.Service
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

	authenticator, err := httpsession.NewAuthenticator(httpsession.AuthenticatorDeps{
		SessionService: deps.SessionsService,
		Cookies:        cookiesMgr,
		Logger:         deps.Logger,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to init authenticator: %w", err)
	}

	asura, err := httpasura.New(httpasura.Config{
		Endpoint:    "/asura",
		Middlewares: chi.Middlewares{authenticator.Middleware},
	},
		httpasura.Deps{
			AsuraApp: deps.AsuraApp,
			Logger:   deps.Logger,
		})

	if err != nil {
		return nil, fmt.Errorf("failed to init asura controller: %w", err)
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
		httpauth.Config{Endpoint: "/auth"},
		httpauth.Deps{
			AuthService: deps.AuthService,
			Cookies:     cookiesMgr,
			Logger:      deps.Logger,
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

	c := &ctrls{
		Asura:         asura,
		Setup:         setup,
		Health:        healthCtrl,
		Auth:          authCtrl,
		Users:         usersCtrl,
		OIDCProviders: oidcProvidersCtrl,
	}

	return c, nil
}
