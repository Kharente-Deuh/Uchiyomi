// SPDX-License-Identifier: AGPL-3.0-or-later

//nolint:goconst
package pg_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/feed"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/feed/repository/pg"
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

func newFeedRepo(t *testing.T) (*pg.PGFeedRepository, sqlmock.Sqlmock) {
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

func TestListPageCountAndOrderSQL(t *testing.T) {
	t.Parallel()

	r, mock := newFeedRepo(t)

	userID := uuid.New()
	comicID := uuid.New()
	now := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)

	// count: userID, now
	mock.ExpectQuery(`COUNT\(DISTINCT comics.id\).*early_access_until IS NULL OR chapters.early_access_until <=`).
		WithArgs(userID, now).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	// page: userID, now (unlock join), now (availability CASE), limit, offset
	mock.ExpectQuery(`early_access_until IS NULL OR chapters.early_access_until <=.*ORDER BY.*comics.title`).
		WithArgs(userID, now, now, 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "source", "slug", "title", "status", "comic_type",
		}).AddRow(
			comicID, string(comicSource), comicSlug, comicTitle, string(comicStatus), string(comicType),
		))

	got, err := r.ListPage(context.Background(), feed.ListPageOpts{
		Now:    now,
		UserID: userID,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListPage: %v", err)
	}

	if got.Total != 1 || len(got.Items) != 1 {
		t.Fatalf("ListPage() = %+v", got)
	}

	item := got.Items[0]
	if item.ID != comicID || item.Source != comicSource || item.Slug != comicSlug ||
		item.Title != comicTitle || item.Status != comicStatus || item.Type != comicType {
		t.Errorf("Item = %+v", item)
	}
}

func TestListPageFilters(t *testing.T) {
	t.Parallel()

	r, mock := newFeedRepo(t)

	userID := uuid.New()
	now := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	src := comicSource
	typ := comicType

	// count: userID, now, source, type
	mock.ExpectQuery(`COUNT\(DISTINCT comics.id\).*comics.source.*comic_type`).
		WithArgs(userID, now, string(src), string(typ)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	// page: userID, now, source, type, now (CASE), limit, offset
	mock.ExpectQuery(`comics.source.*comic_type`).
		WithArgs(userID, now, string(src), string(typ), now, 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "source", "slug", "title", "status", "comic_type",
		}))

	got, err := r.ListPage(context.Background(), feed.ListPageOpts{
		Now:    now,
		Source: &src,
		Type:   &typ,
		UserID: userID,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("ListPage: %v", err)
	}

	if got.Total != 0 || len(got.Items) != 0 {
		t.Errorf("ListPage() = %+v", got)
	}
}

func TestListPageOffsetLimit(t *testing.T) {
	t.Parallel()

	r, mock := newFeedRepo(t)

	userID := uuid.New()
	now := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`COUNT\(DISTINCT comics.id\).*early_access_until IS NULL`).
		WithArgs(userID, now).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(40)))

	mock.ExpectQuery(`early_access_until IS NULL.*LIMIT`).
		WithArgs(userID, now, now, 10, 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "source", "slug", "title", "status", "comic_type",
		}))

	got, err := r.ListPage(context.Background(), feed.ListPageOpts{
		Now:    now,
		UserID: userID,
		Limit:  10,
		Offset: 20,
	})
	if err != nil {
		t.Fatalf("ListPage: %v", err)
	}

	if got.Total != 40 {
		t.Errorf("Total = %d, want 40", got.Total)
	}
}

func TestListUnlockedChaptersEmptyIDsSkipsQuery(t *testing.T) {
	t.Parallel()

	r, _ := newFeedRepo(t)

	got, err := r.ListUnlockedChapters(context.Background(), feed.ListChaptersOpts{
		Now:      time.Now().UTC(),
		ComicIDs: nil,
	})
	if err != nil {
		t.Fatalf("ListUnlockedChapters: %v", err)
	}

	if got != nil {
		t.Errorf("ListUnlockedChapters() = %+v, want nil", got)
	}
}

func TestListUnlockedChaptersFiltersLocked(t *testing.T) {
	t.Parallel()

	r, mock := newFeedRepo(t)

	comicID := uuid.New()
	chapterID := uuid.New()
	userID := uuid.New()
	now := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	published := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)

	// GORM expands IN ? to ($1) with PreferSimpleProtocol: comicID, now
	mock.ExpectQuery(`chapters.comic_id IN.*chapters.early_access_until IS NULL OR chapters.early_access_until <=`).
		WithArgs(userID, comicID, now).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "comic_id", "title", "number", "published_at", "early_access_until", "download", "has_progress",
		}).AddRow(chapterID, comicID, "Ch 10", 10.0, published, until, 2, true))

	got, err := r.ListUnlockedChapters(context.Background(), feed.ListChaptersOpts{
		Now:      now,
		ComicIDs: []uuid.UUID{comicID},
		UserID:   userID,
	})
	if err != nil {
		t.Fatalf("ListUnlockedChapters: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}

	ch := got[0]
	if ch.ID != chapterID || ch.ComicID != comicID || ch.Title != "Ch 10" ||
		ch.Number != 10.0 || ch.Download != 2 || !ch.PublishedAt.Equal(published) ||
		ch.EarlyAccessUntil == nil || !ch.EarlyAccessUntil.Equal(until) || !ch.HasProgress {
		t.Errorf("LatestChapter = %+v", ch)
	}
}
