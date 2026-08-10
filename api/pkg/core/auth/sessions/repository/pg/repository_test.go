// SPDX-License-Identifier: AGPL-3.0-or-later

package pg_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions/repository/pg"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgtest"
)

func newRepo(t *testing.T) (*pg.PGSessionsRepository, sqlmock.Sqlmock) {
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

func TestInsert(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "sessions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	userID := uuid.New()
	expires := time.Now().Add(time.Hour).UTC().Truncate(time.Second)

	got, err := r.Insert(context.Background(), sessions.InsertSessionOpts{
		UserID:     userID,
		AuthMethod: sessions.AuthMethodPassword,
		TokenHash:  []byte("0123456789abcdef0123456789abcdef"),
		ExpiresAt:  expires,
	})
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}

	if got.UserID != userID {
		t.Errorf("UserID = %v, want %v", got.UserID, userID)
	}

	if got.AuthMethod != sessions.AuthMethodPassword {
		t.Errorf("AuthMethod = %v, want %v", got.AuthMethod, sessions.AuthMethodPassword)
	}

	if !got.ExpiresAt.Equal(expires) {
		t.Errorf("ExpiresAt = %v, want %v", got.ExpiresAt, expires)
	}

	if got.ID == uuid.Nil {
		t.Error("Insert did not generate ID")
	}

	if got.CreatedAt.IsZero() {
		t.Error("Insert did not timestamp session")
	}
}

func TestInsertError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("disk full")
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "sessions"`).WillReturnError(sentinel)
	mock.ExpectRollback()

	got, err := r.Insert(context.Background(), sessions.InsertSessionOpts{UserID: uuid.New()})
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, original error no longer reachable", err)
	}

	if got != nil {
		t.Errorf("Insert returned %+v in addition to the error", got)
	}
}

func TestInsertAndGetByTokenHashRoundTripProviderIdentity(t *testing.T) {
	t.Parallel()

	providerID := uuid.New()
	providerSID := "provider-subject"

	tests := map[string]struct {
		providerID  *uuid.UUID
		providerSID *string
	}{
		"provider identity set":            {providerID: &providerID, providerSID: &providerSID},
		"plain password session (nil/nil)": {providerID: nil, providerSID: nil},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, mock := newRepo(t)

			mock.ExpectBegin()
			mock.ExpectQuery(`INSERT INTO "sessions"`).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
			mock.ExpectCommit()

			hash := []byte("0123456789abcdef0123456789abcdef")

			inserted, err := r.Insert(context.Background(), sessions.InsertSessionOpts{
				UserID:      uuid.New(),
				AuthMethod:  sessions.AuthMethodPassword,
				TokenHash:   hash,
				ExpiresAt:   time.Now().Add(time.Hour).UTC().Truncate(time.Second),
				ProviderID:  tc.providerID,
				ProviderSID: tc.providerSID,
			})
			if err != nil {
				t.Fatalf("Insert: %v", err)
			}

			assertProviderIDEqual(t, "Insert", inserted.ProviderID, tc.providerID)
			assertProviderSIDEqual(t, "Insert", inserted.ProviderSID, tc.providerSID)

			userID := uuid.New()
			created := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
			expires := time.Now().Add(time.Hour).UTC().Truncate(time.Second)

			var providerIDCol, providerSIDCol any
			if tc.providerID != nil {
				providerIDCol = *tc.providerID
			}

			if tc.providerSID != nil {
				providerSIDCol = *tc.providerSID
			}

			mock.ExpectQuery(`SELECT .* FROM "sessions" JOIN "users" "User" ON`).
				WithArgs(hash, 1).
				WillReturnRows(
					sqlmock.NewRows([]string{
						"id", "user_id", "token_hash", "auth_method", "created_at", "expires_at",
						"provider_id", "provider_sid",
						"User__id", "User__name", "User__is_admin", "User__created_at", "User__updated_at",
					}).AddRow(
						uuid.New(), userID, hash, "password", created, expires,
						providerIDCol, providerSIDCol,
						userID, "alice", true, created, created,
					),
				)

			got, _, err := r.GetByTokenHash(context.Background(), hash)
			if err != nil {
				t.Fatalf("GetByTokenHash: %v", err)
			}

			assertProviderIDEqual(t, "GetByTokenHash", got.ProviderID, tc.providerID)
			assertProviderSIDEqual(t, "GetByTokenHash", got.ProviderSID, tc.providerSID)
		})
	}
}

func assertProviderIDEqual(t *testing.T, op string, got, want *uuid.UUID) {
	t.Helper()

	switch {
	case want == nil:
		if got != nil {
			t.Errorf("%s: ProviderID = %v, want nil", op, *got)
		}
	case got == nil || *got != *want:
		t.Errorf("%s: ProviderID = %v, want %v", op, got, *want)
	}
}

func assertProviderSIDEqual(t *testing.T, op string, got, want *string) {
	t.Helper()

	switch {
	case want == nil:
		if got != nil {
			t.Errorf("%s: ProviderSID = %v, want nil", op, *got)
		}
	case got == nil || *got != *want:
		t.Errorf("%s: ProviderSID = %v, want %v", op, got, *want)
	}
}

func TestGetByTokenHashJoinsUserInOneQuery(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	id := uuid.New()
	userID := uuid.New()
	created := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	expires := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	hash := []byte("0123456789abcdef0123456789abcdef")

	mock.ExpectQuery(`SELECT .* FROM "sessions" JOIN "users" "User" ON`).
		WithArgs(hash, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "user_id", "token_hash", "auth_method", "created_at", "expires_at",
				"User__id", "User__name", "User__is_admin", "User__created_at", "User__updated_at",
			}).AddRow(
				id, userID, hash, "password", created, expires,
				userID, "alice", true, created, created,
			),
		)

	session, user, err := r.GetByTokenHash(context.Background(), hash)
	if err != nil {
		t.Fatalf("GetByTokenHash: %v", err)
	}

	if session.ID != id || session.UserID != userID {
		t.Errorf("session = %+v", session)
	}

	if session.AuthMethod != sessions.AuthMethodPassword {
		t.Errorf("AuthMethod = %v, want password", session.AuthMethod)
	}

	if !session.ExpiresAt.Equal(expires) {
		t.Errorf("ExpiresAt = %v, want %v", session.ExpiresAt, expires)
	}

	if user == nil || user.ID != userID || user.Name != "alice" || !user.IsAdmin {
		t.Errorf("user = %+v", user)
	}
}

func TestGetByTokenHashDoesNotFilterOnExpiry(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	hash := []byte("0123456789abcdef0123456789abcdef")
	mock.ExpectQuery(`SELECT`).
		WithArgs(hash, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

	if _, _, err := r.GetByTokenHash(context.Background(), hash); err != nil {
		t.Fatalf("GetByTokenHash: %v", err)
	}
}

func TestGetByTokenHashNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectQuery(`SELECT`).WillReturnRows(sqlmock.NewRows([]string{"id"}))

	session, user, err := r.GetByTokenHash(context.Background(), []byte("inconnu"))
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want domain.ErrNotFound", err)
	}

	if session != nil || user != nil {
		t.Errorf("GetByTokenHash returned (%+v, %+v) in addition to the error", session, user)
	}
}

func TestGetByTokenHashError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("connection reset")
	mock.ExpectQuery(`SELECT`).WillReturnError(sentinel)

	_, _, err := r.GetByTokenHash(context.Background(), []byte("x"))
	if errors.Is(err, domain.ErrNotFound) {
		t.Error("SQL failure must not be translated to ErrNotFound")
	}

	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, original error no longer reachable", err)
	}
}

func TestUpdateExpiry(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	id := uuid.New()
	expires := time.Now().Add(30 * 24 * time.Hour).UTC().Truncate(time.Second)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "sessions" SET "expires_at"=\$1 WHERE id = \$2`).
		WithArgs(expires, id).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := r.UpdateExpiry(context.Background(), id, expires); err != nil {
		t.Fatalf("UpdateExpiry: %v", err)
	}
}

func TestUpdateExpiryError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("deadlock detected")
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "sessions"`).WillReturnError(sentinel)
	mock.ExpectRollback()

	err := r.UpdateExpiry(context.Background(), uuid.New(), time.Now())
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, original error no longer reachable", err)
	}
}

func TestDeleteByTokenHashIgnoresMissingRow(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	hash := []byte("0123456789abcdef0123456789abcdef")

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "sessions" WHERE token_hash = \$1`).
		WithArgs(hash).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	if err := r.DeleteByTokenHash(context.Background(), hash); err != nil {
		t.Fatalf("DeleteByTokenHash: %v", err)
	}
}

func TestDeleteByUserID(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "sessions" WHERE user_id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	if err := r.DeleteByUserID(context.Background(), id); err != nil {
		t.Fatalf("DeleteByUserID: %v", err)
	}
}

func TestDeleteByUserAndProvider(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	userID, providerID := uuid.New(), uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "sessions" WHERE user_id = \$1 AND provider_id = \$2`).
		WithArgs(userID, providerID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	if err := r.DeleteByUserAndProvider(context.Background(), userID, providerID); err != nil {
		t.Fatalf("DeleteByUserAndProvider: %v", err)
	}
}

func TestDeleteExpiredReturnsRowCount(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	now := time.Now().UTC().Truncate(time.Second)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "sessions" WHERE expires_at <= \$1`).
		WithArgs(now).
		WillReturnResult(sqlmock.NewResult(0, 12))
	mock.ExpectCommit()

	got, err := r.DeleteExpired(context.Background(), now)
	if err != nil {
		t.Fatalf("DeleteExpired: %v", err)
	}

	if got != 12 {
		t.Errorf("DeleteExpired() = %d, want 12", got)
	}
}

func TestDeleteExpiredError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("connection reset")
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "sessions"`).WillReturnError(sentinel)
	mock.ExpectRollback()

	got, err := r.DeleteExpired(context.Background(), time.Now())
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, original error no longer reachable", err)
	}

	if got != 0 {
		t.Errorf("DeleteExpired() = %d, want 0 on error", got)
	}
}
