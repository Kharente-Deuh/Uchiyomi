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
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/password"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/password/repository/pg"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgtest"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
)

const newHash = "new"

func duplicateKeyErr() error {
	return &pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint"}
}

func newRepo(t *testing.T) (*pg.PGPasswordCredsRepository, sqlmock.Sqlmock) {
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

	if _, err := pg.New(pg.Deps{}); err == nil || err.Error() != "deps.Validate: db is required" {
		t.Errorf("pg.New(pg.Deps{}) = %v, want %q", err, "deps.Validate: db is required")
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
	mock.ExpectExec(`INSERT INTO "password_credentials"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = tr.WithinTx(context.Background(), transaction.TxOpts{}, func(ctx context.Context) error {
		_, createErr := r.Create(ctx, password.UpsertPasswordCredsOpts{UserID: uuid.New(), Hash: "$2a$10$x"})

		if createErr != nil {
			return fmt.Errorf("r.Create: %w", createErr)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("WithinTx: %v", err)
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	userID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "password_credentials"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	before := time.Now()

	got, err := r.Create(context.Background(), password.UpsertPasswordCredsOpts{UserID: userID, Hash: "argon2$xxx"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if got.UserID != userID || got.Hash != "argon2$xxx" {
		t.Errorf("Create() = %+v", got)
	}

	if got.UpdatedAt.Before(before) {
		t.Errorf("UpdatedAt = %v, not set by repository", got.UpdatedAt)
	}
}

func TestCreateDuplicate(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "password_credentials"`).WillReturnError(duplicateKeyErr())
	mock.ExpectRollback()

	got, err := r.Create(context.Background(), password.UpsertPasswordCredsOpts{UserID: uuid.New(), Hash: "h"})
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

	sentinel := errors.New("connection refused")
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "password_credentials"`).WillReturnError(sentinel)
	mock.ExpectRollback()

	_, err := r.Create(context.Background(), password.UpsertPasswordCredsOpts{UserID: uuid.New(), Hash: "h"})
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, original error no longer reachable", err)
	}

	if errors.Is(err, domain.ErrAlreadyExists) {
		t.Error("SQL failure must not be translated to ErrAlreadyExists")
	}
}

func TestGetByUserID(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	userID := uuid.New()
	updated := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(`SELECT \* FROM "password_credentials" WHERE user_id = \$1`).
		WithArgs(userID, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"user_id", "hash", "updated_at"}).
				AddRow(userID, "argon2$yyy", updated),
		)

	got, err := r.GetByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}

	want := password.PasswordCreds{UserID: userID, Hash: "argon2$yyy", UpdatedAt: updated}
	if got == nil || *got != want {
		t.Errorf("GetByUserID() = %+v, want %+v", got, want)
	}
}

func TestGetByUserIDNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectQuery(`SELECT \* FROM "password_credentials"`).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}))

	if _, err := r.GetByUserID(context.Background(), uuid.New()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByUserID = %v, want domain.ErrNotFound", err)
	}
}

func TestGetByUserIDError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("timeout")
	mock.ExpectQuery(`SELECT \* FROM "password_credentials"`).WillReturnError(sentinel)

	_, err := r.GetByUserID(context.Background(), uuid.New())
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v", err)
	}

	if errors.Is(err, domain.ErrNotFound) {
		t.Error("SQL failure must not be translated to ErrNotFound")
	}
}

func TestUpdateByUserID(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	userID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "password_credentials" SET "hash"=\$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	opts := password.UpsertPasswordCredsOpts{UserID: userID, Hash: newHash}
	if err := r.UpdateByUserID(context.Background(), opts); err != nil {
		t.Fatalf("UpdateByUserID: %v", err)
	}
}

func TestUpdateByUserIDNoRowAffected(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "password_credentials"`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := r.UpdateByUserID(context.Background(), password.UpsertPasswordCredsOpts{UserID: uuid.New(), Hash: newHash})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("UpdateByUserID = %v, want domain.ErrNotFound", err)
	}
}

func TestUpdateByUserIDError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("deadlock detected")
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "password_credentials"`).WillReturnError(sentinel)
	mock.ExpectRollback()

	err := r.UpdateByUserID(context.Background(), password.UpsertPasswordCredsOpts{UserID: uuid.New(), Hash: newHash})
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, original error no longer reachable", err)
	}
}
