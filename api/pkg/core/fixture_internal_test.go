// SPDX-License-Identifier: AGPL-3.0-or-later

package core

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	sessionshttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/gateway/http"
	healthhttp "github.com/kharente-deuh/uchiyomi-server/pkg/core/health/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/setup"
	httpsetup "github.com/kharente-deuh/uchiyomi-server/pkg/core/setup/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	httpusers "github.com/kharente-deuh/uchiyomi-server/pkg/core/users/gateway/http"
	asura "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/core"
	asuradomain "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/domain"
	httasura "github.com/kharente-deuh/uchiyomi-server/pkg/sources/asurascans/pkg/gateway/http"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/fncache"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/health"
)

const (
	runDeadline = 10 * time.Second

	pollInterval = 2 * time.Millisecond

	notImplemented = "non implémenté"
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

func (fakeAuthService) LoginWithPwd(context.Context, auth.LoginWithPwdOpts) (*sessions.IssuedSession, error) {
	return nil, errors.New(notImplemented)
}

func (fakeAuthService) CreateUserWithPwd(context.Context, auth.CreateUserWithPwdOpts) (*users.User, error) {
	return nil, errors.New(notImplemented)
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

func newTestAsura(t *testing.T, logger *slog.Logger) *asura.App {
	t.Helper()

	app, err := asura.New(asura.Dependencies{
		Logger:      logger,
		SearchCache: newTestCache[asuradomain.SearchOpts, asuradomain.SearchResult](t, "search", logger),
		GetInfosBySlugCache: newTestCache[string, asuradomain.GetInfosBySlugResponse](
			t, "infos", logger,
		),
		GetChaptersListBySeriesCache: newTestCache[string, []asuradomain.Chapter](
			t, "chapters", logger,
		),
		GetImageURLsByChapter: newTestCache[asuradomain.GetImageURLsByChapterOpts, []string](
			t, "images", logger,
		),
	})
	if err != nil {
		t.Fatalf("asura.New: %v", err)
	}

	return app
}

func newTestApp(t *testing.T, db Database, port int) (*App, *health.Registry) {
	t.Helper()

	logger := slog.New(slog.DiscardHandler)
	registry := NewHealthRegistry(db)
	asuraApp := newTestAsura(t, logger)

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

	setupCtrl, err := httpsetup.New(
		httpsetup.Config{Endpoint: "/setup"},
		httpsetup.Deps{Logger: logger, SetupService: fakeSetupService{}, Cookies: cookies},
	)
	if err != nil {
		t.Fatalf("httpsetup.New: %v", err)
	}

	authCtrl, err := httpauth.New(
		httpauth.Config{Endpoint: "/auth"},
		httpauth.Deps{Logger: logger, AuthService: fakeAuthService{}, Cookies: cookies},
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

	asuraCtrl, err := httasura.New(
		httasura.Config{Endpoint: "/asura"},
		httasura.Deps{Logger: logger, AsuraApp: asuraApp},
	)
	if err != nil {
		t.Fatalf("httasura.New: %v", err)
	}

	healthCtrl, err := healthhttp.New(
		healthhttp.Config{ProbeTimeout: healthhttp.DefaultProbeTimeout},
		healthhttp.Deps{Registry: registry, Logger: logger},
	)
	if err != nil {
		t.Fatalf("healthhttp.New: %v", err)
	}

	app, err := New(
		Config{ServerPort: port, AllowedOrigins: []string{"*"}},
		Deps{
			DB:         db,
			SetupCtrl:  setupCtrl,
			AuthCtrl:   authCtrl,
			UsersCtrl:  usersCtrl,
			AsuraCtrl:  asuraCtrl,
			HealthCtrl: healthCtrl,
			Logger:     logger,
			Health:     registry,
			Asura:      asuraApp,
			Sessions:   sessionsApp,
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
		t.Fatal("Run n'a pas rendu la main après annulation du contexte")

		return nil
	}
}

func (r *runningApp) waitListening(t *testing.T) {
	t.Helper()

	r.waitFor(t, "le serveur n'écoute pas", func() bool {
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
		t.Fatalf("décodage de %q: %v", raw, err)
	}

	return body
}
