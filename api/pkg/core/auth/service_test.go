// SPDX-License-Identifier: AGPL-3.0-or-later

package auth_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/hash"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/password"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidc"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/crypto"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
)

const (
	userName         = "alice"
	pwd              = "hunter2hunter2"
	token            = "letoken"
	publicURL        = "https://app.example.com"
	oidcRedirectURI  = "https://app.example.com/callback"
	testCipherKey    = "abcdefghijklmnopqrstuvwxyz012345"
	usernameClaimKey = "preferred_username"
	testSubject      = "sub-123"
)

type txCtxKey struct{}

type clock struct{ n int }

func (c *clock) tick() int {
	c.n++

	return c.n
}

type fakeTransactor struct {
	clock     *clock
	commitErr error
	opts      transaction.TxOpts
	beganAt   int
	calls     int
}

func (f *fakeTransactor) WithinTx(
	ctx context.Context,
	opts transaction.TxOpts,
	fn func(context.Context) error,
) error {
	f.calls++
	f.opts = opts
	f.beganAt = f.clock.tick()

	if err := fn(context.WithValue(ctx, txCtxKey{}, true)); err != nil {
		return err
	}

	return f.commitErr
}

type fakeHashService struct {
	clock      *clock
	err        error
	matchErr   error
	got        []byte
	gotHashed  []byte
	gotCompare []byte
	hashedAt   int
	hashCalls  int
	matchCalls int
	match      bool
}

func (f *fakeHashService) Hash(toHash []byte) ([]byte, error) {
	f.hashCalls++
	f.got = toHash
	f.hashedAt = f.clock.tick()

	if f.err != nil {
		return nil, f.err
	}

	return []byte("hashed:" + string(toHash)), nil
}

func (f *fakeHashService) Match(hashed []byte, toCompare []byte) (bool, error) {
	f.matchCalls++
	f.gotHashed = hashed
	f.gotCompare = toCompare

	if f.matchErr != nil {
		return false, f.matchErr
	}

	return f.match, nil
}

type fakeUsersRepository struct {
	byNameErr   error
	byIDErr     error
	updateErr   error
	countErr    error
	err         error
	user        *users.User
	byID        map[uuid.UUID]*users.User
	gotName     string
	gotOpts     users.CreateUserOpts
	nameCalls   int
	byIDCalls   int
	updateCalls int
	adminCount  int
	countCalls  int
	gotUpdate   users.UpdateUserOpts
	gotID       uuid.UUID
	inTx        bool
}

func (f *fakeUsersRepository) Create(ctx context.Context, opts users.CreateUserOpts) (*users.User, error) {
	f.gotOpts = opts
	f.inTx, _ = ctx.Value(txCtxKey{}).(bool)

	if f.err != nil {
		return nil, f.err
	}

	return f.user, nil
}

func (f *fakeUsersRepository) GetByUsername(_ context.Context, name string) (*users.User, error) {
	f.nameCalls++
	f.gotName = name

	if f.byNameErr != nil {
		return nil, f.byNameErr
	}

	return f.user, nil
}

func (f *fakeUsersRepository) CountAdmins(context.Context) (int, error) {
	f.countCalls++

	if f.countErr != nil {
		return 0, f.countErr
	}

	return f.adminCount, nil
}

func (f *fakeUsersRepository) GetByID(_ context.Context, id uuid.UUID) (*users.User, error) {
	f.byIDCalls++
	f.gotID = id

	if f.byIDErr != nil {
		return nil, f.byIDErr
	}

	if u, ok := f.byID[id]; ok {
		return u, nil
	}

	return nil, domain.ErrNotFound
}

func (f *fakeUsersRepository) Update(_ context.Context, opts users.UpdateUserOpts) (*users.User, error) {
	f.updateCalls++
	f.gotUpdate = opts

	if f.updateErr != nil {
		return nil, f.updateErr
	}

	updated := *f.user
	updated.IsAdmin = opts.IsAdmin

	return &updated, nil
}

type fakePwdRepository struct {
	err       error
	getErr    error
	creds     *password.PasswordCreds
	gotOpts   password.UpsertPasswordCredsOpts
	gotUserID uuid.UUID
	calls     int
	getCalls  int
	inTx      bool
}

func (f *fakePwdRepository) Create(
	ctx context.Context,
	opts password.UpsertPasswordCredsOpts,
) (*password.PasswordCreds, error) {
	f.calls++
	f.gotOpts = opts
	f.inTx, _ = ctx.Value(txCtxKey{}).(bool)

	if f.err != nil {
		return nil, f.err
	}

	return &password.PasswordCreds{UserID: opts.UserID, Hash: opts.Hash}, nil
}

func (f *fakePwdRepository) GetByUserID(_ context.Context, userID uuid.UUID) (*password.PasswordCreds, error) {
	f.getCalls++
	f.gotUserID = userID

	if f.getErr != nil {
		return nil, f.getErr
	}

	return f.creds, nil
}

func (f *fakePwdRepository) UpdateByUserID(context.Context, password.UpsertPasswordCredsOpts) error {
	panic("UpdateByUserID is not used by the auth service")
}

type fakeSessionService struct {
	err         error
	revokeErr   error
	issued      *sessions.IssuedSession
	gotToken    string
	gotOpts     sessions.CreateSessionOpts
	calls       int
	revokeCalls int
}

func (f *fakeSessionService) Create(
	_ context.Context,
	opts sessions.CreateSessionOpts,
) (*sessions.IssuedSession, error) {
	f.calls++
	f.gotOpts = opts

	if f.err != nil {
		return nil, f.err
	}

	return f.issued, nil
}

func (f *fakeSessionService) Authenticate(context.Context, string) (*sessions.AuthenticatedSession, error) {
	panic("Authenticate is not used by the auth service")
}

func (f *fakeSessionService) Revoke(_ context.Context, token string) error {
	f.revokeCalls++
	f.gotToken = token

	if f.revokeErr != nil {
		return f.revokeErr
	}

	return nil
}

func (f *fakeSessionService) RevokeAllForUser(context.Context, uuid.UUID) error {
	panic("RevokeAllForUser is not used by the auth service")
}

type fakes struct {
	now   time.Time
	clock *clock
	hs    *fakeHashService
	ur    *fakeUsersRepository
	pr    *fakePwdRepository
	tr    *fakeTransactor
	ss    *fakeSessionService
	opr   *fakeOIDCProvidersRepo
	fir   *fakeFederatedIdentitiesRepo
	oc    *fakeOIDCClient
	sc    *crypto.Cipher
	logs  *bytes.Buffer
	cfg   auth.Config
}

func newFakes() *fakes {
	c := &clock{}
	user := &users.User{ID: uuid.New(), Name: userName}
	provider := &oidcproviders.OIDCProvider{
		ID:            uuid.New(),
		DisplayName:   "Example IdP",
		IssuerURL:     "https://idp.example.com",
		ClientID:      "client-id",
		UsernameClaim: usernameClaimKey,
	}

	sc, err := crypto.New(crypto.Config{Key: []byte(testCipherKey)})
	if err != nil {
		panic(err)
	}

	return &fakes{
		clock: c,
		cfg:   auth.Config{PublicURL: publicURL, RedirectURI: oidcRedirectURI, StateCookieTTL: 10 * time.Minute},
		now:   time.Date(2026, time.August, 10, 12, 0, 0, 0, time.UTC),
		hs:    &fakeHashService{clock: c, match: true},
		ur:    &fakeUsersRepository{user: user, byID: map[uuid.UUID]*users.User{user.ID: user}},
		pr:    &fakePwdRepository{creds: &password.PasswordCreds{UserID: user.ID, Hash: "hashed:" + pwd}},
		tr:    &fakeTransactor{clock: c},
		ss: &fakeSessionService{issued: &sessions.IssuedSession{
			Session: sessions.Session{ID: uuid.New(), UserID: user.ID},
			Token:   token,
		}},
		opr: &fakeOIDCProvidersRepo{provider: provider},
		fir: &fakeFederatedIdentitiesRepo{getErr: domain.ErrNotFound},
		oc: &fakeOIDCClient{
			authCodeURL: "https://idp.example.com/authorize?state=letstate",
			tokenSet: &oidcproviders.TokenSet{
				Subject: testSubject,
				SID:     "sid-123",
				Claims:  map[string]any{usernameClaimKey: userName},
			},
		},
		sc:   sc,
		logs: &bytes.Buffer{},
	}
}

func (f *fakes) svc(t *testing.T) *auth.Service {
	t.Helper()

	logger := slog.New(slog.NewJSONHandler(f.logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

	svc, err := auth.New(f.cfg, auth.Deps{
		HashService:                   f.hs,
		UsersRepository:               f.ur,
		PwdRepository:                 f.pr,
		Transactor:                    f.tr,
		SessionService:                f.ss,
		OIDCProvidersRepository:       f.opr,
		FederatedIdentitiesRepository: f.fir,
		OIDCClient:                    f.oc,
		StateCipher:                   f.sc,
		Logger:                        logger,
		Now:                           func() time.Time { return f.now },
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return svc
}

func TestNewRequiresTransactor(t *testing.T) {
	t.Parallel()

	f := newFakes()

	svc, err := auth.New(
		auth.Config{PublicURL: publicURL, RedirectURI: oidcRedirectURI},
		auth.Deps{HashService: f.hs, UsersRepository: f.ur, PwdRepository: f.pr},
	)
	if err == nil {
		t.Fatal("New without transactor must fail")
	}

	if svc != nil {
		t.Error("New returned a service in addition to the error")
	}

	if want := "deps.Validate: transactor is required"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
}

func TestCreateUserWithPwdWrapsBothWritesInOneTransaction(t *testing.T) {
	t.Parallel()

	f := newFakes()

	got, err := f.svc(t).CreateUserWithPwd(context.Background(), auth.CreateUserWithPwdOpts{
		Name:     userName,
		Password: pwd,
		IsAdmin:  true,
	})
	if err != nil {
		t.Fatalf("CreateUserWithPwd: %v", err)
	}

	if got == nil || got.Name != userName {
		t.Errorf("CreateUserWithPwd() = %+v", got)
	}

	if f.tr.calls != 1 {
		t.Errorf("WithinTx called %d times, want 1", f.tr.calls)
	}

	if !f.ur.inTx {
		t.Error("UsersRepository.Create did not receive transactional ctx")
	}

	if !f.pr.inTx {
		t.Error("PwdRepository.Create did not receive transactional ctx")
	}
}

func TestCreateUserWithPwdHashesBeforeOpeningTransaction(t *testing.T) {
	t.Parallel()

	f := newFakes()

	if _, err := f.svc(t).CreateUserWithPwd(context.Background(), auth.CreateUserWithPwdOpts{
		Name:     userName,
		Password: pwd,
	}); err != nil {
		t.Fatalf("CreateUserWithPwd: %v", err)
	}

	if f.hs.hashedAt == 0 || f.tr.beganAt == 0 {
		t.Fatalf("calls not observed: hashedAt=%d beganAt=%d", f.hs.hashedAt, f.tr.beganAt)
	}

	if f.hs.hashedAt > f.tr.beganAt {
		t.Errorf("hash performed inside transaction (hashedAt=%d > beganAt=%d)", f.hs.hashedAt, f.tr.beganAt)
	}
}

func TestCreateUserWithPwdUsesDefaultIsolation(t *testing.T) {
	t.Parallel()

	f := newFakes()

	if _, err := f.svc(t).CreateUserWithPwd(context.Background(), auth.CreateUserWithPwdOpts{Name: "a"}); err != nil {
		t.Fatalf("CreateUserWithPwd: %v", err)
	}

	if f.tr.opts.Isolation != transaction.IsolationDefault {
		t.Errorf("Isolation = %v, want IsolationDefault", f.tr.opts.Isolation)
	}
}

func TestCreateUserWithPwdPassesHashToPwdRepository(t *testing.T) {
	t.Parallel()

	f := newFakes()

	if _, err := f.svc(t).CreateUserWithPwd(context.Background(), auth.CreateUserWithPwdOpts{
		Name:     userName,
		Password: pwd,
	}); err != nil {
		t.Fatalf("CreateUserWithPwd: %v", err)
	}

	if string(f.hs.got) != pwd {
		t.Errorf("Hash received %q", f.hs.got)
	}

	if f.pr.gotOpts.Hash != "hashed:"+pwd {
		t.Errorf("PwdRepository received hash %q", f.pr.gotOpts.Hash)
	}

	if f.pr.gotOpts.UserID != f.ur.user.ID {
		t.Errorf("PwdRepository received UserID %v, want %v", f.pr.gotOpts.UserID, f.ur.user.ID)
	}
}

func TestCreateUserWithPwdReturnsNoUserWhenPwdWriteFails(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.pr.err = errors.New("disk full")

	got, err := f.svc(t).CreateUserWithPwd(context.Background(), auth.CreateUserWithPwdOpts{Name: userName})
	if !errors.Is(err, f.pr.err) {
		t.Errorf("err = %v, original error no longer reachable", err)
	}

	if got != nil {
		t.Errorf("CreateUserWithPwd returned %+v although transaction failed", got)
	}
}

func TestCreateUserWithPwdReturnsNoUserWhenCommitFails(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.tr.commitErr = errors.New("commit refused")

	got, err := f.svc(t).CreateUserWithPwd(context.Background(), auth.CreateUserWithPwdOpts{Name: userName})
	if !errors.Is(err, f.tr.commitErr) {
		t.Errorf("err = %v, want commit error", err)
	}

	if got != nil {
		t.Errorf("CreateUserWithPwd returned %+v although commit failed", got)
	}
}

func TestCreateUserWithPwdPropagatesSentinels(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		setup func(*fakes)
		want  error
	}{
		"password too long": {
			setup: func(f *fakes) { f.hs.err = hash.ErrStringTooLong },
			want:  hash.ErrStringTooLong,
		},
		"name already taken": {
			setup: func(f *fakes) { f.ur.err = domain.ErrAlreadyExists },
			want:  domain.ErrAlreadyExists,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newFakes()
			tc.setup(f)

			got, err := f.svc(t).CreateUserWithPwd(context.Background(), auth.CreateUserWithPwdOpts{Name: "a"})
			if !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}

			if got != nil {
				t.Errorf("CreateUserWithPwd returned %+v in addition to the error", got)
			}
		})
	}
}

func TestCreateUserWithPwdSkipsTransactionWhenHashFails(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.hs.err = hash.ErrStringTooLong

	if _, err := f.svc(t).CreateUserWithPwd(context.Background(), auth.CreateUserWithPwdOpts{Name: "a"}); err == nil {
		t.Fatal("CreateUserWithPwd must fail")
	}

	if f.tr.calls != 0 {
		t.Errorf("WithinTx called %d times although hash failed", f.tr.calls)
	}

	if f.pr.calls != 0 {
		t.Errorf("PwdRepository.Create called %d times although hash failed", f.pr.calls)
	}
}

func TestLoginWithPwdIssuesPasswordSession(t *testing.T) {
	t.Parallel()

	f := newFakes()

	got, err := f.svc(t).LoginWithPwd(context.Background(), auth.LoginWithPwdOpts{Username: userName, Password: pwd})
	if err != nil {
		t.Fatalf("LoginWithPwd: %v", err)
	}

	if got == nil || got.Session == nil || got.Session.Token != token {
		t.Fatalf("LoginWithPwd() = %+v", got)
	}

	if got.User == nil || got.User.ID != f.ur.user.ID {
		t.Errorf("LoginWithPwd() user = %+v, want %v", got.User, f.ur.user.ID)
	}

	if f.ur.gotName != userName {
		t.Errorf("GetByUsername received %q, want %q", f.ur.gotName, userName)
	}

	if f.pr.gotUserID != f.ur.user.ID {
		t.Errorf("GetByUserID received %v, want %v", f.pr.gotUserID, f.ur.user.ID)
	}

	if string(f.hs.gotHashed) != f.pr.creds.Hash || string(f.hs.gotCompare) != pwd {
		t.Errorf("Match received (%q, %q)", f.hs.gotHashed, f.hs.gotCompare)
	}

	if f.ss.gotOpts.UserID != f.ur.user.ID || f.ss.gotOpts.AuthMethod != sessions.AuthMethodPassword {
		t.Errorf("CreateSessionOpts = %+v", f.ss.gotOpts)
	}
}

func TestLoginWithPwdOpensNoTransaction(t *testing.T) {
	t.Parallel()

	f := newFakes()

	if _, err := f.svc(t).LoginWithPwd(
		context.Background(),
		auth.LoginWithPwdOpts{Username: userName, Password: pwd},
	); err != nil {
		t.Fatalf("LoginWithPwd: %v", err)
	}

	if f.tr.calls != 0 {
		t.Errorf("WithinTx called %d times by LoginWithPwd", f.tr.calls)
	}
}

func TestLoginWithPwdRejectsBadCredentials(t *testing.T) {
	t.Parallel()

	tests := map[string]func(*fakes){
		"unknown user": func(f *fakes) { f.ur.byNameErr = domain.ErrNotFound },
		"no password":       func(f *fakes) { f.pr.getErr = domain.ErrNotFound },
		"wrong password": func(f *fakes) { f.hs.match = false },
	}

	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newFakes()
			setup(f)

			got, err := f.svc(t).LoginWithPwd(context.Background(), auth.LoginWithPwdOpts{
				Username: userName,
				Password: pwd,
			})
			if !errors.Is(err, auth.ErrInvalidLoginPwd) {
				t.Errorf("err = %v, want ErrInvalidLoginPwd", err)
			}

			if got != nil {
				t.Errorf("LoginWithPwd returned %+v in addition to the error", got)
			}

			if f.ss.calls != 0 {
				t.Errorf("Create called %d times although login failed", f.ss.calls)
			}
		})
	}
}

func TestLoginWithPwdBurnsHashTimeOnUnknownAccount(t *testing.T) {
	t.Parallel()

	tests := map[string]func(*fakes){
		"unknown user": func(f *fakes) { f.ur.byNameErr = domain.ErrNotFound },
		"no password":       func(f *fakes) { f.pr.getErr = domain.ErrNotFound },
	}

	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newFakes()
			setup(f)

			if _, err := f.svc(t).LoginWithPwd(context.Background(), auth.LoginWithPwdOpts{
				Username: userName,
				Password: pwd,
			}); !errors.Is(err, auth.ErrInvalidLoginPwd) {
				t.Fatalf("err = %v, want ErrInvalidLoginPwd", err)
			}

			if f.hs.hashCalls != 1 {
				t.Errorf("Hash called %d times, want 1", f.hs.hashCalls)
			}

			if string(f.hs.got) != pwd {
				t.Errorf("Hash received %q, want provided password", f.hs.got)
			}
		})
	}
}

func TestLoginWithPwdStaysInvalidWhenBurnHashFails(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.ur.byNameErr = domain.ErrNotFound
	f.hs.err = hash.ErrStringTooLong

	if _, err := f.svc(t).LoginWithPwd(context.Background(), auth.LoginWithPwdOpts{
		Username: userName,
		Password: pwd,
	}); !errors.Is(err, auth.ErrInvalidLoginPwd) {
		t.Errorf("err = %v, want ErrInvalidLoginPwd", err)
	}
}

func TestLoginWithPwdPropagatesInfrastructureErrors(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("connection refused")

	tests := map[string]func(*fakes){
		"user lookup": func(f *fakes) { f.ur.byNameErr = sentinel },
		"credentials lookup": func(f *fakes) { f.pr.getErr = sentinel },
		"comparaison de hash": func(f *fakes) { f.hs.matchErr = sentinel },
		"session issuance": func(f *fakes) { f.ss.err = sentinel },
	}

	for name, setup := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newFakes()
			setup(f)

			got, err := f.svc(t).LoginWithPwd(context.Background(), auth.LoginWithPwdOpts{
				Username: userName,
				Password: pwd,
			})
			if !errors.Is(err, sentinel) {
				t.Errorf("err = %v, original error no longer reachable", err)
			}

			if errors.Is(err, auth.ErrInvalidLoginPwd) {
				t.Errorf("err = %v, infrastructure failure must not present as wrong password", err)
			}

			if got != nil {
				t.Errorf("LoginWithPwd returned %+v in addition to the error", got)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	t.Parallel()

	providerID := uuid.New()
	endSessionURL := "https://idp.example.com/logout?client_id=client-id&" +
		"post_logout_redirect_uri=https%3A%2F%2Fapp.example.com%2Flogin"

	tests := map[string]struct {
		session    sessions.Session
		setup      func(*fakes)
		wantURL    string
		wantRevoke bool
		wantErr    error
	}{
		"password session → empty EndSessionURL": {
			session:    sessions.Session{AuthMethod: sessions.AuthMethodPassword},
			wantRevoke: true,
		},
		"oidc + EndSessionURL supported → URL returned": {
			session: sessions.Session{
				AuthMethod: sessions.AuthMethodOIDC,
				ProviderID: &providerID,
			},
			setup: func(f *fakes) {
				f.opr.provider.ID = providerID
				f.oc.endSessionURL = endSessionURL
				f.oc.endSessionSupported = true
			},
			wantURL:    endSessionURL,
			wantRevoke: true,
		},
		"oidc + EndSessionURL not supported → empty": {
			session: sessions.Session{
				AuthMethod: sessions.AuthMethodOIDC,
				ProviderID: &providerID,
			},
			setup: func(f *fakes) {
				f.opr.provider.ID = providerID
				f.oc.endSessionSupported = false
			},
			wantRevoke: true,
		},
		"oidc + provider not found → empty": {
			session: sessions.Session{
				AuthMethod: sessions.AuthMethodOIDC,
				ProviderID: ptrUUID(uuid.New()),
			},
			setup: func(f *fakes) {
				f.opr.err = domain.ErrNotFound
			},
			wantRevoke: true,
		},
		"oidc + EndSessionURL error → empty (logout succeeds)": {
			session: sessions.Session{
				AuthMethod: sessions.AuthMethodOIDC,
				ProviderID: &providerID,
			},
			setup: func(f *fakes) {
				f.opr.provider.ID = providerID
				f.oc.endSessionErr = oidc.ErrClientUnavailable
			},
			wantRevoke: true,
		},
		"revoke error": {
			session: sessions.Session{AuthMethod: sessions.AuthMethodPassword},
			setup: func(f *fakes) {
				f.ss.revokeErr = sessions.ErrInvalidSession
			},
			wantErr: sessions.ErrInvalidSession,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := newFakes()
			if tc.setup != nil {
				tc.setup(f)
			}

			got, err := f.svc(t).Logout(context.Background(), auth.LogoutOpts{
				Token:   token,
				Session: tc.session,
			})

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("Logout = %v, want %v", err, tc.wantErr)
				}

				if got != nil {
					t.Fatalf("Logout = %+v, want nil result on error", got)
				}

				return
			}

			if err != nil {
				t.Fatalf("Logout: %v", err)
			}

			if got == nil {
				t.Fatal("Logout returned nil result")
			}

			if got.EndSessionURL != tc.wantURL {
				t.Errorf("EndSessionURL = %q, want %q", got.EndSessionURL, tc.wantURL)
			}

			if tc.wantRevoke && f.ss.revokeCalls != 1 {
				t.Errorf("Revoke called %d times, want 1", f.ss.revokeCalls)
			}

			if f.ss.gotToken != token {
				t.Errorf("Revoke received %q, want %q", f.ss.gotToken, token)
			}

			if tc.wantURL != "" {
				if f.oc.endSessionCalls != 1 {
					t.Errorf("EndSessionURL called %d times, want 1", f.oc.endSessionCalls)
				}

				if f.oc.gotPostLogoutRedirect != publicURL+"/login" {
					t.Errorf(
						"post_logout_redirect_uri = %q, want %q",
						f.oc.gotPostLogoutRedirect,
						publicURL+"/login",
					)
				}
			}
		})
	}
}

func ptrUUID(id uuid.UUID) *uuid.UUID {
	return &id
}
