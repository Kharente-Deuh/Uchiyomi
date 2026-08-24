// SPDX-License-Identifier: AGPL-3.0-or-later

//nolint:goconst
package pg_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readingprogress"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readingprogress/repository/pg"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgtest"
)

func newReadingProgressRepo(t *testing.T) (*pg.PGReadingProgressRepository, sqlmock.Sqlmock) {
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

func readingProgressSelectColumns() []string {
	return []string{"id", "library_entry_id", "chapter_id", "page", "updated_at"}
}

func TestGetLatestByUserAndComicJoinsLibraryEntry(t *testing.T) {
	t.Parallel()

	r, mock := newReadingProgressRepo(t)

	userID := uuid.New()
	otherUserID := uuid.New()
	comicID := uuid.New()
	chapterID := uuid.New()
	updatedAt := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`JOIN "library_entries" "LibraryEntry".*WHERE "LibraryEntry"\.(user_id|"user_id")`).
		WithArgs(userID, comicID, 1).
		WillReturnRows(
			sqlmock.NewRows(readingProgressSelectColumns()).AddRow(
				uuid.New(), uuid.New(), chapterID, 5, updatedAt,
			),
		)

	got, err := r.GetLatestByUserAndComic(context.Background(), readingprogress.ListOpts{
		UserID:  userID,
		ComicID: comicID,
	})
	if err != nil {
		t.Fatalf("GetLatestByUserAndComic: %v", err)
	}

	if got == nil {
		t.Fatal("GetLatestByUserAndComic() = nil, want progress")
	}

	want := readingprogress.Progress{
		ChapterID: chapterID,
		Page:      5,
		UpdatedAt: updatedAt,
	}
	if *got != want {
		t.Errorf("GetLatestByUserAndComic() = %+v, want %+v", *got, want)
	}

	if otherUserID == userID {
		t.Fatal("unlucky uuid collision")
	}
}

func TestGetLatestOrdersByUpdatedAtDesc(t *testing.T) {
	t.Parallel()

	r, mock := newReadingProgressRepo(t)

	userID := uuid.New()
	comicID := uuid.New()

	mock.ExpectQuery(`ORDER BY.*updated_at`).
		WithArgs(userID, comicID, 1).
		WillReturnRows(sqlmock.NewRows(readingProgressSelectColumns()))

	got, err := r.GetLatestByUserAndComic(context.Background(), readingprogress.ListOpts{
		UserID:  userID,
		ComicID: comicID,
	})
	if err != nil {
		t.Fatalf("GetLatestByUserAndComic: %v", err)
	}

	if got != nil {
		t.Errorf("GetLatestByUserAndComic() = %+v, want nil", got)
	}
}

func TestListByUserAndChapterIDsEmptySkipsQuery(t *testing.T) {
	t.Parallel()

	r, _ := newReadingProgressRepo(t)

	got, err := r.ListByUserAndChapterIDs(context.Background(), readingprogress.MapOpts{
		UserID: uuid.New(),
		IDs:    nil,
	})
	if err != nil {
		t.Fatalf("ListByUserAndChapterIDs: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("ListByUserAndChapterIDs() = %+v, want empty", got)
	}
}

func TestListByUserAndChapterIDsJoinsLibraryEntry(t *testing.T) {
	t.Parallel()

	r, mock := newReadingProgressRepo(t)

	userID := uuid.New()
	chapterID := uuid.New()
	updatedAt := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`JOIN "library_entries" "LibraryEntry".*chapter_id`).
		WithArgs(userID, chapterID).
		WillReturnRows(
			sqlmock.NewRows(readingProgressSelectColumns()).AddRow(
				uuid.New(), uuid.New(), chapterID, 9, updatedAt,
			),
		)

	got, err := r.ListByUserAndChapterIDs(context.Background(), readingprogress.MapOpts{
		UserID: userID,
		IDs:    []uuid.UUID{chapterID},
	})
	if err != nil {
		t.Fatalf("ListByUserAndChapterIDs: %v", err)
	}

	if len(got) != 1 || got[0].ChapterID != chapterID || got[0].Page != 9 {
		t.Errorf("ListByUserAndChapterIDs() = %+v", got)
	}
}

func TestGetNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newReadingProgressRepo(t)

	userID := uuid.New()
	comicID := uuid.New()
	chapterID := uuid.New()

	mock.ExpectQuery(`JOIN "library_entries" "LibraryEntry".*WHERE "LibraryEntry"\.(user_id|"user_id")`).
		WithArgs(userID, comicID, chapterID, 1).
		WillReturnRows(sqlmock.NewRows(readingProgressSelectColumns()))

	got, err := r.Get(context.Background(), readingprogress.GetOpts{
		UserID:    userID,
		ComicID:   comicID,
		ChapterID: chapterID,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Get = %v, want domain.ErrNotFound", err)
	}

	if got != nil {
		t.Errorf("Get returned %+v in addition to the error", got)
	}
}

func TestUpsertInserts(t *testing.T) {
	t.Parallel()

	r, mock := newReadingProgressRepo(t)

	userID := uuid.New()
	comicID := uuid.New()
	chapterID := uuid.New()
	entryID := uuid.New()
	updatedAt := time.Date(2024, 6, 2, 8, 30, 0, 0, time.UTC)
	page := 12

	mock.ExpectQuery(`FROM "library_entries".*user_id`).
		WithArgs(userID, comicID, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "user_id", "comic_id", "added_at"}).
				AddRow(entryID, userID, comicID, time.Now()),
		)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "reading_progress"`).
		WithArgs(
			updatedAt,
			entryID,
			chapterID,
			page,
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	got, err := r.Upsert(context.Background(), readingprogress.UpsertOpts{
		UserID:    userID,
		ComicID:   comicID,
		ChapterID: chapterID,
		Page:      page,
		UpdatedAt: updatedAt,
	})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	want := readingprogress.Progress{
		ChapterID: chapterID,
		Page:      page,
		UpdatedAt: updatedAt,
	}
	if got != want {
		t.Errorf("Upsert() = %+v, want %+v", got, want)
	}
}

func TestUpsertConflictUpdates(t *testing.T) {
	t.Parallel()

	r, mock := newReadingProgressRepo(t)

	userID := uuid.New()
	comicID := uuid.New()
	chapterID := uuid.New()
	entryID := uuid.New()
	updatedAt := time.Date(2024, 6, 3, 9, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`FROM "library_entries".*user_id`).
		WithArgs(userID, comicID, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "user_id", "comic_id", "added_at"}).
				AddRow(entryID, userID, comicID, time.Now()),
		)

	mock.ExpectBegin()
	mock.ExpectQuery(`ON CONFLICT`).
		WithArgs(
			updatedAt,
			entryID,
			chapterID,
			7,
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	got, err := r.Upsert(context.Background(), readingprogress.UpsertOpts{
		UserID:    userID,
		ComicID:   comicID,
		ChapterID: chapterID,
		Page:      7,
		UpdatedAt: updatedAt,
	})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	want := readingprogress.Progress{
		ChapterID: chapterID,
		Page:      7,
		UpdatedAt: updatedAt,
	}
	if got != want {
		t.Errorf("Upsert() = %+v, want %+v", got, want)
	}
}

func TestDeleteByUserAndChapterIDs(t *testing.T) {
	t.Parallel()

	r, mock := newReadingProgressRepo(t)
	userID := uuid.New()
	ch1 := uuid.New()
	ch2 := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(
		`DELETE FROM "reading_progress" WHERE chapter_id IN \(\$1,\$2\) AND `+
			`library_entry_id IN \(SELECT (id|"id") FROM "library_entries" WHERE user_id = \$3\)`,
	).
		WithArgs(ch1, ch2, userID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	err := r.DeleteByUserAndChapterIDs(context.Background(), readingprogress.DeleteProgressOpts{
		UserID:     userID,
		ChapterIDs: []uuid.UUID{ch1, ch2},
	})
	if err != nil {
		t.Fatalf("DeleteByUserAndChapterIDs: %v", err)
	}
}

func TestDeleteByUserAndChapterIDsEmpty(t *testing.T) {
	t.Parallel()

	r, _ := newReadingProgressRepo(t)

	err := r.DeleteByUserAndChapterIDs(context.Background(), readingprogress.DeleteProgressOpts{
		UserID:     uuid.New(),
		ChapterIDs: []uuid.UUID{},
	})
	if err != nil {
		t.Fatalf("DeleteByUserAndChapterIDs with empty IDs should return nil, got: %v", err)
	}
}
