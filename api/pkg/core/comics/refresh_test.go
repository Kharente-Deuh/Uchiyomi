// SPDX-License-Identifier: AGPL-3.0-or-later

package comics_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

func TestRefreshChapterListsListsOngoingAndHiatusPerSource(t *testing.T) {
	t.Parallel()

	repo := &fakeComicsRepository{}
	deps := validServiceDeps()
	deps.ComicsRepository = repo

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	if err := svc.RefreshChapterLists(context.Background()); err != nil {
		t.Fatalf("RefreshChapterLists: %v", err)
	}

	if repo.listByStatusesCalls != 1 {
		t.Fatalf("ListByStatuses called %d times, want 1", repo.listByStatusesCalls)
	}

	if repo.lastListByStatuses.Source != testSource {
		t.Errorf("source = %q, want %q", repo.lastListByStatuses.Source, testSource)
	}

	want := []sources.SeriesStatus{sources.SeriesStatusOngoing, sources.SeriesStatusHiatus}
	if len(repo.lastListByStatuses.Statuses) != 2 ||
		repo.lastListByStatuses.Statuses[0] != want[0] ||
		repo.lastListByStatuses.Statuses[1] != want[1] {
		t.Errorf("statuses = %v, want %v", repo.lastListByStatuses.Statuses, want)
	}
}

func TestRefreshChapterListsCreatesOnlyMissingChaptersAndEnqueuesThem(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	existing := chapters.Chapter{
		ID:                uuid.New(),
		ComicID:           comicID,
		SourceChapterSlug: testChapterSlug,
		Number:            1,
		Download:          100,
	}
	newSrc := sources.SourceChapter{
		SourceChapterSlug: testChapter2Slug,
		Number:            2,
		Title:             "Chapter 2",
		PageCount:         18,
		PublishedAt:       time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
	}
	created := chapters.Chapter{
		ID:                uuid.New(),
		ComicID:           comicID,
		SourceChapterSlug: newSrc.SourceChapterSlug,
		Number:            newSrc.Number,
		Title:             newSrc.Title,
		PagesNb:           newSrc.PageCount,
		PublishedAt:       newSrc.PublishedAt,
	}

	repo := &fakeComicsRepository{
		listByStatusesResult: []comics.Comic{{
			ID:     comicID,
			Source: testSource,
			Slug:   testSlug,
			Status: sources.SeriesStatusOngoing,
		}},
	}
	chaptersSvc := &fakeChaptersService{
		listByComicIDResult: []chapters.Chapter{existing},
		createAllResult:     []chapters.Chapter{created},
	}
	source := &fakeSource{
		infos: &sources.GetInfosBySlugResponse{
			Status:       sources.SeriesStatusOngoing,
			ChapterCount: 2,
		},
		chapters: []sources.SourceChapter{
			{
				SourceChapterSlug: existing.SourceChapterSlug,
				Number:            1,
				Title:             testChapterTitle,
			},
			newSrc,
		},
	}

	deps := validServiceDeps()
	deps.ComicsRepository = repo
	deps.ChaptersService = chaptersSvc
	deps.Sources = sources.SourceMap{testSource: source}

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	if err := svc.RefreshChapterLists(context.Background()); err != nil {
		t.Fatalf("RefreshChapterLists: %v", err)
	}

	if !source.lastInfosFresh || !source.lastChaptersFresh {
		t.Errorf("fresh infos=%v chapters=%v, want both true", source.lastInfosFresh, source.lastChaptersFresh)
	}

	if chaptersSvc.createAllCalls != 1 {
		t.Fatalf("CreateAll called %d times, want 1", chaptersSvc.createAllCalls)
	}

	if len(chaptersSvc.lastCreateAllChapters) != 1 ||
		chaptersSvc.lastCreateAllChapters[0].SourceChapterSlug != testChapter2Slug {
		t.Errorf("CreateAll chapters = %+v, want only %s", chaptersSvc.lastCreateAllChapters, testChapter2Slug)
	}

	if chaptersSvc.enqueueDownloadableCalls != 1 {
		t.Fatalf("EnqueueDownloadable called %d times, want 1", chaptersSvc.enqueueDownloadableCalls)
	}

	if len(chaptersSvc.lastEnqueueChapters) != 1 || chaptersSvc.lastEnqueueChapters[0].ID != created.ID {
		t.Errorf("enqueued = %+v, want created chapter", chaptersSvc.lastEnqueueChapters)
	}

	if repo.updateStatusCalls != 1 {
		t.Fatalf("UpdateStatusAndChapterCount called %d times, want 1", repo.updateStatusCalls)
	}

	if repo.lastUpdateStatus.ID != comicID ||
		repo.lastUpdateStatus.Status != sources.SeriesStatusOngoing ||
		repo.lastUpdateStatus.ChapterCount != 2 {
		t.Errorf("update opts = %+v", repo.lastUpdateStatus)
	}
}

func TestRefreshChapterListsFetchesChaptersWhenInfosBecomeCompleted(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	repo := &fakeComicsRepository{
		listByStatusesResult: []comics.Comic{{
			ID:     comicID,
			Source: testSource,
			Slug:   testSlug,
			Status: sources.SeriesStatusOngoing,
		}},
	}
	source := &fakeSource{
		infos: &sources.GetInfosBySlugResponse{
			Status:       sources.SeriesStatusCompleted,
			ChapterCount: 200,
		},
		chapters: []sources.SourceChapter{{
			SourceChapterSlug: "final",
			Number:            200,
		}},
	}
	chaptersSvc := &fakeChaptersService{}

	deps := validServiceDeps()
	deps.ComicsRepository = repo
	deps.ChaptersService = chaptersSvc
	deps.Sources = sources.SourceMap{testSource: source}

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	if err := svc.RefreshChapterLists(context.Background()); err != nil {
		t.Fatalf("RefreshChapterLists: %v", err)
	}

	if source.chaptersCalls != 1 {
		t.Errorf("GetChaptersBySlug called %d times, want 1", source.chaptersCalls)
	}

	if repo.lastUpdateStatus.Status != sources.SeriesStatusCompleted {
		t.Errorf("status = %q, want completed", repo.lastUpdateStatus.Status)
	}
}

func TestRefreshChapterListsContinuesAfterComicFailure(t *testing.T) {
	t.Parallel()

	failID := uuid.New()
	okID := uuid.New()
	repo := &fakeComicsRepository{
		listByStatusesResult: []comics.Comic{
			{ID: failID, Source: testSource, Slug: "broken", Status: sources.SeriesStatusOngoing},
			{ID: okID, Source: testSource, Slug: testSlug, Status: sources.SeriesStatusOngoing},
		},
	}
	okSource := &fakeSource{
		infos: &sources.GetInfosBySlugResponse{
			Status:       sources.SeriesStatusOngoing,
			ChapterCount: 1,
		},
		chapters: []sources.SourceChapter{{SourceChapterSlug: "c1", Number: 1}},
	}
	failing := &failThenOKSource{
		failSlug: "broken",
		ok:       okSource,
		failErr:  errors.New("gone"),
	}

	deps := validServiceDeps()
	deps.ComicsRepository = repo
	deps.ChaptersService = &fakeChaptersService{}
	deps.Sources = sources.SourceMap{testSource: failing}

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	if err := svc.RefreshChapterLists(context.Background()); err != nil {
		t.Fatalf("RefreshChapterLists: %v", err)
	}

	if failing.okCalls == 0 {
		t.Fatal("second comic was not refreshed after the first failed")
	}
}

type failThenOKSource struct {
	failErr  error
	ok       *fakeSource
	failSlug string
	okCalls  int
}

func (f *failThenOKSource) GetInfosBySlug(
	ctx context.Context,
	opts sources.GetInfosBySlugOpts,
) (*sources.GetInfosBySlugResponse, error) {
	if opts.Slug == f.failSlug {
		return nil, f.failErr
	}

	f.okCalls++

	return f.ok.GetInfosBySlug(ctx, opts)
}

func (f *failThenOKSource) GetChaptersBySlug(
	ctx context.Context,
	opts sources.GetChaptersBySlugOpts,
) ([]sources.SourceChapter, error) {
	return f.ok.GetChaptersBySlug(ctx, opts)
}

func (f *failThenOKSource) GetPageURLsByChapter(context.Context, sources.GetPageURLsByChapterOpts) ([]string, error) {
	return nil, nil
}

func TestRefreshChapterListsSkipsCreateWhenNoMissingChapters(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	repo := &fakeComicsRepository{
		listByStatusesResult: []comics.Comic{{
			ID:     comicID,
			Source: testSource,
			Slug:   testSlug,
			Status: sources.SeriesStatusHiatus,
		}},
	}
	chaptersSvc := &fakeChaptersService{
		listByComicIDResult: []chapters.Chapter{{
			ComicID:           comicID,
			SourceChapterSlug: testChapterSlug,
		}},
	}
	source := &fakeSource{
		infos: &sources.GetInfosBySlugResponse{
			Status:       sources.SeriesStatusHiatus,
			ChapterCount: 1,
		},
		chapters: []sources.SourceChapter{{SourceChapterSlug: testChapterSlug, Number: 1}},
	}

	deps := validServiceDeps()
	deps.ComicsRepository = repo
	deps.ChaptersService = chaptersSvc
	deps.Sources = sources.SourceMap{testSource: source}

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	if err := svc.RefreshChapterLists(context.Background()); err != nil {
		t.Fatalf("RefreshChapterLists: %v", err)
	}

	if chaptersSvc.createAllCalls != 0 {
		t.Errorf("CreateAll called %d times, want 0", chaptersSvc.createAllCalls)
	}

	if chaptersSvc.enqueueDownloadableCalls != 0 {
		t.Errorf("EnqueueDownloadable called %d times, want 0", chaptersSvc.enqueueDownloadableCalls)
	}

	if repo.updateStatusCalls != 1 {
		t.Errorf("UpdateStatusAndChapterCount called %d times, want 1", repo.updateStatusCalls)
	}
}
