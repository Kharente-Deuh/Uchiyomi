// SPDX-License-Identifier: AGPL-3.0-or-later

//nolint:goconst
package pg_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readersettings"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readersettings/repository/pg"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgtest"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

func newReaderSettingsRepo(t *testing.T) (*pg.PGReaderSettingsRepository, sqlmock.Sqlmock) {
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

func TestListByUserFiltersUserID(t *testing.T) {
	t.Parallel()

	r, mock := newReaderSettingsRepo(t)

	userID := uuid.New()
	otherUserID := uuid.New()
	id := uuid.New()

	mock.ExpectQuery(`FROM "reader_settings".*user_id`).
		WithArgs(userID).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "user_id", "comic_type", "reading_mode", "page_scale", "double_page",
			}).AddRow(
				id, userID, "manga",
				string(readersettings.ReadingModePagedRTL),
				string(readersettings.PageScaleFitScreen),
				true,
			),
		)

	got, err := r.ListByUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("ListByUser() len = %d, want 1", len(got))
	}

	want := readersettings.Profile{
		Type:        sources.SeriesTypeManga,
		ReadingMode: readersettings.ReadingModePagedRTL,
		PageScale:   readersettings.PageScaleFitScreen,
		DoublePage:  true,
	}
	if got[0] != want {
		t.Errorf("ListByUser()[0] = %+v, want %+v", got[0], want)
	}

	if otherUserID == userID {
		t.Fatal("unlucky uuid collision")
	}
}

func TestUpsertInserts(t *testing.T) {
	t.Parallel()

	r, mock := newReaderSettingsRepo(t)

	userID := uuid.New()
	opts := mangaUpsertOpts(userID)

	expectUpsertQuery(mock, userID, `INSERT INTO "reader_settings"`)

	got, err := r.Upsert(context.Background(), opts)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	want := readersettings.Profile{
		Type:        sources.SeriesTypeManga,
		ReadingMode: opts.ReadingMode,
		PageScale:   opts.PageScale,
		DoublePage:  opts.DoublePage,
	}
	if got != want {
		t.Errorf("Upsert() = %+v, want %+v", got, want)
	}
}

func TestUpsertConflictUpdates(t *testing.T) {
	t.Parallel()

	r, mock := newReaderSettingsRepo(t)

	userID := uuid.New()
	opts := mangaUpsertOpts(userID)

	expectUpsertQuery(mock, userID, `ON CONFLICT`)

	got, err := r.Upsert(context.Background(), opts)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	want := readersettings.Profile{
		Type:        sources.SeriesTypeManga,
		ReadingMode: opts.ReadingMode,
		PageScale:   opts.PageScale,
		DoublePage:  opts.DoublePage,
	}
	if got != want {
		t.Errorf("Upsert() = %+v, want %+v", got, want)
	}
}

func mangaUpsertOpts(userID uuid.UUID) readersettings.UpsertOpts {
	return readersettings.UpsertOpts{
		UserID:      userID,
		Type:        sources.SeriesTypeManga,
		ReadingMode: readersettings.ReadingModePagedRTL,
		PageScale:   readersettings.PageScaleFitScreen,
		DoublePage:  false,
	}
}

func expectUpsertQuery(mock sqlmock.Sqlmock, userID uuid.UUID, sqlRegex string) {
	mock.ExpectBegin()
	mock.ExpectQuery(sqlRegex).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			string(readersettings.ReadingModePagedRTL),
			string(readersettings.PageScaleFitScreen),
			"manga",
			userID,
			false,
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()
}
