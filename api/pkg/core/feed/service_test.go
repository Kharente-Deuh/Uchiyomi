// SPDX-License-Identifier: AGPL-3.0-or-later

//nolint:goconst
package feed_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/feed"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

type fakeFeedRepository struct {
	page           feed.Page
	pageErr        error
	chapters       []feed.LatestChapter
	chaptersErr    error
	lastPageOpts   feed.ListPageOpts
	lastChapters   feed.ListChaptersOpts
	pageCalls      int
	chaptersCalls  int
}

func (f *fakeFeedRepository) ListPage(_ context.Context, opts feed.ListPageOpts) (feed.Page, error) {
	f.pageCalls++
	f.lastPageOpts = opts

	return f.page, f.pageErr
}

func (f *fakeFeedRepository) ListUnlockedChapters(
	_ context.Context, opts feed.ListChaptersOpts,
) ([]feed.LatestChapter, error) {
	f.chaptersCalls++
	f.lastChapters = opts

	return f.chapters, f.chaptersErr
}

func frozenAug16() time.Time {
	return time.Date(2026, time.August, 16, 12, 0, 0, 0, time.UTC)
}

func frozenAug21() time.Time {
	return time.Date(2026, time.August, 21, 0, 0, 0, 0, time.UTC)
}

func newService(t *testing.T, repo feed.FeedRepository, now func() time.Time) *feed.Service {
	t.Helper()

	svc, err := feed.NewService(feed.Deps{FeedRepository: repo, Now: now})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	return svc
}

func TestNewServiceValidatesDeps(t *testing.T) {
	t.Parallel()

	_, err := feed.NewService(feed.Deps{})
	if err == nil {
		t.Fatal("NewService with empty deps must fail")
	}
}

func TestGetSkipsChaptersWhenPageEmpty(t *testing.T) {
	t.Parallel()

	repo := &fakeFeedRepository{page: feed.Page{Items: nil, Total: 0}}
	svc := newService(t, repo, frozenAug16)
	userID := uuid.New()

	got, err := svc.Get(context.Background(), feed.GetOpts{UserID: userID, Limit: 10})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if repo.chaptersCalls != 0 {
		t.Errorf("ListUnlockedChapters calls = %d, want 0", repo.chaptersCalls)
	}

	if got.Total != 0 || len(got.Items) != 0 {
		t.Errorf("Get() = %+v", got)
	}

	if repo.lastPageOpts.UserID != userID || !repo.lastPageOpts.Now.Equal(frozenAug16()) {
		t.Errorf("ListPage opts = %+v", repo.lastPageOpts)
	}
}

func TestGetOmitsAllLockedWhenRepoReturnsEmpty(t *testing.T) {
	t.Parallel()

	repo := &fakeFeedRepository{page: feed.Page{Items: []feed.Item{}, Total: 0}}
	svc := newService(t, repo, frozenAug16)

	got, err := svc.Get(context.Background(), feed.GetOpts{UserID: uuid.New(), Limit: 10})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got.Total != 0 || len(got.Items) != 0 {
		t.Errorf("all-locked comic must be absent, got %+v", got)
	}
}

func TestGetAugustScenarioLatestChapters(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	ch10ID := uuid.New()
	ch11ID := uuid.New()
	item := feed.Item{
		ID:     comicID,
		Title:  "Solo Leveling",
		Slug:   "solo-leveling",
		Source: sources.SourceAsuraScans,
		Status: sources.SeriesStatusOngoing,
		Type:   sources.SeriesTypeManhwa,
	}
	ch10 := feed.LatestChapter{
		ID:          ch10ID,
		ComicID:     comicID,
		Number:      10,
		PublishedAt: time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
		Download:    100,
	}
	ch11 := feed.LatestChapter{
		ID:               ch11ID,
		ComicID:          comicID,
		Number:           11,
		PublishedAt:      time.Date(2026, time.August, 10, 0, 0, 0, 0, time.UTC),
		EarlyAccessUntil: time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC),
		Download:         0,
	}

	t.Run("16 Aug only ch10", func(t *testing.T) {
		t.Parallel()

		repo := &fakeFeedRepository{
			page:     feed.Page{Items: []feed.Item{item}, Total: 1},
			chapters: []feed.LatestChapter{ch10},
		}
		svc := newService(t, repo, frozenAug16)

		got, err := svc.Get(context.Background(), feed.GetOpts{UserID: uuid.New(), Limit: 10})
		if err != nil {
			t.Fatalf("Get: %v", err)
		}

		if len(got.Items) != 1 || len(got.Items[0].LatestChapters) != 1 {
			t.Fatalf("items/chapters = %+v", got.Items)
		}

		if got.Items[0].LatestChapters[0].ID != ch10ID {
			t.Errorf("chapter = %s, want ch10", got.Items[0].LatestChapters[0].ID)
		}
	})

	t.Run("21 Aug ch11 first then ch10", func(t *testing.T) {
		t.Parallel()

		repo := &fakeFeedRepository{
			page:     feed.Page{Items: []feed.Item{item}, Total: 1},
			chapters: []feed.LatestChapter{ch10, ch11},
		}
		svc := newService(t, repo, frozenAug21)

		got, err := svc.Get(context.Background(), feed.GetOpts{UserID: uuid.New(), Limit: 10})
		if err != nil {
			t.Fatalf("Get: %v", err)
		}

		chs := got.Items[0].LatestChapters
		if len(chs) != 2 || chs[0].ID != ch11ID || chs[1].ID != ch10ID {
			t.Errorf("latestChapters = %+v, want ch11 then ch10", chs)
		}
	})
}

func TestGetCapsLatestChaptersAtThree(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	item := feed.Item{ID: comicID, Title: "A"}
	now := frozenAug21()
	var chs []feed.LatestChapter
	for i := 1; i <= 5; i++ {
		chs = append(chs, feed.LatestChapter{
			ID:          uuid.New(),
			ComicID:     comicID,
			Number:      float64(i),
			PublishedAt: now.Add(time.Duration(i) * time.Hour),
			Download:    100,
		})
	}

	repo := &fakeFeedRepository{
		page:     feed.Page{Items: []feed.Item{item}, Total: 1},
		chapters: chs,
	}
	svc := newService(t, repo, frozenAug21)

	got, err := svc.Get(context.Background(), feed.GetOpts{UserID: uuid.New(), Limit: 10})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	gotChs := got.Items[0].LatestChapters
	if len(gotChs) != 3 {
		t.Fatalf("len = %d, want 3", len(gotChs))
	}

	if gotChs[0].Number != 5 || gotChs[1].Number != 4 || gotChs[2].Number != 3 {
		t.Errorf("order by availability desc = %v %v %v", gotChs[0].Number, gotChs[1].Number, gotChs[2].Number)
	}
}

func TestGetPreservesPageOrderAndTotal(t *testing.T) {
	t.Parallel()

	a := uuid.New()
	b := uuid.New()
	repo := &fakeFeedRepository{
		page: feed.Page{
			Items: []feed.Item{{ID: a, Title: "A"}, {ID: b, Title: "B"}},
			Total: 12,
		},
		chapters: []feed.LatestChapter{
			{ID: uuid.New(), ComicID: b, Number: 1, PublishedAt: frozenAug16()},
			{ID: uuid.New(), ComicID: a, Number: 2, PublishedAt: frozenAug16()},
		},
	}
	svc := newService(t, repo, frozenAug16)

	got, err := svc.Get(context.Background(), feed.GetOpts{UserID: uuid.New(), Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got.Total != 12 || got.Items[0].ID != a || got.Items[1].ID != b {
		t.Errorf("page = %+v", got)
	}

	if len(got.Items[0].LatestChapters) != 1 || len(got.Items[1].LatestChapters) != 1 {
		t.Errorf("chapters not attached per comic: %+v", got.Items)
	}
}

func TestGetPassesFiltersAndNow(t *testing.T) {
	t.Parallel()

	src := sources.SourceAsuraScans
	typ := sources.SeriesTypeManhwa
	st := sources.SeriesStatusOngoing
	userID := uuid.New()
	repo := &fakeFeedRepository{page: feed.Page{Items: nil, Total: 0}}
	svc := newService(t, repo, frozenAug16)

	_, err := svc.Get(context.Background(), feed.GetOpts{
		UserID: userID,
		Source: &src,
		Type:   &typ,
		Status: &st,
		Limit:  5,
		Offset: 2,
	})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	o := repo.lastPageOpts
	if o.UserID != userID || o.Limit != 5 || o.Offset != 2 || o.Source == nil || *o.Source != src {
		t.Errorf("ListPage opts = %+v", o)
	}

	if o.Type == nil || *o.Type != typ || o.Status == nil || *o.Status != st {
		t.Errorf("filters = %+v", o)
	}

	if !o.Now.Equal(frozenAug16()) {
		t.Errorf("Now = %v", o.Now)
	}
}

func TestGetWrapsRepositoryError(t *testing.T) {
	t.Parallel()

	repo := &fakeFeedRepository{pageErr: errors.New("db down")}
	svc := newService(t, repo, frozenAug16)

	_, err := svc.Get(context.Background(), feed.GetOpts{UserID: uuid.New()})
	if err == nil {
		t.Fatal("error expected")
	}
}
