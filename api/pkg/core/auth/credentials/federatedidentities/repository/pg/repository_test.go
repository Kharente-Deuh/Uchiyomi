// SPDX-License-Identifier: AGPL-3.0-or-later

package pg_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/federatedidentities"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/credentials/federatedidentities/repository/pg"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgtest"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
)

const (
	testSubject = "sub-123"
	testEmail   = "bob@example.com"
	claimsEmail = "email"
)

func duplicateKeyErr() error {
	return &pgconn.PgError{
		Code:    "23505",
		Message: `duplicate key value violates unique constraint "idx_fedid_provider_subject"`,
	}
}

func newRepo(t *testing.T) (*pg.PGFederatedIdentitiesRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock := pgtest.New(t)

	r, err := pg.New(pg.Deps{DB: db})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return r, mock
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
	mock.ExpectQuery(`INSERT INTO "federated_identities"`).
		WillReturnRows(sqlmock.NewRows([]string{"claims", "id"}).AddRow(nil, uuid.New()))
	mock.ExpectCommit()

	err = tr.WithinTx(context.Background(), transaction.TxOpts{}, func(ctx context.Context) error {
		_, createErr := r.Create(ctx, federatedidentities.CreateFederatedIdentityOpts{
			UserID:     uuid.New(),
			ProviderID: uuid.New(),
			Subject:    "sub-1",
		})

		if createErr != nil {
			return fmt.Errorf("r.Create: %w", createErr)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("WithinTx: %v", err)
	}
}

func TestNewValidatesDeps(t *testing.T) {
	t.Parallel()

	if _, err := pg.New(pg.Deps{}); err == nil || err.Error() != "deps.Validate: db is required" {
		t.Errorf("pg.New(pg.Deps{}) = %v, want %q", err, "deps.Validate: db is required")
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	userID, providerID := uuid.New(), uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "federated_identities"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	before := time.Now()

	got, err := r.Create(context.Background(), federatedidentities.CreateFederatedIdentityOpts{
		UserID:     userID,
		ProviderID: providerID,
		Subject:    testSubject,
		Claims:     map[string]any{claimsEmail: testEmail},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if got.UserID != userID || got.ProviderID != providerID || got.Subject != testSubject {
		t.Errorf("Create() = %+v", got)
	}

	if got.ID == uuid.Nil {
		t.Error("Create did not generate ID")
	}

	if !reflect.DeepEqual(got.Claims, map[string]any{claimsEmail: testEmail}) {
		t.Errorf("Claims = %v", got.Claims)
	}

	if got.CreatedAt.Before(before) || got.LastLoginAt.Before(before) {
		t.Errorf("timestamps not set: created=%v lastLogin=%v", got.CreatedAt, got.LastLoginAt)
	}
}

func TestCreateDuplicate(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "federated_identities"`).WillReturnError(duplicateKeyErr())
	mock.ExpectRollback()

	got, err := r.Create(context.Background(), federatedidentities.CreateFederatedIdentityOpts{
		UserID:     uuid.New(),
		ProviderID: uuid.New(),
		Subject:    testSubject,
	})
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
	mock.ExpectQuery(`INSERT INTO "federated_identities"`).WillReturnError(sentinel)
	mock.ExpectRollback()

	_, err := r.Create(context.Background(), federatedidentities.CreateFederatedIdentityOpts{})
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v", err)
	}

	if errors.Is(err, domain.ErrAlreadyExists) {
		t.Error("SQL failure must not be translated to ErrAlreadyExists")
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	id, userID, providerID := uuid.New(), uuid.New(), uuid.New()
	created := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	lastLogin := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(`SELECT \* FROM "federated_identities" WHERE provider_id = \$1 AND subject = \$2`).
		WithArgs(providerID, testSubject, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "user_id", "provider_id", "subject", "claims", "created_at", "last_login_at"}).
				AddRow(id, userID, providerID, testSubject, []byte(`{"email":"bob@example.com"}`), created, lastLogin),
		)

	got, err := r.Get(context.Background(), federatedidentities.GetFederatedIdentityOpts{
		ProviderID: providerID,
		Subject:    testSubject,
	})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got.ID != id || got.UserID != userID || got.ProviderID != providerID {
		t.Errorf("Get() = %+v", got)
	}

	if !reflect.DeepEqual(got.Claims, map[string]any{claimsEmail: testEmail}) {
		t.Errorf("Claims = %v", got.Claims)
	}

	if !got.CreatedAt.Equal(created) || !got.LastLoginAt.Equal(lastLogin) {
		t.Errorf("horodatages = created:%v lastLogin:%v", got.CreatedAt, got.LastLoginAt)
	}
}

func TestGetNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectQuery(`SELECT \* FROM "federated_identities"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := r.Get(context.Background(), federatedidentities.GetFederatedIdentityOpts{
		ProviderID: uuid.New(),
		Subject:    "inconnu",
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get = %v, want domain.ErrNotFound", err)
	}
}

func TestGetError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("timeout")
	mock.ExpectQuery(`SELECT \* FROM "federated_identities"`).WillReturnError(sentinel)

	_, err := r.Get(context.Background(), federatedidentities.GetFederatedIdentityOpts{})
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v", err)
	}

	if errors.Is(err, domain.ErrNotFound) {
		t.Error("SQL failure must not be translated to ErrNotFound")
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "federated_identities" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := r.Update(context.Background(), federatedidentities.UpdateFederatedIdentityOpts{
		ID:          uuid.New(),
		Claims:      map[string]any{claimsEmail: "new@example.com"},
		LastLoginAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestUpdateNoRowAffected(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "federated_identities"`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := r.Update(context.Background(), federatedidentities.UpdateFederatedIdentityOpts{
		ID:          uuid.New(),
		Claims:      map[string]any{"a": "b"},
		LastLoginAt: time.Now(),
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Update = %v, want domain.ErrNotFound", err)
	}
}

func TestUpdateError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("deadlock detected")
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "federated_identities"`).WillReturnError(sentinel)
	mock.ExpectRollback()

	err := r.Update(context.Background(), federatedidentities.UpdateFederatedIdentityOpts{
		ID:          uuid.New(),
		Claims:      map[string]any{"a": "b"},
		LastLoginAt: time.Now(),
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, original error no longer reachable", err)
	}
}
