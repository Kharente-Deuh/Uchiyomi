// SPDX-License-Identifier: AGPL-3.0-or-later

package pg_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users/repository/pg"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgtest"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
)

const (
	userNameBob   = "bob"
	userNameAlice = "alice"
	colName       = "name"
)

func duplicateKeyErr() error {
	return &pgconn.PgError{Code: "23505", Message: `duplicate key value violates unique constraint "idx_users_name"`}
}

func newRepo(t *testing.T) (*pg.PGUsersRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock := pgtest.New(t)

	r, err := pg.New(pg.Deps{DB: db})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return r, mock
}

func TestNewValidatesDeps(t *testing.T) {
	t.Parallel()

	r, err := pg.New(pg.Deps{})
	if err == nil {
		t.Fatal("New without DB must fail")
	}

	if r != nil {
		t.Error("New returned a repository in addition to the error")
	}

	if want := "deps.Validate: db is required"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
}

func TestCountAdmins(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "users" WHERE is_admin = \$1`).
		WithArgs(true).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	got, err := r.CountAdmins(context.Background())
	if err != nil {
		t.Fatalf("CountAdmins: %v", err)
	}

	if got != 3 {
		t.Errorf("CountAdmins() = %d, want 3", got)
	}
}

func TestCountAdminsZero(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectQuery(`SELECT count`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	got, err := r.CountAdmins(context.Background())
	if err != nil {
		t.Fatalf("CountAdmins: %v", err)
	}

	if got != 0 {
		t.Errorf("CountAdmins() = %d, want 0", got)
	}
}

func TestCountAdminsError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("connection refused")
	mock.ExpectQuery(`SELECT count`).WillReturnError(sentinel)

	got, err := r.CountAdmins(context.Background())
	if err == nil {
		t.Fatal("CountAdmins must propagate SQL error")
	}

	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, original error no longer reachable", err)
	}

	if got != 0 {
		t.Errorf("CountAdmins() = %d, want 0 on error", got)
	}
}

func TestGetByID(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	id := uuid.New()
	created := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	updated := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(`SELECT \* FROM "users" WHERE id = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs(id, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", colName, "is_admin", "created_at", "updated_at"}).
				AddRow(id, userNameBob, true, created, updated),
		)

	got, err := r.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	want := users.User{ID: id, Name: userNameBob, IsAdmin: true, CreatedAt: created, UpdatedAt: updated}
	if got == nil || *got != want {
		t.Errorf("GetByID() = %+v, want %+v", got, want)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectQuery(`SELECT \* FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", colName}))

	got, err := r.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID = %v, want domain.ErrNotFound", err)
	}

	if got != nil {
		t.Errorf("GetByID returned %+v in addition to the error", got)
	}
}

func TestGetByIDError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("connection reset")
	mock.ExpectQuery(`SELECT \* FROM "users"`).WillReturnError(sentinel)

	_, err := r.GetByID(context.Background(), uuid.New())
	if errors.Is(err, domain.ErrNotFound) {
		t.Error("SQL failure must not be translated to ErrNotFound")
	}

	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, original error no longer reachable", err)
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	before := time.Now()

	got, err := r.Create(context.Background(), users.CreateUserOpts{Name: userNameAlice, IsAdmin: true})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if got.Name != userNameAlice || !got.IsAdmin {
		t.Errorf("Create() = %+v", got)
	}

	if got.ID == uuid.Nil {
		t.Error("Create did not generate ID")
	}

	if got.CreatedAt.Before(before) || got.UpdatedAt.Before(before) {
		t.Errorf("timestamps not set: created=%v updated=%v", got.CreatedAt, got.UpdatedAt)
	}
}

func TestCreateJoinsAmbientTransaction(t *testing.T) {
	t.Parallel()

	db, mock := pgtest.New(t)

	r, err := pg.New(pg.Deps{DB: db})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	tr, err := pgtx.New(pgtx.Deps{DB: db})
	if err != nil {
		t.Fatalf("pgtx.New: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	var created *users.User

	err = tr.WithinTx(context.Background(), transaction.TxOpts{}, func(ctx context.Context) error {
		var createErr error

		created, createErr = r.Create(ctx, users.CreateUserOpts{Name: userNameAlice, IsAdmin: true})

		if createErr != nil {
			return fmt.Errorf("r.Create: %w", createErr)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("WithinTx: %v", err)
	}

	if created == nil || created.Name != userNameAlice {
		t.Errorf("Create() = %+v", created)
	}
}

func TestCreateOutsideTransaction(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	if _, err := r.Create(context.Background(), users.CreateUserOpts{Name: userNameBob}); err != nil {
		t.Fatalf("Create: %v", err)
	}
}

func TestCreateDuplicate(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).WillReturnError(duplicateKeyErr())
	mock.ExpectRollback()

	got, err := r.Create(context.Background(), users.CreateUserOpts{Name: userNameBob})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("Create = %v, want domain.ErrAlreadyExists", err)
	}

	if got != nil {
		t.Errorf("Create returned %+v in addition to the error", got)
	}
}

func TestCreateError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("disk full")
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).WillReturnError(sentinel)
	mock.ExpectRollback()

	_, err := r.Create(context.Background(), users.CreateUserOpts{Name: userNameBob})
	if errors.Is(err, domain.ErrAlreadyExists) {
		t.Error("SQL failure must not be translated to ErrAlreadyExists")
	}

	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, original error no longer reachable", err)
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	id := uuid.New()
	created := time.Now().Add(-2 * time.Hour).UTC().Truncate(time.Second)
	updated := time.Now().UTC().Truncate(time.Second)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users" SET "is_admin"=\$1,"updated_at"=\$2 WHERE id = \$3`).
		WithArgs(true, sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	mock.ExpectQuery(`SELECT \* FROM "users" WHERE id = \$1 ORDER BY "users"\."id" LIMIT \$2`).
		WithArgs(id, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", colName, "is_admin", "created_at", "updated_at"}).
				AddRow(id, userNameBob, true, created, updated),
		)

	got, err := r.Update(context.Background(), users.UpdateUserOpts{ID: id, IsAdmin: true})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	want := users.User{ID: id, Name: userNameBob, IsAdmin: true, CreatedAt: created, UpdatedAt: updated}
	if got == nil || *got != want {
		t.Errorf("Update() = %+v, want %+v", got, want)
	}
}

func TestUpdateNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users" SET "is_admin"=\$1,"updated_at"=\$2 WHERE id = \$3`).
		WithArgs(false, sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	got, err := r.Update(context.Background(), users.UpdateUserOpts{ID: id, IsAdmin: false})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Update = %v, want domain.ErrNotFound", err)
	}

	if got != nil {
		t.Errorf("Update returned %+v in addition to the error", got)
	}
}
