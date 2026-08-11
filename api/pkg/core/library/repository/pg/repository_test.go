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

const (
	libraryComicSource = "asurascans"
	libraryComicSlug   = "solo-leveling"
	libraryComicTitle  = "Solo Leveling"
	libraryComicStatus = "completed"
	libraryComicType   = "manhwa"
)

func libraryEntryCols() []string {
	return []string{"id", "user_id", "comic_id", "added_at"}
}

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

func TestLibraryGetByID(t *testing.T) {
	t.Parallel()

	r, mock := newLibraryRepo(t)

	id := uuid.New()
	userID := uuid.New()
	comicID := uuid.New()
	addedAt := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(`SELECT \* FROM "library_entries" WHERE id = \$1`).
		WithArgs(id, 1).
		WillReturnRows(
			sqlmock.NewRows(libraryEntryCols()).
				AddRow(id, userID, comicID, addedAt),
		)

	got, err := r.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	want := library.Entry{ID: id, UserID: userID, ComicID: comicID, AddedAt: addedAt}
	if got == nil || *got != want {
		t.Errorf("GetByID() = %+v, want %+v", got, want)
	}
}

func TestLibraryGetByUserAndComic(t *testing.T) {
	t.Parallel()

	r, mock := newLibraryRepo(t)

	id := uuid.New()
	userID := uuid.New()
	comicID := uuid.New()
	addedAt := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(`SELECT \* FROM "library_entries" WHERE user_id = \$1 AND comic_id = \$2`).
		WithArgs(userID, comicID, 1).
		WillReturnRows(
			sqlmock.NewRows(libraryEntryCols()).
				AddRow(id, userID, comicID, addedAt),
		)

	got, err := r.GetByUserAndComic(context.Background(), userID, comicID)
	if err != nil {
		t.Fatalf("GetByUserAndComic: %v", err)
	}

	if got.ComicID != comicID || got.UserID != userID {
		t.Errorf("GetByUserAndComic() = %+v", got)
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

	got, err := r.Create(context.Background(), library.CreateEntryOpts{
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

	got, err := r.Create(context.Background(), library.CreateEntryOpts{
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

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "library_entries" WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := r.Delete(context.Background(), id)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestLibraryDeleteNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newLibraryRepo(t)

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "library_entries" WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := r.Delete(context.Background(), id)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete = %v, want domain.ErrNotFound", err)
	}
}

func TestLibraryListByUser(t *testing.T) {
	t.Parallel()

	r, mock := newLibraryRepo(t)

	userID := uuid.New()
	entryID := uuid.New()
	comicID := uuid.New()
	addedAt := time.Now().UTC().Truncate(time.Second)
	created := addedAt
	updated := addedAt

	mock.ExpectQuery(`SELECT \* FROM "library_entries" WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnRows(
			sqlmock.NewRows(libraryEntryCols()).
				AddRow(entryID, userID, comicID, addedAt),
		)

	mock.ExpectQuery(`SELECT \* FROM "comics" WHERE "comics"\."id" = \$1`).
		WithArgs(comicID).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "source", "slug", "title", "status", "comic_type",
				"chapter_count", "author", "artist", "description", "rating", "release_year",
				"source_url", "external_cover_url", "local_cover_path", "created_at", "updated_at",
			}).AddRow(
				comicID, libraryComicSource, libraryComicSlug, libraryComicTitle, libraryComicStatus, libraryComicType,
				200, "Chugong", "Dubu", "desc", 9.5, 2018,
				"", "", "", created, updated,
			),
		)

	got, err := r.ListByUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("ListByUser() len = %d, want 1", len(got))
	}

	if got[0].Entry.ID != entryID || got[0].Comic.ID != comicID {
		t.Errorf("ListByUser() = %+v", got[0])
	}
}
