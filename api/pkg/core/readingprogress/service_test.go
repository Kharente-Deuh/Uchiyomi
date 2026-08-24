// SPDX-License-Identifier: AGPL-3.0-or-later

package readingprogress_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readingprogress"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
)

type fakeTransactor struct {
	err   error
	calls int
}

func (f *fakeTransactor) WithinTx(ctx context.Context, _ transaction.TxOpts, fn func(context.Context) error) error {
	f.calls++
	if err := fn(ctx); err != nil {
		return err
	}

	return f.err
}

type fakeRepo struct {
	existing    *readingprogress.Progress
	listErr     error
	getErr      error
	upsertErr   error
	rows        []readingprogress.Progress
	upserts     []readingprogress.UpsertOpts
	lastUpsert  readingprogress.UpsertOpts
	lastGet     readingprogress.GetOpts
	listed      readingprogress.ListOpts
	listCalls   int
	getCalls    int
	upsertCalls int
}

func (f *fakeRepo) GetLatestByUserAndComic(
	_ context.Context,
	opts readingprogress.ListOpts,
) (*readingprogress.Progress, error) {
	f.listCalls++
	f.listed = opts
	if f.listErr != nil {
		return nil, f.listErr
	}

	if len(f.rows) == 0 {
		return nil, nil
	}

	row := f.rows[0]

	return &row, nil
}

func (f *fakeRepo) ListByUserAndChapterIDs(
	_ context.Context,
	_ readingprogress.MapOpts,
) ([]readingprogress.Progress, error) {
	f.listCalls++
	if f.listErr != nil {
		return nil, f.listErr
	}

	return f.rows, nil
}

func (f *fakeRepo) Get(_ context.Context, opts readingprogress.GetOpts) (*readingprogress.Progress, error) {
	f.getCalls++
	f.lastGet = opts
	if f.getErr != nil {
		return nil, f.getErr
	}

	return f.existing, nil
}

func (f *fakeRepo) Upsert(_ context.Context, opts readingprogress.UpsertOpts) (readingprogress.Progress, error) {
	f.upsertCalls++
	f.lastUpsert = opts
	f.upserts = append(f.upserts, opts)

	return readingprogress.Progress{
		ChapterID: opts.ChapterID,
		Page:      opts.Page,
		UpdatedAt: opts.UpdatedAt,
	}, f.upsertErr
}

type fakeLibrary struct {
	err     error
	inLib   bool
	userID  uuid.UUID
	comicID uuid.UUID
	calls   int
}

func (f *fakeLibrary) ExistsByUserAndComic(_ context.Context, userID, comicID uuid.UUID) (bool, error) {
	f.calls++
	f.userID = userID
	f.comicID = comicID

	return f.inLib, f.err
}

type fakeComics struct {
	err    error
	exists bool
	id     uuid.UUID
	calls  int
}

func (f *fakeComics) Exists(_ context.Context, id uuid.UUID) (bool, error) {
	f.calls++
	f.id = id

	return f.exists, f.err
}

type fakeChapters struct {
	byID        map[uuid.UUID]chapters.Chapter
	chapter     *chapters.Chapter
	getErr      error
	getByIdsErr error
	lastIDs     []uuid.UUID
	lastID      uuid.UUID
	getCalls    int
}

func (f *fakeChapters) GetByID(_ context.Context, id uuid.UUID) (*chapters.Chapter, error) {
	f.getCalls++
	f.lastID = id
	if f.getErr != nil {
		return nil, f.getErr
	}

	if f.byID != nil {
		ch, ok := f.byID[id]
		if !ok {
			return nil, domain.ErrNotFound
		}

		return &ch, nil
	}

	return f.chapter, nil
}

func (f *fakeChapters) GetByIds(_ context.Context, ids []uuid.UUID) ([]chapters.Chapter, error) {
	f.lastIDs = ids
	if f.getByIdsErr != nil {
		return nil, f.getByIdsErr
	}

	if f.byID != nil {
		out := make([]chapters.Chapter, 0, len(ids))
		for _, id := range ids {
			if ch, ok := f.byID[id]; ok {
				out = append(out, ch)
			}
		}

		return out, nil
	}

	if f.chapter != nil {
		return []chapters.Chapter{*f.chapter}, nil
	}

	return nil, nil
}

func newService(
	t *testing.T,
	repo readingprogress.Repository,
	lib readingprogress.LibraryMembership,
	comics readingprogress.ComicLookup,
	ch readingprogress.ChapterLookup,
) *readingprogress.Service {
	t.Helper()

	svc, err := readingprogress.NewService(readingprogress.Deps{
		Repository: repo,
		Transactor: &fakeTransactor{},
		Library:    lib,
		Comics:     comics,
		Chapters:   ch,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	return svc
}

func TestNewServiceValidatesDeps(t *testing.T) {
	t.Parallel()

	_, err := readingprogress.NewService(readingprogress.Deps{})
	if err == nil {
		t.Fatal("NewService with empty deps must fail")
	}
}

func TestNewServiceRequiresTransactor(t *testing.T) {
	t.Parallel()

	_, err := readingprogress.NewService(readingprogress.Deps{
		Repository: &fakeRepo{},
		Library:    &fakeLibrary{inLib: true},
		Comics:     &fakeComics{exists: true},
		Chapters:   &fakeChapters{},
	})
	if err == nil {
		t.Fatal("NewService without transactor must fail")
	}
}

func TestSaveTwoChaptersContinueIsLastOpened(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicID := uuid.New()
	ch10 := uuid.New()
	ch3 := uuid.New()
	repo := &fakeRepo{getErr: domain.ErrNotFound}
	chap := &fakeChapters{chapter: &chapters.Chapter{ID: ch10, ComicID: comicID, PagesNb: 30, Download: 40}}
	svc := newService(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, chap)

	if _, err := svc.Save(context.Background(), readingprogress.SaveOpts{
		UserID: userID, ChapterID: ch10, Page: 5,
	}); err != nil {
		t.Fatalf("Save ch10: %v", err)
	}

	chap.chapter = &chapters.Chapter{ID: ch3, ComicID: comicID, PagesNb: 20, Download: 100}
	got, err := svc.Save(context.Background(), readingprogress.SaveOpts{
		UserID: userID, ChapterID: ch3, Page: 12,
	})
	if err != nil {
		t.Fatalf("Save ch3: %v", err)
	}

	if got.ChapterID != ch3 || got.Page != 12 {
		t.Errorf("Save ch3 = %+v", got)
	}

	if repo.lastUpsert.ChapterID != ch3 || repo.lastUpsert.UserID != userID {
		t.Errorf("last upsert = %+v", repo.lastUpsert)
	}

	if chap.chapter.Download != 100 {
		t.Fatal("test setup")
	}
}

func TestSaveDoesNotReadDownload(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chID := uuid.New()
	repo := &fakeRepo{getErr: domain.ErrNotFound}
	svc := newService(t, repo, &fakeLibrary{inLib: true}, &fakeComics{}, &fakeChapters{
		chapter: &chapters.Chapter{ID: chID, ComicID: comicID, PagesNb: 10, Download: 40},
	})

	if _, err := svc.Save(context.Background(), readingprogress.SaveOpts{
		UserID: uuid.New(), ChapterID: chID, Page: 2,
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if repo.upsertCalls != 1 {
		t.Fatalf("upsert calls = %d", repo.upsertCalls)
	}
}

func TestSaveLowerPageKeepsFurthestAndWrites(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chID := uuid.New()
	repo := &fakeRepo{
		existing: &readingprogress.Progress{ChapterID: chID, Page: 20, UpdatedAt: time.Unix(1, 0).UTC()},
	}
	svc := newService(t, repo, &fakeLibrary{inLib: true}, &fakeComics{}, &fakeChapters{
		chapter: &chapters.Chapter{ID: chID, ComicID: comicID, PagesNb: 30},
	})

	got, err := svc.Save(context.Background(), readingprogress.SaveOpts{
		UserID: uuid.New(), ChapterID: chID, Page: 8,
	})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	if got.Page != 20 {
		t.Errorf("page = %d, want 20", got.Page)
	}

	if repo.upsertCalls != 1 || repo.lastUpsert.Page != 20 {
		t.Errorf("upsert page = %d, calls = %d", repo.lastUpsert.Page, repo.upsertCalls)
	}
}

func TestSaveGetNotFoundTreatsAsNoStoredRow(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chID := uuid.New()
	repo := &fakeRepo{getErr: domain.ErrNotFound}
	svc := newService(t, repo, &fakeLibrary{inLib: true}, &fakeComics{}, &fakeChapters{
		chapter: &chapters.Chapter{ID: chID, ComicID: comicID, PagesNb: 10},
	})

	got, err := svc.Save(context.Background(), readingprogress.SaveOpts{
		UserID: uuid.New(), ChapterID: chID, Page: 3,
	})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	if got.Page != 3 {
		t.Errorf("page = %d, want 3", got.Page)
	}

	if repo.upsertCalls != 1 {
		t.Fatalf("upsert calls = %d", repo.upsertCalls)
	}
}

func TestSaveGetErrorWraps(t *testing.T) {
	t.Parallel()

	getErr := errors.New("db down")
	comicID := uuid.New()
	chID := uuid.New()
	repo := &fakeRepo{getErr: getErr}
	svc := newService(t, repo, &fakeLibrary{inLib: true}, &fakeComics{}, &fakeChapters{
		chapter: &chapters.Chapter{ID: chID, ComicID: comicID, PagesNb: 10},
	})

	_, err := svc.Save(context.Background(), readingprogress.SaveOpts{
		UserID: uuid.New(), ChapterID: chID, Page: 1,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, getErr) {
		t.Errorf("err = %v, want wrapped %v", err, getErr)
	}
}

func TestSaveForbiddenWhenNotInLibrary(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chID := uuid.New()
	repo := &fakeRepo{}
	svc := newService(t, repo, &fakeLibrary{inLib: false}, &fakeComics{exists: true}, &fakeChapters{
		chapter: &chapters.Chapter{ID: chID, ComicID: comicID, PagesNb: 10},
	})

	_, err := svc.Save(context.Background(), readingprogress.SaveOpts{
		UserID: uuid.New(), ChapterID: chID, Page: 1,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}

	if repo.upsertCalls != 0 {
		t.Error("upsert on forbidden")
	}
}

func TestSaveNotFoundWhenComicMissing(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chID := uuid.New()
	repo := &fakeRepo{}
	svc := newService(t, repo, &fakeLibrary{inLib: false}, &fakeComics{exists: false}, &fakeChapters{
		chapter: &chapters.Chapter{ID: chID, ComicID: comicID, PagesNb: 10},
	})

	_, err := svc.Save(context.Background(), readingprogress.SaveOpts{
		UserID: uuid.New(), ChapterID: chID, Page: 1,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestSaveUnknownChapter(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	svc := newService(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, &fakeChapters{
		getErr: domain.ErrNotFound,
	})

	_, err := svc.Save(context.Background(), readingprogress.SaveOpts{
		UserID: uuid.New(), ChapterID: uuid.New(), Page: 1,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}

	if repo.upsertCalls != 0 {
		t.Error("upsert for unknown chapter")
	}
}

func TestSaveRejectsPageAbovePagesNb(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chID := uuid.New()
	repo := &fakeRepo{}
	svc := newService(t, repo, &fakeLibrary{inLib: true}, &fakeComics{}, &fakeChapters{
		chapter: &chapters.Chapter{ID: chID, ComicID: comicID, PagesNb: 15},
	})

	_, err := svc.Save(context.Background(), readingprogress.SaveOpts{
		UserID: uuid.New(), ChapterID: chID, Page: 16,
	})
	if !errors.Is(err, readingprogress.ErrInvalid) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}

	if repo.upsertCalls != 0 {
		t.Error("upsert on invalid page")
	}
}

func TestListEmptyContinueNull(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicID := uuid.New()
	repo := &fakeRepo{rows: nil}
	svc := newService(t, repo, &fakeLibrary{inLib: true}, &fakeComics{}, &fakeChapters{})

	got, err := svc.List(context.Background(), readingprogress.ListOpts{UserID: userID, ComicID: comicID})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if got.Continue != nil {
		t.Errorf("continue = %+v, want nil", got.Continue)
	}

	if repo.listed.UserID != userID || repo.listed.ComicID != comicID {
		t.Errorf("list opts = %+v", repo.listed)
	}
}

func TestListClampsWithoutWrite(t *testing.T) {
	t.Parallel()

	chID := uuid.New()
	repo := &fakeRepo{rows: []readingprogress.Progress{{
		ChapterID: chID,
		Page:      18,
		UpdatedAt: time.Unix(2, 0).UTC(),
	}}}
	svc := newService(t, repo, &fakeLibrary{inLib: true}, &fakeComics{}, &fakeChapters{
		chapter: &chapters.Chapter{ID: chID, PagesNb: 12, Download: 50},
	})

	got, err := svc.List(context.Background(), readingprogress.ListOpts{UserID: uuid.New(), ComicID: uuid.New()})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if got.Continue == nil || got.Continue.Page != 12 || got.Continue.ChapterID != chID {
		t.Errorf("continue = %+v", got.Continue)
	}

	if repo.upsertCalls != 0 {
		t.Error("List must not upsert")
	}
}

func TestListForbidden(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	svc := newService(t, repo, &fakeLibrary{inLib: false}, &fakeComics{exists: true}, &fakeChapters{})

	_, err := svc.List(context.Background(), readingprogress.ListOpts{UserID: uuid.New(), ComicID: uuid.New()})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}

	if repo.listCalls != 0 {
		t.Error("listed without membership")
	}
}
