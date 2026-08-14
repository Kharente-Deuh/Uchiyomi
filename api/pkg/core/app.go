// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpauth "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/gateway/http"
	httpoidcproviders "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	httpcomics "github.com/kharente-deuh/uchiyomi-server/pkg/core/comics/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
	httpcovers "github.com/kharente-deuh/uchiyomi-server/pkg/core/covers/gateway/http"
	httphealth "github.com/kharente-deuh/uchiyomi-server/pkg/core/health/gateway/http"
	httpsetup "github.com/kharente-deuh/uchiyomi-server/pkg/core/setup/gateway/http"
	httpusers "github.com/kharente-deuh/uchiyomi-server/pkg/core/users/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/health"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/logging"
	"github.com/kharente-deuh/uchiyomi-server/pkg/webui"
	"golang.org/x/sync/errgroup"

	asura "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/core"
	httpasura "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/gateway/http"
)

const (
	portMaxValue    = 65535
	shutdownTimeout = 5 * time.Second

	APIPrefix = "/api"
	apiPrefix = APIPrefix
)

type Config struct {
	AllowedOrigins []string
	ServerPort     int
}

func (cfg *Config) Validate() error {
	if cfg.ServerPort <= 0 {
		return errors.New("serverPort must be greater than 0")
	}

	if cfg.ServerPort > portMaxValue {
		return fmt.Errorf("serverPort must not exceed %d", portMaxValue)
	}

	return nil
}

type Database interface {
	Migrate() error
	Ping(ctx context.Context) error
	Close()
}

type Deps struct {
	DB Database

	SetupCtrl         *httpsetup.Controller
	AsuraCtrl         *httpasura.Controller
	CoversCtrl        *httpcovers.Controller
	HealthCtrl        *httphealth.Controller
	AuthCtrl          *httpauth.Controller
	UsersCtrl         *httpusers.Controller
	ComicsCtrl        *httpcomics.Controller
	OIDCProvidersCtrl *httpoidcproviders.Controller

	Logger *slog.Logger
	Health *health.Registry

	Asura            *asura.App
	Covers           *covers.App
	Downloads        interface{ Run(context.Context) error }
	Sessions         *sessions.App
	OIDCRevalidation interface{ Run(context.Context) error }
}

func (deps *Deps) Validate() error {
	if deps.SetupCtrl == nil {
		return errors.New("setupCtrl is required")
	}

	if deps.AsuraCtrl == nil {
		return errors.New("asuraCtrl is required")
	}

	if deps.CoversCtrl == nil {
		return errors.New("coversCtrl is required")
	}

	if deps.Logger == nil {
		return errors.New("logger is required")
	}

	if deps.DB == nil {
		return errors.New("db is required")
	}

	if deps.Asura == nil {
		return errors.New("asura is required")
	}

	if deps.Covers == nil {
		return errors.New("covers is required")
	}

	if deps.Downloads == nil {
		return errors.New("downloads is required")
	}

	if deps.Sessions == nil {
		return errors.New("sessions is required")
	}

	if deps.OIDCRevalidation == nil {
		return errors.New("oidcRevalidation is required")
	}

	if deps.HealthCtrl == nil {
		return errors.New("healthCtrl is required")
	}

	if deps.AuthCtrl == nil {
		return errors.New("authCtrl is required")
	}

	if deps.UsersCtrl == nil {
		return errors.New("usersCtrl is required")
	}

	if deps.OIDCProvidersCtrl == nil {
		return errors.New("oidcProvidersCtrl is required")
	}

	if deps.ComicsCtrl == nil {
		return errors.New("comicsCtrl is required")
	}

	if deps.Health == nil {
		return errors.New("health is required")
	}

	return nil
}

type App struct {
	deps Deps
	cfg  Config
}

func New(cfg Config, deps Deps) (*App, error) {
	var err error
	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err = deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	return &App{deps: deps, cfg: cfg}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.deps.DB.Close()

	srv := a.initServer()
	errG, errCtx := errgroup.WithContext(ctx)

	errG.Go(a.runServer(srv))
	errG.Go(a.startup(errCtx, errG))
	errG.Go(func() error {
		<-errCtx.Done()
		a.stopServer(srv)

		return nil
	})

	//nolint:wrapcheck
	return errG.Wait()
}

func (a *App) startup(ctx context.Context, errG *errgroup.Group) func() error {
	return func() error {
		if err := a.deps.DB.Migrate(); err != nil {
			a.deps.Health.Set(componentMigrations, err)

			return fmt.Errorf("db.Migrate: %w", err)
		}

		a.deps.Health.Set(componentMigrations, nil)

		errG.Go(a.runComponent(ctx, componentAsura, a.deps.Asura.Run))
		errG.Go(a.runComponent(ctx, componentCovers, a.deps.Covers.Run))
		errG.Go(a.runComponent(ctx, componentDownloads, a.deps.Downloads.Run))
		errG.Go(a.runComponent(ctx, componentSessions, a.deps.Sessions.Run))
		errG.Go(a.runComponent(ctx, componentOIDCRevalidation, a.deps.OIDCRevalidation.Run))

		return nil
	}
}

func (a *App) runComponent(ctx context.Context, name string, run func(context.Context) error) func() error {
	return func() error {
		a.deps.Health.Set(name, nil)

		if err := run(ctx); err != nil {
			a.deps.Health.Set(name, err)

			return fmt.Errorf("%s.Run: %w", name, err)
		}

		return nil
	}
}

func (a *App) initServer() *http.Server {
	ui, _ := webui.Handler()
	r := a.newRouter(ui)
	a.logRoutes(r)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", a.cfg.ServerPort),
		Handler: r,
	}
}

func (a *App) logRoutes(r chi.Router) {
	walk := func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		a.deps.Logger.Info(fmt.Sprintf("[%s] %s", method, route))

		return nil
	}

	if err := chi.Walk(r, walk); err != nil {
		a.deps.Logger.Warn("failed to list routes", logging.Err(err))
	}
}

func (a *App) newRouter(ui http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Route(apiPrefix, func(r chi.Router) {
		r.NotFound(http.NotFound)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Recoverer)
			a.deps.HealthCtrl.InitRouter(r)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequestID)
			r.Use(middleware.ClientIPFromRemoteAddr)
			r.Use(a.requestLogger)
			r.Use(middleware.Recoverer)

			if len(a.cfg.AllowedOrigins) > 0 {
				r.Use(cors.Handler(cors.Options{
					AllowedOrigins:   a.cfg.AllowedOrigins,
					AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
					AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
					ExposedHeaders:   []string{"Link"},
					AllowCredentials: true,
					MaxAge:           300,
				}))
			}
			r.Use(a.requireMigrations)

			a.deps.SetupCtrl.InitRouter(r)
			a.deps.AuthCtrl.InitRouter(r)
			a.deps.UsersCtrl.InitRouter(r)
			a.deps.OIDCProvidersCtrl.InitRouter(r)
			a.deps.ComicsCtrl.InitRouter(r)
			r.Route("/sources", func(r chi.Router) {
				a.deps.CoversCtrl.InitRouter(r)
				a.deps.AsuraCtrl.InitRouter(r)
			})
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Recoverer)
		a.deps.HealthCtrl.InitRouter(r)
	})

	if ui != nil {
		r.NotFound(ui.ServeHTTP)
	}

	return r
}

func (a *App) runServer(srv *http.Server) func() error {
	return func() error {
		a.deps.Logger.Info("HTTP Server listening", "port", a.cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("srv.ListenAndServe: %w", err)
		}

		return nil
	}
}

func (a *App) stopServer(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	a.deps.Logger.Info("shutting down HTTP server")

	if err := srv.Shutdown(ctx); err != nil {
		a.deps.Logger.ErrorContext(ctx, "failed to shutdown http server", "error", err)
	}
}
