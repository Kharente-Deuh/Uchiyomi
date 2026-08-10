// SPDX-License-Identifier: AGPL-3.0-or-later

package setup_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/setup"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
)

type fakeUsersRepository struct {
	countErr   error
	count      int
	countCalls int
}

func (f *fakeUsersRepository) CountAdmins(context.Context) (int, error) {
	f.countCalls++

	return f.count, f.countErr
}

func (f *fakeUsersRepository) GetByID(context.Context, uuid.UUID) (*users.User, error) {
	panic("GetByID must not be called by the setup service")
}

func (f *fakeUsersRepository) Create(context.Context, users.CreateUserOpts) (*users.User, error) {
	panic("Create must not be called by the setup service")
}

func (f *fakeUsersRepository) GetByUsername(context.Context, string) (*users.User, error) {
	panic("GetByUsername must not be called by the setup service")
}

func (f *fakeUsersRepository) Update(context.Context, users.UpdateUserOpts) (*users.User, error) {
	panic("Update must not be called by the setup service")
}

func depsFor(repo users.UsersRepository) setup.Deps {
	return setup.Deps{
		UsersRepository: repo,
		AuthService:     &fakeAuthService{},
		Transactor:      &fakeTransactor{},
		SessionService:  defaultSessionStub(),
	}
}

func TestNewRejectsMissingRepository(t *testing.T) {
	t.Parallel()

	svc, err := setup.New(setup.Deps{})
	if err == nil {
		t.Fatal("New without repository must fail")
	}

	if svc != nil {
		t.Errorf("New returned a service (%v) in addition to the error", svc)
	}

	if got := err.Error(); got != "deps.Validate: authService is required" {
		t.Errorf("err = %q, want %q", got, "deps.Validate: authService is required")
	}
}

func TestNewSucceeds(t *testing.T) {
	t.Parallel()

	svc, err := setup.New(depsFor(&fakeUsersRepository{}))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if svc == nil {
		t.Fatal("New returned nil service without error")
	}
}

func TestIsSetupRequired(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		count int
		want  bool
	}{
		"no admin => setup required":      {count: 0, want: true},
		"one admin => setup not required": {count: 1, want: false},
		"multiple admins => not required": {count: 12, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := &fakeUsersRepository{count: tc.count}
			svc, err := setup.New(depsFor(repo))
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			got, err := svc.IsSetupRequired(context.Background())
			if err != nil {
				t.Fatalf("IsSetupRequired: %v", err)
			}

			if got != tc.want {
				t.Errorf("IsSetupRequired() = %v, want %v (count=%d)", got, tc.want, tc.count)
			}

			if repo.countCalls != 1 {
				t.Errorf("CountAdmins called %d times, want 1", repo.countCalls)
			}
		})
	}
}

func TestIsSetupRequiredPropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("connection refused")
	svc, err := setup.New(depsFor(&fakeUsersRepository{count: 3, countErr: sentinel}))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	got, err := svc.IsSetupRequired(context.Background())
	if err == nil {
		t.Fatal("IsSetupRequired must propagate repository error")
	}

	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, original error no longer reachable via errors.Is", err)
	}

	if want := "s.deps.UsersRepository.CountAdmins: connection refused"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}

	if got {
		t.Error("IsSetupRequired = true although repository failed")
	}
}

func TestIsSetupRequiredPassesContext(t *testing.T) {
	t.Parallel()

	type ctxKey string

	var seen context.Context

	repo := &ctxCapturingRepository{fakeUsersRepository: fakeUsersRepository{}, seen: &seen}
	svc, err := setup.New(depsFor(repo))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	key := ctxKey("trace-id")
	ctx := context.WithValue(context.Background(), key, "abc123")

	if _, err := svc.IsSetupRequired(ctx); err != nil {
		t.Fatalf("IsSetupRequired: %v", err)
	}

	if seen == nil {
		t.Fatal("repository received no context")
	}

	if got := seen.Value(key); got != "abc123" {
		t.Errorf("context lost along the way: value = %v, want %q", got, "abc123")
	}
}

type ctxCapturingRepository struct {
	seen *context.Context
	fakeUsersRepository
}

func (c *ctxCapturingRepository) CountAdmins(ctx context.Context) (int, error) {
	*c.seen = ctx

	return c.fakeUsersRepository.CountAdmins(ctx)
}
