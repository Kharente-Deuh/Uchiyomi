// SPDX-License-Identifier: AGPL-3.0-or-later

package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/hash"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/password"
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
	err         error
	user        *users.User
	byID        map[uuid.UUID]*users.User
	gotName     string
	gotOpts     users.CreateUserOpts
	nameCalls   int
	byIDCalls   int
	updateCalls int
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
	panic("CountAdmins n'est pas utilisée par le service auth")
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
	panic("UpdateByUserID n'est pas utilisée par le service auth")
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
	panic("Authenticate n'est pas utilisée par le service auth")
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
	panic("RevokeAllForUser n'est pas utilisée par le service auth")
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
		cfg:   auth.Config{RedirectURI: oidcRedirectURI, StateCookieTTL: 10 * time.Minute},
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
		sc: sc,
	}
}

func (f *fakes) svc(t *testing.T) *auth.Service {
	t.Helper()

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
		auth.Config{RedirectURI: oidcRedirectURI},
		auth.Deps{HashService: f.hs, UsersRepository: f.ur, PwdRepository: f.pr},
	)
	if err == nil {
		t.Fatal("New sans transactor doit échouer")
	}

	if svc != nil {
		t.Error("New a renvoyé un service en plus de l'erreur")
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
		t.Errorf("WithinTx appelée %d fois, want 1", f.tr.calls)
	}

	if !f.ur.inTx {
		t.Error("UsersRepository.Create n'a pas reçu le ctx transactionnel")
	}

	if !f.pr.inTx {
		t.Error("PwdRepository.Create n'a pas reçu le ctx transactionnel")
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
		t.Fatalf("appels non observés: hashedAt=%d beganAt=%d", f.hs.hashedAt, f.tr.beganAt)
	}

	if f.hs.hashedAt > f.tr.beganAt {
		t.Errorf("hash effectué dans la transaction (hashedAt=%d > beganAt=%d)", f.hs.hashedAt, f.tr.beganAt)
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
		t.Errorf("Hash a reçu %q", f.hs.got)
	}

	if f.pr.gotOpts.Hash != "hashed:"+pwd {
		t.Errorf("PwdRepository a reçu le hash %q", f.pr.gotOpts.Hash)
	}

	if f.pr.gotOpts.UserID != f.ur.user.ID {
		t.Errorf("PwdRepository a reçu UserID %v, want %v", f.pr.gotOpts.UserID, f.ur.user.ID)
	}
}

func TestCreateUserWithPwdReturnsNoUserWhenPwdWriteFails(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.pr.err = errors.New("disk full")

	got, err := f.svc(t).CreateUserWithPwd(context.Background(), auth.CreateUserWithPwdOpts{Name: userName})
	if !errors.Is(err, f.pr.err) {
		t.Errorf("err = %v, l'erreur d'origine n'est plus atteignable", err)
	}

	if got != nil {
		t.Errorf("CreateUserWithPwd a renvoyé %+v alors que la transaction a échoué", got)
	}
}

func TestCreateUserWithPwdReturnsNoUserWhenCommitFails(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.tr.commitErr = errors.New("commit refusé")

	got, err := f.svc(t).CreateUserWithPwd(context.Background(), auth.CreateUserWithPwdOpts{Name: userName})
	if !errors.Is(err, f.tr.commitErr) {
		t.Errorf("err = %v, want l'erreur de commit", err)
	}

	if got != nil {
		t.Errorf("CreateUserWithPwd a renvoyé %+v alors que le commit a échoué", got)
	}
}

func TestCreateUserWithPwdPropagatesSentinels(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		setup func(*fakes)
		want  error
	}{
		"mot de passe trop long": {
			setup: func(f *fakes) { f.hs.err = hash.ErrStringTooLong },
			want:  hash.ErrStringTooLong,
		},
		"nom déjà pris": {
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
				t.Errorf("CreateUserWithPwd a renvoyé %+v en plus de l'erreur", got)
			}
		})
	}
}

func TestCreateUserWithPwdSkipsTransactionWhenHashFails(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.hs.err = hash.ErrStringTooLong

	if _, err := f.svc(t).CreateUserWithPwd(context.Background(), auth.CreateUserWithPwdOpts{Name: "a"}); err == nil {
		t.Fatal("CreateUserWithPwd doit échouer")
	}

	if f.tr.calls != 0 {
		t.Errorf("WithinTx appelée %d fois alors que le hash a échoué", f.tr.calls)
	}

	if f.pr.calls != 0 {
		t.Errorf("PwdRepository.Create appelée %d fois alors que le hash a échoué", f.pr.calls)
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
		t.Errorf("GetByUsername a reçu %q, want %q", f.ur.gotName, userName)
	}

	if f.pr.gotUserID != f.ur.user.ID {
		t.Errorf("GetByUserID a reçu %v, want %v", f.pr.gotUserID, f.ur.user.ID)
	}

	if string(f.hs.gotHashed) != f.pr.creds.Hash || string(f.hs.gotCompare) != pwd {
		t.Errorf("Match a reçu (%q, %q)", f.hs.gotHashed, f.hs.gotCompare)
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
		t.Errorf("WithinTx appelée %d fois par LoginWithPwd", f.tr.calls)
	}
}

func TestLoginWithPwdRejectsBadCredentials(t *testing.T) {
	t.Parallel()

	tests := map[string]func(*fakes){
		"utilisateur inconnu": func(f *fakes) { f.ur.byNameErr = domain.ErrNotFound },
		"aucun mot de passe":  func(f *fakes) { f.pr.getErr = domain.ErrNotFound },
		"mot de passe erroné": func(f *fakes) { f.hs.match = false },
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
				t.Errorf("LoginWithPwd a renvoyé %+v en plus de l'erreur", got)
			}

			if f.ss.calls != 0 {
				t.Errorf("Create appelée %d fois alors que le login a échoué", f.ss.calls)
			}
		})
	}
}

func TestLoginWithPwdBurnsHashTimeOnUnknownAccount(t *testing.T) {
	t.Parallel()

	tests := map[string]func(*fakes){
		"utilisateur inconnu": func(f *fakes) { f.ur.byNameErr = domain.ErrNotFound },
		"aucun mot de passe":  func(f *fakes) { f.pr.getErr = domain.ErrNotFound },
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
				t.Errorf("Hash appelée %d fois, want 1", f.hs.hashCalls)
			}

			if string(f.hs.got) != pwd {
				t.Errorf("Hash a reçu %q, want le mot de passe fourni", f.hs.got)
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
		"lecture utilisateur": func(f *fakes) { f.ur.byNameErr = sentinel },
		"lecture credentials": func(f *fakes) { f.pr.getErr = sentinel },
		"comparaison de hash": func(f *fakes) { f.hs.matchErr = sentinel },
		"émission de session": func(f *fakes) { f.ss.err = sentinel },
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
				t.Errorf("err = %v, l'erreur d'origine n'est plus atteignable", err)
			}

			if errors.Is(err, auth.ErrInvalidLoginPwd) {
				t.Errorf("err = %v, une panne ne doit pas se présenter comme un mauvais mot de passe", err)
			}

			if got != nil {
				t.Errorf("LoginWithPwd a renvoyé %+v en plus de l'erreur", got)
			}
		})
	}
}

func TestLogoutForwardsTokenToRevoke(t *testing.T) {
	t.Parallel()

	f := newFakes()

	if err := f.svc(t).Logout(context.Background(), token); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	if f.ss.revokeCalls != 1 {
		t.Errorf("Revoke appelée %d fois, want 1", f.ss.revokeCalls)
	}

	if f.ss.gotToken != token {
		t.Errorf("Revoke a reçu %q, want %q", f.ss.gotToken, token)
	}
}

func TestLogoutPropagatesRevokeError(t *testing.T) {
	t.Parallel()

	f := newFakes()
	f.ss.revokeErr = sessions.ErrInvalidSession

	err := f.svc(t).Logout(context.Background(), token)
	if !errors.Is(err, sessions.ErrInvalidSession) {
		t.Errorf("err = %v, want ErrInvalidSession", err)
	}
}

func TestLogoutReturnsNilOnSuccess(t *testing.T) {
	t.Parallel()

	f := newFakes()

	if err := f.svc(t).Logout(context.Background(), token); err != nil {
		t.Errorf("Logout() = %v, want nil", err)
	}
}
