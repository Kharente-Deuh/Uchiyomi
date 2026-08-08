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
		t.Fatal("New sans DB doit échouer")
	}

	if r != nil {
		t.Error("New a renvoyé un repository en plus de l'erreur")
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
		t.Error("Insert n'a pas généré d'ID")
	}

	if got.CreatedAt.IsZero() {
		t.Error("Insert n'a pas horodaté la session")
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
		t.Errorf("err = %v, l'erreur d'origine n'est plus atteignable", err)
	}

	if got != nil {
		t.Errorf("Insert a renvoyé %+v en plus de l'erreur", got)
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
		t.Errorf("GetByTokenHash a renvoyé (%+v, %+v) en plus de l'erreur", session, user)
	}
}

func TestGetByTokenHashError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("connection reset")
	mock.ExpectQuery(`SELECT`).WillReturnError(sentinel)

	_, _, err := r.GetByTokenHash(context.Background(), []byte("x"))
	if errors.Is(err, domain.ErrNotFound) {
		t.Error("une panne SQL ne doit pas être traduite en ErrNotFound")
	}

	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, l'erreur d'origine n'est plus atteignable", err)
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
		t.Errorf("err = %v, l'erreur d'origine n'est plus atteignable", err)
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
		t.Errorf("err = %v, l'erreur d'origine n'est plus atteignable", err)
	}

	if got != 0 {
		t.Errorf("DeleteExpired() = %d, want 0 en cas d'erreur", got)
	}
}
