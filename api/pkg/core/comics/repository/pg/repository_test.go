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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics/repository/pg"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgtest"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

const (
	comicSource = sources.SourceAsuraScans
	comicSlug   = "solo-leveling"
	comicTitle  = "Solo Leveling"
	comicStatus = sources.SeriesStatusCompleted
	comicType   = sources.SeriesTypeManhwa
)

func duplicateComicKeyErr() error {
	return &pgconn.PgError{
		Code:    "23505",
		Message: `duplicate key value violates unique constraint "idx_comic_source_slug"`,
	}
}

func newComicsRepo(t *testing.T) (*pg.PGComicsRepository, sqlmock.Sqlmock) {
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

func TestGetByID(t *testing.T) {
	t.Parallel()

	r, mock := newComicsRepo(t)

	id := uuid.New()
	userID := uuid.New()
	created := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	updated := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(`FROM "comics".*comics.id = \$1 AND library_entries.user_id = \$2`).
		WithArgs(id, userID, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "source", "slug", "title", "status", "comic_type",
				"chapter_count", "author", "artist", "description",
				"genres", "alt_titles", "created_at", "updated_at",
			}).AddRow(
				id, string(comicSource), comicSlug, comicTitle, string(comicStatus), string(comicType),
				200, "Chugong", "Dubu", "desc",
				"{}", "{}", created, updated,
			),
		)

	got, err := r.GetByID(context.Background(), comics.GetByIDOpts{ID: id, UserID: userID})
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.ID != id || got.Source != comicSource || got.Slug != comicSlug || got.Title != comicTitle {
		t.Errorf("GetByID() = %+v", got)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newComicsRepo(t)

	userID := uuid.New()

	mock.ExpectQuery(`FROM "comics".*comics.id = \$1 AND library_entries.user_id = \$2`).
		WithArgs(sqlmock.AnyArg(), userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	got, err := r.GetByID(context.Background(), comics.GetByIDOpts{ID: uuid.New(), UserID: userID})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID = %v, want domain.ErrNotFound", err)
	}

	if got != nil {
		t.Errorf("GetByID returned %+v in addition to the error", got)
	}
}

func TestGetBySourceSlug(t *testing.T) {
	t.Parallel()

	r, mock := newComicsRepo(t)

	id := uuid.New()
	userID := uuid.New()
	created := time.Now().UTC().Truncate(time.Second)
	updated := created

	mock.ExpectQuery(
		`FROM "comics".*comics.slug = \$1 AND comics.source = \$2 AND library_entries.user_id = \$3`,
	).
		WithArgs(comicSlug, string(comicSource), userID, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "source", "slug", "title", "status", "comic_type",
				"chapter_count", "author", "artist", "description",
				"genres", "alt_titles", "created_at", "updated_at",
			}).AddRow(
				id, string(comicSource), comicSlug, comicTitle, string(comicStatus), string(comicType),
				200, "Chugong", "Dubu", "desc",
				"{}", "{}", created, updated,
			),
		)

	got, err := r.GetBySourceSlug(context.Background(), comics.GetBySourceSlugOpts{
		UserID: userID,
		Source: comicSource,
		Slug:   comicSlug,
	})
	if err != nil {
		t.Fatalf("GetBySourceSlug: %v", err)
	}

	if got.Slug != comicSlug {
		t.Errorf("GetBySourceSlug() slug = %q, want %q", got.Slug, comicSlug)
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	r, mock := newComicsRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "comics"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	before := time.Now()

	got, err := r.Create(context.Background(), comics.CreateComicOpts{
		Source:       comicSource,
		Slug:         comicSlug,
		Title:        comicTitle,
		Status:       comicStatus,
		Type:         comicType,
		Genres:       []string{"action", "fantasy"},
		ChapterCount: 200,
		Author:       "Chugong",
		Artist:       "Dubu",
		Description:  "desc",
		AltTitles:    []string{"Na Honjaman Level Up"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if got.Slug != comicSlug || got.Type != comicType {
		t.Errorf("Create() = %+v", got)
	}

	if got.ID == uuid.Nil {
		t.Error("Create did not generate ID")
	}

	if got.CreatedAt.Before(before) || got.UpdatedAt.Before(before) {
		t.Errorf("Create timestamps too early: created=%v updated=%v", got.CreatedAt, got.UpdatedAt)
	}
}

func TestCreateDuplicate(t *testing.T) {
	t.Parallel()

	r, mock := newComicsRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "comics"`).
		WillReturnError(duplicateComicKeyErr())
	mock.ExpectRollback()

	_, err := r.Create(context.Background(), comics.CreateComicOpts{
		Source: comicSource,
		Slug:   comicSlug,
		Title:  comicTitle,
		Status: comicStatus,
		Type:   comicType,
	})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("Create = %v, want domain.ErrAlreadyExists", err)
	}
}

func TestGetMany(t *testing.T) {
	t.Parallel()

	r, mock := newComicsRepo(t)

	userID := uuid.New()
	id := uuid.New()
	created := time.Now().UTC().Truncate(time.Second)
	updated := created

	mock.ExpectQuery(`FROM "comics".*library_entries.user_id = \$1`).
		WithArgs(userID, 10).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "source", "slug", "title", "status", "comic_type",
				"chapter_count", "author", "artist", "description",
				"genres", "alt_titles", "created_at", "updated_at",
			}).AddRow(
				id, string(comicSource), comicSlug, comicTitle, string(comicStatus), string(comicType),
				200, "Chugong", "Dubu", "desc",
				"{}", "{}", created, updated,
			),
		)

	got, err := r.GetMany(context.Background(), comics.GetManyOpts{
		UserID: &userID,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("GetMany: %v", err)
	}

	if len(got) != 1 || got[0].ID != id {
		t.Errorf("GetMany() = %+v", got)
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	r, mock := newComicsRepo(t)

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "comics" WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := r.Delete(context.Background(), id)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestDeleteNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newComicsRepo(t)

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "comics" WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := r.Delete(context.Background(), id)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete = %v, want domain.ErrNotFound", err)
	}
}
