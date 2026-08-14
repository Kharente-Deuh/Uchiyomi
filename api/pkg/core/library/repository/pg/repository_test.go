// SPDX-License-Identifier: AGPL-3.0-or-later

package pg_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/library"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/library/repository/pg"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgtest"
)

func duplicateLibraryEntryKeyErr() error {
	return &pgconn.PgError{
		Code:    "23505",
		Message: `duplicate key value violates unique constraint "idx_library_entry_user_comic"`,
	}
}

func newLibraryRepo(t *testing.T) (*pg.PGLibraryRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock := pgtest.New(t)

	r, err := pg.New(pg.Deps{DB: db})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return r, mock
}

func TestLibraryNewValidatesDeps(t *testing.T) {
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

func TestLibraryCreate(t *testing.T) {
	t.Parallel()

	r, mock := newLibraryRepo(t)

	userID := uuid.New()
	comicID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "library_entries"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	before := time.Now()

	got, err := r.Create(context.Background(), library.CreateOpts{
		UserID:  userID,
		ComicID: comicID,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if got.UserID != userID || got.ComicID != comicID {
		t.Errorf("Create() = %+v", got)
	}

	if got.ID == uuid.Nil {
		t.Error("Create did not generate ID")
	}

	if got.AddedAt.Before(before) {
		t.Errorf("added_at not set: %v", got.AddedAt)
	}
}

func TestLibraryCreateDuplicateKey(t *testing.T) {
	t.Parallel()

	r, mock := newLibraryRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "library_entries"`).WillReturnError(duplicateLibraryEntryKeyErr())
	mock.ExpectRollback()

	got, err := r.Create(context.Background(), library.CreateOpts{
		UserID:  uuid.New(),
		ComicID: uuid.New(),
	})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("Create = %v, want domain.ErrAlreadyExists", err)
	}

	if got != nil {
		t.Errorf("Create returned %+v in addition to the error", got)
	}
}

func TestLibraryDelete(t *testing.T) {
	t.Parallel()

	r, mock := newLibraryRepo(t)

	userID := uuid.New()
	comicID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "library_entries" WHERE user_id = \$1 AND comic_id = \$2`).
		WithArgs(userID, comicID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := r.Delete(context.Background(), library.DeleteOpts{
		UserID:  userID,
		ComicID: comicID,
	})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestLibraryDeleteNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newLibraryRepo(t)

	userID := uuid.New()
	comicID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "library_entries" WHERE user_id = \$1 AND comic_id = \$2`).
		WithArgs(userID, comicID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := r.Delete(context.Background(), library.DeleteOpts{
		UserID:  userID,
		ComicID: comicID,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete = %v, want domain.ErrNotFound", err)
	}
}

func TestLibraryExistsByComicID(t *testing.T) {
	t.Parallel()

	r, mock := newLibraryRepo(t)

	comicID := uuid.New()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "library_entries" WHERE comic_id = \$1`).
		WithArgs(comicID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	exists, err := r.ExistsByComicID(context.Background(), comicID)
	if err != nil {
		t.Fatalf("ExistsByComicID: %v", err)
	}

	if !exists {
		t.Error("ExistsByComicID = false, want true")
	}
}

func TestLibraryExistsByComicIDNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newLibraryRepo(t)

	comicID := uuid.New()

	mock.ExpectQuery(`SELECT count`).
		WithArgs(comicID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	exists, err := r.ExistsByComicID(context.Background(), comicID)
	if err != nil {
		t.Fatalf("ExistsByComicID: %v", err)
	}

	if exists {
		t.Error("ExistsByComicID = true, want false")
	}
}
