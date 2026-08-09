// SPDX-License-Identifier: AGPL-3.0-or-later

package setup_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/setup"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
)

const (
	adminName     = "alice"
	adminPassword = "hunter2hunter2"
)

type txCtxKey struct{}

type fakeTransactor struct {
	commitErr error
	opts      transaction.TxOpts
	calls     int
}

func (f *fakeTransactor) WithinTx(ctx context.Context, opts transaction.TxOpts, fn func(context.Context) error) error {
	f.calls++
	f.opts = opts

	if err := fn(context.WithValue(ctx, txCtxKey{}, true)); err != nil {
		return err
	}

	return f.commitErr
}

type fakeAuthService struct {
	user    *users.User
	err     error
	gotOpts auth.CreateUserWithPwdOpts
	calls   int
	gotInTx bool
}

func (f *fakeAuthService) CreateUserWithPwd(
	ctx context.Context,
	opts auth.CreateUserWithPwdOpts,
) (*users.User, error) {
	f.calls++
	f.gotOpts = opts
	f.gotInTx, _ = ctx.Value(txCtxKey{}).(bool)

	if f.err != nil {
		return nil, f.err
	}

	return f.user, nil
}

func (f *fakeAuthService) LoginWithPwd(context.Context, auth.LoginWithPwdOpts) (*auth.LoginResult, error) {
	panic("LoginWithPwd ne doit pas être appelée par DoSetup")
}

func (f *fakeAuthService) Logout(context.Context, string) error {
	panic("Logout ne doit pas être appelée par DoSetup")
}

type stubSessionService struct {
	err     error
	issued  *sessions.IssuedSession
	gotOpts sessions.CreateSessionOpts
	inTx    bool
	calls   int
}

func (s *stubSessionService) Create(
	ctx context.Context,
	opts sessions.CreateSessionOpts,
) (*sessions.IssuedSession, error) {
	s.calls++
	s.gotOpts = opts
	s.inTx, _ = ctx.Value(txCtxKey{}).(bool)

	if s.err != nil {
		return nil, s.err
	}

	return s.issued, nil
}

func (s *stubSessionService) Authenticate(context.Context, string) (*sessions.AuthenticatedSession, error) {
	panic("Authenticate ne doit pas être appelée par DoSetup")
}

func (s *stubSessionService) Revoke(context.Context, string) error {
	panic("Revoke ne doit pas être appelée par DoSetup")
}

func (s *stubSessionService) RevokeAllForUser(context.Context, uuid.UUID) error {
	panic("RevokeAllForUser ne doit pas être appelée par DoSetup")
}

func defaultSessionStub() *stubSessionService {
	return &stubSessionService{issued: &sessions.IssuedSession{
		Session: sessions.Session{ID: uuid.New()},
		Token:   "letoken",
	}}
}

func newSvc(
	t *testing.T,
	repo users.UsersRepository,
	as auth.AuthService,
	tr transaction.Transactor,
	ss sessions.SessionService,
) *setup.Service {
	t.Helper()

	svc, err := setup.New(setup.Deps{UsersRepository: repo, AuthService: as, Transactor: tr, SessionService: ss})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return svc
}

func TestDoSetupCreatesAdminInsideTransaction(t *testing.T) {
	t.Parallel()

	repo := &fakeUsersRepository{count: 0}
	as := &fakeAuthService{user: &users.User{ID: uuid.New(), Name: adminName, IsAdmin: true}}
	tr := &fakeTransactor{}

	svc := newSvc(t, repo, as, tr, defaultSessionStub())

	_, err := svc.DoSetup(context.Background(), setup.DoSetupOpts{Username: adminName, Password: adminPassword})
	if err != nil {
		t.Fatalf("DoSetup: %v", err)
	}

	if as.gotOpts.Name != adminName || !as.gotOpts.IsAdmin {
		t.Errorf("CreateUserWithPwd reçoit %+v, want un admin nommé %q", as.gotOpts, adminName)
	}

	if tr.calls != 1 {
		t.Errorf("WithinTx appelée %d fois, want 1", tr.calls)
	}

	if !as.gotInTx {
		t.Error("CreateUserWithPwd n'a pas reçu le ctx transactionnel")
	}

	if !as.gotOpts.IsAdmin {
		t.Error("le premier utilisateur doit être créé admin")
	}

	if as.gotOpts.Name != adminName || as.gotOpts.Password != adminPassword {
		t.Errorf("CreateUserWithPwd opts = %+v", as.gotOpts)
	}
}

func TestDoSetupDemandsSerializable(t *testing.T) {
	t.Parallel()

	tr := &fakeTransactor{}
	svc := newSvc(t, &fakeUsersRepository{count: 0}, &fakeAuthService{user: &users.User{}}, tr, defaultSessionStub())

	if _, err := svc.DoSetup(context.Background(), setup.DoSetupOpts{Username: "a", Password: "b"}); err != nil {
		t.Fatalf("DoSetup: %v", err)
	}

	if tr.opts.Isolation != transaction.IsolationSerializable {
		t.Errorf("Isolation = %v, want IsolationSerializable", tr.opts.Isolation)
	}
}

func TestDoSetupRefusesWhenAdminExists(t *testing.T) {
	t.Parallel()

	as := &fakeAuthService{user: &users.User{}}
	svc := newSvc(t, &fakeUsersRepository{count: 1}, as, &fakeTransactor{}, defaultSessionStub())

	got, err := svc.DoSetup(context.Background(), setup.DoSetupOpts{Username: "eve", Password: "pwned"})
	if !errors.Is(err, setup.ErrSetupNotNeeded) {
		t.Errorf("DoSetup = %v, want setup.ErrSetupNotNeeded", err)
	}

	if got != nil {
		t.Errorf("DoSetup a renvoyé %+v en plus de l'erreur", got)
	}

	if as.calls != 0 {
		t.Errorf("CreateUserWithPwd appelée %d fois alors qu'un admin existe", as.calls)
	}
}

func TestDoSetupReadsGuardInsideTransaction(t *testing.T) {
	t.Parallel()

	var seenTx bool

	repo := &ctxProbeRepository{seen: &seenTx}
	svc := newSvc(t, repo, &fakeAuthService{user: &users.User{}}, &fakeTransactor{}, defaultSessionStub())

	if _, err := svc.DoSetup(context.Background(), setup.DoSetupOpts{Username: "a", Password: "b"}); err != nil {
		t.Fatalf("DoSetup: %v", err)
	}

	if !seenTx {
		t.Error("CountAdmins a été appelée hors de la transaction")
	}
}

type ctxProbeRepository struct {
	seen *bool
	fakeUsersRepository
}

func (c *ctxProbeRepository) CountAdmins(ctx context.Context) (int, error) {
	*c.seen, _ = ctx.Value(txCtxKey{}).(bool)

	return c.fakeUsersRepository.CountAdmins(ctx)
}

func TestDoSetupPropagatesAuthError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("mot de passe trop long")
	as := &fakeAuthService{err: sentinel}
	svc := newSvc(t, &fakeUsersRepository{count: 0}, as, &fakeTransactor{}, defaultSessionStub())

	got, err := svc.DoSetup(context.Background(), setup.DoSetupOpts{Username: "a", Password: "b"})
	if err == nil {
		t.Fatal("DoSetup doit remonter l'erreur de l'auth service")
	}

	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, l'erreur d'origine n'est plus atteignable via errors.Is", err)
	}

	if got != nil {
		t.Errorf("DoSetup a renvoyé %+v en plus de l'erreur", got)
	}
}

func TestDoSetupPropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("connection refused")
	as := &fakeAuthService{user: &users.User{}}
	svc := newSvc(t, &fakeUsersRepository{countErr: sentinel}, as, &fakeTransactor{}, defaultSessionStub())

	_, err := svc.DoSetup(context.Background(), setup.DoSetupOpts{Username: "a", Password: "b"})
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, want l'erreur du repository", err)
	}

	if as.calls != 0 {
		t.Errorf("CreateUserWithPwd appelée %d fois alors que la garde a échoué", as.calls)
	}
}

func TestDoSetupReturnsNoUserWhenCommitFails(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("commit refusé")
	tr := &fakeTransactor{commitErr: sentinel}
	as := &fakeAuthService{user: &users.User{ID: uuid.New(), Name: adminName}}
	svc := newSvc(t, &fakeUsersRepository{count: 0}, as, tr, defaultSessionStub())

	got, err := svc.DoSetup(context.Background(), setup.DoSetupOpts{Username: "a", Password: "b"})
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, want l'erreur de la transaction", err)
	}

	if got != nil {
		t.Errorf("DoSetup a renvoyé %+v alors que la transaction a échoué", got)
	}
}

func TestDoSetupIssuesPasswordSession(t *testing.T) {
	t.Parallel()

	want := &users.User{ID: uuid.New(), Name: adminName, IsAdmin: true}
	ss := defaultSessionStub()
	svc := newSvc(t, &fakeUsersRepository{count: 0}, &fakeAuthService{user: want}, &fakeTransactor{}, ss)

	got, err := svc.DoSetup(context.Background(), setup.DoSetupOpts{Username: adminName, Password: adminPassword})
	if err != nil {
		t.Fatalf("DoSetup: %v", err)
	}

	if got == nil {
		t.Fatal("DoSetup n'a pas émis de session")
	}

	if ss.gotOpts.AuthMethod != sessions.AuthMethodPassword {
		t.Errorf("AuthMethod = %v, want password", ss.gotOpts.AuthMethod)
	}

	if ss.gotOpts.UserID != want.ID {
		t.Errorf("UserID = %v, want %v", ss.gotOpts.UserID, want.ID)
	}
}

func TestDoSetupIssuesSessionOutsideTransaction(t *testing.T) {
	t.Parallel()

	ss := defaultSessionStub()
	svc := newSvc(t, &fakeUsersRepository{count: 0}, &fakeAuthService{user: &users.User{}}, &fakeTransactor{}, ss)

	if _, err := svc.DoSetup(context.Background(), setup.DoSetupOpts{Username: "a", Password: "b"}); err != nil {
		t.Fatalf("DoSetup: %v", err)
	}

	if ss.inTx {
		t.Error("Create a reçu le ctx transactionnel")
	}
}

func TestDoSetupSignalsSessionFailureWithoutSession(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("disk full")
	ss := defaultSessionStub()
	ss.err = sentinel
	as := &fakeAuthService{user: &users.User{ID: uuid.New(), Name: adminName}}
	svc := newSvc(t, &fakeUsersRepository{count: 0}, as, &fakeTransactor{}, ss)

	got, err := svc.DoSetup(context.Background(), setup.DoSetupOpts{Username: adminName, Password: adminPassword})
	if !errors.Is(err, setup.ErrSessionNotIssued) {
		t.Errorf("err = %v, want ErrSessionNotIssued", err)
	}

	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, la cause d'origine n'est plus atteignable", err)
	}

	if got != nil {
		t.Errorf("session = %+v, want nil", got)
	}

	if as.calls != 1 {
		t.Errorf("CreateUserWithPwd appelée %d fois : le compte doit rester créé", as.calls)
	}
}

func TestDoSetupSkipsSessionWhenTransactionFails(t *testing.T) {
	t.Parallel()

	ss := defaultSessionStub()
	tr := &fakeTransactor{commitErr: errors.New("commit refusé")}
	svc := newSvc(t, &fakeUsersRepository{count: 0}, &fakeAuthService{user: &users.User{}}, tr, ss)

	got, err := svc.DoSetup(context.Background(), setup.DoSetupOpts{Username: adminName, Password: adminPassword})
	if err == nil {
		t.Fatal("DoSetup doit échouer")
	}

	if got != nil {
		t.Errorf("DoSetup a renvoyé %+v alors que la transaction a échoué", got)
	}

	if ss.calls != 0 {
		t.Errorf("Create appelée %d fois alors qu'aucun administrateur n'existe", ss.calls)
	}
}
