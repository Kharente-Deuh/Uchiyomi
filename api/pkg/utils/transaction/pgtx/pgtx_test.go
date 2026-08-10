// SPDX-License-Identifier: AGPL-3.0-or-later

package pgtx_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgmodels"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgtest"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
	"gorm.io/gorm"
)

func serializationFailure() error {
	return &pgconn.PgError{Code: "40001", Message: "could not serialize access due to read/write dependencies"}
}

func newTransactor(t *testing.T) (*pgtx.PGTransactor, *gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock := pgtest.New(t)

	tr, err := pgtx.New(pgtx.Deps{DB: db})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return tr, db, mock
}

func TestNewValidatesDeps(t *testing.T) {
	t.Parallel()

	tr, err := pgtx.New(pgtx.Deps{})
	if err == nil {
		t.Fatal("New without DB must fail")
	}

	if tr != nil {
		t.Error("New returned a transactor in addition to the error")
	}

	if want := "deps.Validate: db is required"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
}

func TestFromWithoutTransactionReturnsRoot(t *testing.T) {
	t.Parallel()

	db, _ := pgtest.New(t)

	if got := pgtx.From(context.Background(), db); got != db {
		t.Errorf("From() = %p, want la racine %p", got, db)
	}
}

func TestWithinTxCommitsOnSuccess(t *testing.T) {
	t.Parallel()

	tr, _, mock := newTransactor(t)

	mock.ExpectBegin()
	mock.ExpectCommit()

	err := tr.WithinTx(context.Background(), transaction.TxOpts{}, func(context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("WithinTx: %v", err)
	}
}

func TestWithinTxRollsBackAndPropagatesError(t *testing.T) {
	t.Parallel()

	tr, _, mock := newTransactor(t)

	mock.ExpectBegin()
	mock.ExpectRollback()

	sentinel := errors.New("business rule violated")

	err := tr.WithinTx(context.Background(), transaction.TxOpts{}, func(context.Context) error {
		return sentinel
	})
	if err == nil {
		t.Fatal("WithinTx must propagate callback error")
	}

	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, original error no longer reachable via errors.Is", err)
	}
}

func TestWithinTxPutsTransactionInContext(t *testing.T) {
	t.Parallel()

	tr, db, mock := newTransactor(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT count\(\*\) FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectCommit()

	err := tr.WithinTx(context.Background(), transaction.TxOpts{}, func(ctx context.Context) error {
		inTx := pgtx.From(ctx, db)
		if inTx == db {
			t.Error("callback ctx does not carry transaction")
		}

		if _, err := gorm.G[pgmodels.User](inTx).Count(ctx, "*"); err != nil {
			return fmt.Errorf("Count: %w", err)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("WithinTx: %v", err)
	}
}

func TestWithinTxRetriesOnSerializationFailure(t *testing.T) {
	t.Parallel()

	tr, _, mock := newTransactor(t)

	for range 2 {
		mock.ExpectBegin()
		mock.ExpectRollback()
	}

	mock.ExpectBegin()
	mock.ExpectCommit()

	calls := 0

	err := tr.WithinTx(context.Background(), transaction.TxOpts{Isolation: transaction.IsolationSerializable},
		func(context.Context) error {
			calls++
			if calls < 3 {
				return serializationFailure()
			}

			return nil
		})
	if err != nil {
		t.Fatalf("WithinTx: %v", err)
	}

	if calls != 3 {
		t.Errorf("callback called %d times, want 3", calls)
	}
}

func TestWithinTxGivesUpAfterMaxAttempts(t *testing.T) {
	t.Parallel()

	tr, _, mock := newTransactor(t)

	for range pgtx.MaxAttempts {
		mock.ExpectBegin()
		mock.ExpectRollback()
	}

	calls := 0

	err := tr.WithinTx(context.Background(), transaction.TxOpts{Isolation: transaction.IsolationSerializable},
		func(context.Context) error {
			calls++

			return serializationFailure()
		})
	if err == nil {
		t.Fatal("WithinTx must give up and propagate serialization failure")
	}

	if calls != pgtx.MaxAttempts {
		t.Errorf("callback called %d times, want %d", calls, pgtx.MaxAttempts)
	}
}

func TestWithinTxDoesNotRetryOtherErrors(t *testing.T) {
	t.Parallel()

	tr, _, mock := newTransactor(t)

	mock.ExpectBegin()
	mock.ExpectRollback()

	calls := 0

	err := tr.WithinTx(context.Background(), transaction.TxOpts{Isolation: transaction.IsolationSerializable},
		func(context.Context) error {
			calls++

			return &pgconn.PgError{Code: "23505", Message: "duplicate key"}
		})
	if err == nil {
		t.Fatal("WithinTx must propagate error")
	}

	if calls != 1 {
		t.Errorf("callback called %d times, want 1 (no retry outside 40001)", calls)
	}
}
