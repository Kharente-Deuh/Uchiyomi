// SPDX-License-Identifier: AGPL-3.0-or-later

//nolint:goconst,lll
package pg_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters/repository/pg"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgtest"
)

const (
	chapterSlug  = "chapter-1"
	chapterTitle = "Chapter 1"
)

func duplicateChapterKeyErr() error {
	return &pgconn.PgError{
		Code:    "23505",
		Message: `duplicate key value violates unique constraint "idx_chapter_comic_source_slug"`,
	}
}

func newChaptersRepo(t *testing.T) (*pg.PGChaptersRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock := pgtest.New(t)

	r, err := pg.New(pg.Deps{DB: db})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return r, mock
}

func TestChaptersNewValidatesDeps(t *testing.T) {
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

func TestChaptersCreate(t *testing.T) {
	t.Parallel()

	r, mock := newChaptersRepo(t)

	comicID := uuid.New()
	publishedAt := time.Now().Add(-24 * time.Hour).UTC().Truncate(time.Second)
	earlyAccessUntil := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "chapters"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	got, err := r.Create(context.Background(), chapters.CreateOpts{
		ComicID:           comicID,
		SourceChapterSlug: chapterSlug,
		Number:            1,
		Title:             chapterTitle,
		PagesNb:           42,
		PublishedAt:       publishedAt,
		EarlyAccessUntil:  &earlyAccessUntil,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if got.ComicID != comicID || got.SourceChapterSlug != chapterSlug {
		t.Errorf("Create() = %+v", got)
	}

	if got.Number != 1 || got.Title != chapterTitle || got.PagesNb != 42 {
		t.Errorf("Create() = %+v", got)
	}

	if got.PublishedAt != publishedAt ||
		got.EarlyAccessUntil == nil || !got.EarlyAccessUntil.Equal(earlyAccessUntil) {
		t.Errorf("Create() timestamps = published %v early %v", got.PublishedAt, got.EarlyAccessUntil)
	}

	if got.ID == uuid.Nil {
		t.Error("Create did not generate ID")
	}

	if got.Download != 0 {
		t.Errorf("Create() download = %d, want 0", got.Download)
	}
}

func TestChaptersCreateDuplicateKey(t *testing.T) {
	t.Parallel()

	r, mock := newChaptersRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "chapters"`).WillReturnError(duplicateChapterKeyErr())
	mock.ExpectRollback()

	got, err := r.Create(context.Background(), chapters.CreateOpts{
		ComicID:           uuid.New(),
		SourceChapterSlug: chapterSlug,
		Number:            1,
		Title:             chapterTitle,
		PublishedAt:       time.Now(),
	})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("Create = %v, want domain.ErrAlreadyExists", err)
	}

	if got != nil {
		t.Errorf("Create returned %+v in addition to the error", got)
	}
}

func TestChaptersCreateMany(t *testing.T) {
	t.Parallel()

	r, mock := newChaptersRepo(t)

	comicID := uuid.New()
	publishedAt := time.Now().Add(-24 * time.Hour).UTC().Truncate(time.Second)
	earlyAccessUntil := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "chapters"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()).AddRow(uuid.New()))
	mock.ExpectCommit()

	got, err := r.CreateMany(context.Background(), []chapters.CreateOpts{
		{
			ComicID:           comicID,
			SourceChapterSlug: chapterSlug,
			Number:            1,
			Title:             chapterTitle,
			PagesNb:           42,
			PublishedAt:       publishedAt,
			EarlyAccessUntil:  &earlyAccessUntil,
		},
		{
			ComicID:           comicID,
			SourceChapterSlug: "chapter-2",
			Number:            2,
			Title:             "Chapter 2",
			PagesNb:           30,
			PublishedAt:       publishedAt,
			EarlyAccessUntil:  &earlyAccessUntil,
		},
	})
	if err != nil {
		t.Fatalf("CreateMany: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len(CreateMany()) = %d, want 2", len(got))
	}

	if got[0].ComicID != comicID || got[0].SourceChapterSlug != chapterSlug || got[0].PagesNb != 42 {
		t.Errorf("CreateMany()[0] = %+v", got[0])
	}

	if got[1].Number != 2 || got[1].PagesNb != 30 {
		t.Errorf("CreateMany()[1] = %+v", got[1])
	}
}

func TestChaptersListByComicID(t *testing.T) {
	t.Parallel()

	r, mock := newChaptersRepo(t)

	comicID := uuid.New()
	id := uuid.New()
	publishedAt := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(`FROM "chapters".*comic_id = \$1`).
		WithArgs(comicID).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "comic_id", "source_chapter_slug", "number", "title",
				"pages_nb", "published_at", "early_access_until", "download",
			}).AddRow(
				id, comicID, chapterSlug, 1.0, chapterTitle,
				42, publishedAt, publishedAt, 0,
			),
		)

	got, err := r.ListByComicID(context.Background(), comicID)
	if err != nil {
		t.Fatalf("ListByComicID: %v", err)
	}

	if len(got) != 1 || got[0].ID != id || got[0].ComicID != comicID {
		t.Errorf("ListByComicID() = %+v", got)
	}
}

func TestChaptersListResumable(t *testing.T) {
	t.Parallel()

	r, mock := newChaptersRepo(t)

	comicID := uuid.New()
	id := uuid.New()
	publishedAt := time.Now().UTC().Truncate(time.Second)
	now := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(`FROM "chapters".*\(download > 0 AND download < 100\) OR download = -1 OR \(download = 0 AND \(early_access_until IS NULL OR early_access_until <= \$1\)\)`).
		WithArgs(now).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "comic_id", "source_chapter_slug", "number", "title",
				"pages_nb", "published_at", "early_access_until", "download",
			}).AddRow(
				id, comicID, chapterSlug, 1.0, chapterTitle,
				42, publishedAt, nil, 0,
			),
		)

	got, err := r.ListResumable(context.Background(), now)
	if err != nil {
		t.Fatalf("ListResumable: %v", err)
	}

	if len(got) != 1 || got[0].ID != id || got[0].Download != 0 {
		t.Errorf("ListResumable() = %+v", got)
	}
}

func TestChaptersListEarlyAccessUnlocked(t *testing.T) {
	t.Parallel()

	r, mock := newChaptersRepo(t)

	comicID := uuid.New()
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(`FROM "chapters".*download = 0 AND early_access_until <= \$1`).
		WithArgs(now).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "comic_id", "source_chapter_slug", "number", "title",
				"pages_nb", "published_at", "early_access_until", "download",
			}).AddRow(
				id, comicID, chapterSlug, 1.0, chapterTitle,
				42, now, now.Add(-time.Hour), 0,
			),
		)

	got, err := r.ListEarlyAccessUnlocked(context.Background(), now)
	if err != nil {
		t.Fatalf("ListEarlyAccessUnlocked: %v", err)
	}

	if len(got) != 1 || got[0].ID != id || got[0].Download != 0 {
		t.Errorf("ListEarlyAccessUnlocked() = %+v", got)
	}
}

func chapterSelectColumns() []string {
	return []string{
		"id", "comic_id", "source_chapter_slug", "number", "title",
		"pages_nb", "published_at", "early_access_until", "download",
	}
}

func TestChaptersGetByID(t *testing.T) {
	t.Parallel()

	r, mock := newChaptersRepo(t)

	comicID := uuid.New()
	id := uuid.New()
	publishedAt := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(`FROM "chapters".*id = \$1`).
		WithArgs(id, 1).
		WillReturnRows(
			sqlmock.NewRows(chapterSelectColumns()).AddRow(
				id, comicID, chapterSlug, 1.0, chapterTitle,
				42, publishedAt, publishedAt, 80,
			),
		)

	got, err := r.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.ID != id || got.ComicID != comicID || got.Download != 80 || got.PagesNb != 42 {
		t.Errorf("GetByID() = %+v", got)
	}
}

func TestChaptersGetByIDNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newChaptersRepo(t)

	mock.ExpectQuery(`FROM "chapters".*id = \$1`).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnRows(sqlmock.NewRows(chapterSelectColumns()))

	got, err := r.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID = %v, want domain.ErrNotFound", err)
	}

	if got != nil {
		t.Errorf("GetByID returned %+v in addition to the error", got)
	}
}

func TestChaptersUpdateDownload(t *testing.T) {
	t.Parallel()

	r, mock := newChaptersRepo(t)
	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "chapters" SET`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := r.UpdateDownload(context.Background(), id, 42); err != nil {
		t.Fatalf("UpdateDownload: %v", err)
	}
}

func TestChaptersUpdateDownloadNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newChaptersRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "chapters" SET`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := r.UpdateDownload(context.Background(), uuid.New(), 42)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("UpdateDownload = %v, want domain.ErrNotFound", err)
	}
}

func TestChaptersUpdatePagesNb(t *testing.T) {
	t.Parallel()

	r, mock := newChaptersRepo(t)
	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "chapters" SET`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := r.UpdatePagesNb(context.Background(), id, 30); err != nil {
		t.Fatalf("UpdatePagesNb: %v", err)
	}
}

func TestChaptersUpdatePagesNbNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newChaptersRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "chapters" SET`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := r.UpdatePagesNb(context.Background(), uuid.New(), 30)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("UpdatePagesNb = %v, want domain.ErrNotFound", err)
	}
}

func TestChaptersGetByIds(t *testing.T) {
	t.Parallel()

	r, mock := newChaptersRepo(t)

	comicID := uuid.New()
	id1 := uuid.New()
	id2 := uuid.New()
	publishedAt := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(`FROM "chapters".*id IN`).
		WithArgs(id1, id2).
		WillReturnRows(
			sqlmock.NewRows(chapterSelectColumns()).
				AddRow(id1, comicID, chapterSlug, 1.0, chapterTitle, 42, publishedAt, publishedAt, 0).
				AddRow(id2, comicID, "chapter-2", 2.0, "Chapter 2", 30, publishedAt, publishedAt, 80),
		)

	got, err := r.GetByIds(context.Background(), []uuid.UUID{id1, id2})
	if err != nil {
		t.Fatalf("GetByIds: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len(GetByIds()) = %d, want 2", len(got))
	}

	if got[0].ID != id1 || got[0].ComicID != comicID || got[0].Title != chapterTitle || got[0].PagesNb != 42 {
		t.Errorf("GetByIds()[0] = %+v", got[0])
	}

	if got[1].ID != id2 || got[1].Number != 2 || got[1].Download != 80 {
		t.Errorf("GetByIds()[1] = %+v", got[1])
	}
}

func TestChaptersGetByIdsError(t *testing.T) {
	t.Parallel()

	r, mock := newChaptersRepo(t)

	sentinel := errors.New("connection reset")
	id := uuid.New()

	mock.ExpectQuery(`FROM "chapters".*id IN`).
		WithArgs(id).
		WillReturnError(sentinel)

	got, err := r.GetByIds(context.Background(), []uuid.UUID{id})
	if errors.Is(err, domain.ErrNotFound) {
		t.Error("SQL failure must not be translated to ErrNotFound")
	}

	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, original error no longer reachable", err)
	}

	if got != nil {
		t.Errorf("GetByIds returned %+v in addition to the error", got)
	}
}
