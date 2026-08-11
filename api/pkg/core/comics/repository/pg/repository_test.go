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
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics/repository/pg"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgtest"
)

const (
	comicSource = "asurascans"
	comicSlug   = "solo-leveling"
	comicTitle  = "Solo Leveling"
	comicStatus = "completed"
	comicType   = "manhwa"
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
	created := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	updated := time.Now().UTC().Truncate(time.Second)

	mock.ExpectQuery(`SELECT \* FROM "comics" WHERE id = \$1`).
		WithArgs(id, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "source", "slug", "title", "status", "comic_type",
				"chapter_count", "author", "artist", "description", "rating", "release_year",
				"source_url", "external_cover_url", "local_cover_path", "created_at", "updated_at",
			}).AddRow(
				id, comicSource, comicSlug, comicTitle, comicStatus, comicType,
				200, "Chugong", "Dubu", "desc", 9.5, 2018,
				"https://api.asurascans.com/series/solo-leveling", "https://cover.example/cover.webp",
				"covers/solo-leveling.webp", created, updated,
			),
		)

	got, err := r.GetByID(context.Background(), id)
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

	mock.ExpectQuery(`SELECT \* FROM "comics"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	got, err := r.GetByID(context.Background(), uuid.New())
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
	created := time.Now().UTC().Truncate(time.Second)
	updated := created

	mock.ExpectQuery(`SELECT \* FROM "comics" WHERE source = \$1 AND slug = \$2`).
		WithArgs(comicSource, comicSlug, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "source", "slug", "title", "status", "comic_type",
				"chapter_count", "author", "artist", "description", "rating", "release_year",
				"source_url", "external_cover_url", "local_cover_path", "created_at", "updated_at",
			}).AddRow(
				id, comicSource, comicSlug, comicTitle, comicStatus, comicType,
				200, "Chugong", "Dubu", "desc", 9.5, 2018,
				"", "", "", created, updated,
			),
		)

	got, err := r.GetBySourceSlug(context.Background(), comics.SourceSlugKey{
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
		t.Errorf("timestamps not set: created=%v updated=%v", got.CreatedAt, got.UpdatedAt)
	}
}

func TestCreateDuplicateKey(t *testing.T) {
	t.Parallel()

	r, mock := newComicsRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "comics"`).WillReturnError(duplicateComicKeyErr())
	mock.ExpectRollback()

	got, err := r.Create(context.Background(), comics.CreateComicOpts{
		Source: comicSource,
		Slug:   comicSlug,
		Title:  comicTitle,
	})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("Create = %v, want domain.ErrAlreadyExists", err)
	}

	if got != nil {
		t.Errorf("Create returned %+v in addition to the error", got)
	}
}
