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
	latest      *readingprogress.Progress
	existing    *readingprogress.Progress
	listErr     error
	getErr      error
	upsertErr   error
	stored      []readingprogress.Progress
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

	if f.latest != nil {
		return f.latest, nil
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

	if f.stored != nil {
		return f.stored, nil
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
	out := readingprogress.Progress{
		ChapterID: opts.ChapterID,
		Page:      opts.Page,
		UpdatedAt: opts.UpdatedAt,
	}
	f.latest = &out
	row := out
	f.rows = []readingprogress.Progress{row}

	return out, f.upsertErr
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

func ch(id, comicID uuid.UUID, number float64, pagesNb int) chapters.Chapter {
	return chapters.Chapter{ID: id, ComicID: comicID, Number: number, PagesNb: pagesNb}
}

func markReadSvc(
	t *testing.T,
	repo *fakeRepo,
	lib readingprogress.LibraryMembership,
	comics readingprogress.ComicLookup,
	chaps *fakeChapters,
	tx *fakeTransactor,
) *readingprogress.Service {
	t.Helper()

	svc, err := readingprogress.NewService(readingprogress.Deps{
		Repository: repo,
		Transactor: tx,
		Library:    lib,
		Comics:     comics,
		Chapters:   chaps,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	return svc
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

func TestMarkReadEmptyChapterIDs(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		ids  []uuid.UUID
	}{
		{name: "nil", ids: nil},
		{name: "empty slice", ids: []uuid.UUID{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &fakeRepo{}
			tx := &fakeTransactor{}
			svc := markReadSvc(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, &fakeChapters{}, tx)

			_, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
				UserID:     uuid.New(),
				ComicID:    uuid.New(),
				ChapterIDs: tc.ids,
			})
			if !errors.Is(err, readingprogress.ErrInvalid) {
				t.Errorf("err = %v, want ErrInvalid", err)
			}
			if tx.calls != 0 {
				t.Errorf("tx.calls = %d, want 0", tx.calls)
			}
			if repo.upsertCalls != 0 {
				t.Errorf("repo.upsertCalls = %d, want 0", repo.upsertCalls)
			}
		})
	}
}

//nolint:dupl
func TestMarkReadForbiddenWhenNotInLibrary(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chID := uuid.New()
	repo := &fakeRepo{}
	tx := &fakeTransactor{}
	svc := markReadSvc(t, repo, &fakeLibrary{inLib: false}, &fakeComics{exists: true}, &fakeChapters{
		byID: map[uuid.UUID]chapters.Chapter{
			chID: ch(chID, comicID, 1.0, 10),
		},
	}, tx)

	_, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
		UserID:     uuid.New(),
		ComicID:    comicID,
		ChapterIDs: []uuid.UUID{chID},
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
	if tx.calls != 0 {
		t.Errorf("tx.calls = %d, want 0", tx.calls)
	}
	if repo.upsertCalls != 0 {
		t.Errorf("repo.upsertCalls = %d, want 0", repo.upsertCalls)
	}
}

//nolint:dupl
func TestMarkReadNotFoundWhenComicMissing(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chID := uuid.New()
	repo := &fakeRepo{}
	tx := &fakeTransactor{}
	svc := markReadSvc(t, repo, &fakeLibrary{inLib: false}, &fakeComics{exists: false}, &fakeChapters{
		byID: map[uuid.UUID]chapters.Chapter{
			chID: ch(chID, comicID, 2.5, 10),
		},
	}, tx)

	_, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
		UserID:     uuid.New(),
		ComicID:    comicID,
		ChapterIDs: []uuid.UUID{chID},
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
	if tx.calls != 0 {
		t.Errorf("tx.calls = %d, want 0", tx.calls)
	}
	if repo.upsertCalls != 0 {
		t.Errorf("repo.upsertCalls = %d, want 0", repo.upsertCalls)
	}
}

//nolint:dupl
func TestMarkReadUnknownChapter(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	ch1 := uuid.New()
	ch2 := uuid.New()
	repo := &fakeRepo{}
	tx := &fakeTransactor{}
	svc := markReadSvc(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, &fakeChapters{
		byID: map[uuid.UUID]chapters.Chapter{
			ch1: ch(ch1, comicID, 1.0, 10),
		},
	}, tx)

	_, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
		UserID:     uuid.New(),
		ComicID:    comicID,
		ChapterIDs: []uuid.UUID{ch1, ch2},
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
	if tx.calls != 0 {
		t.Errorf("tx.calls = %d, want 0", tx.calls)
	}
	if repo.upsertCalls != 0 {
		t.Errorf("repo.upsertCalls = %d, want 0", repo.upsertCalls)
	}
}

//nolint:dupl
func TestMarkReadForeignComicChapter(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	otherComicID := uuid.New()
	chID := uuid.New()
	repo := &fakeRepo{}
	tx := &fakeTransactor{}
	svc := markReadSvc(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, &fakeChapters{
		byID: map[uuid.UUID]chapters.Chapter{
			chID: ch(chID, otherComicID, 3.0, 10),
		},
	}, tx)

	_, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
		UserID:     uuid.New(),
		ComicID:    comicID,
		ChapterIDs: []uuid.UUID{chID},
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
	if tx.calls != 0 {
		t.Errorf("tx.calls = %d, want 0", tx.calls)
	}
	if repo.upsertCalls != 0 {
		t.Errorf("repo.upsertCalls = %d, want 0", repo.upsertCalls)
	}
}

//nolint:dupl
func TestMarkReadOnlyZeroPagesNb(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chID := uuid.New()
	repo := &fakeRepo{}
	tx := &fakeTransactor{}
	svc := markReadSvc(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, &fakeChapters{
		byID: map[uuid.UUID]chapters.Chapter{
			chID: ch(chID, comicID, 4.0, 0),
		},
	}, tx)

	_, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
		UserID:     uuid.New(),
		ComicID:    comicID,
		ChapterIDs: []uuid.UUID{chID},
	})
	if !errors.Is(err, readingprogress.ErrInvalid) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}
	if tx.calls != 0 {
		t.Errorf("tx.calls = %d, want 0", tx.calls)
	}
	if repo.upsertCalls != 0 {
		t.Errorf("repo.upsertCalls = %d, want 0", repo.upsertCalls)
	}
}

func TestMarkReadDedupesBeforeLookup(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chID := uuid.New()
	repo := &fakeRepo{}
	tx := &fakeTransactor{}
	chaps := &fakeChapters{
		byID: map[uuid.UUID]chapters.Chapter{
			chID: ch(chID, comicID, 5.0, 10),
		},
	}
	svc := markReadSvc(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, chaps, tx)

	_, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
		UserID:     uuid.New(),
		ComicID:    comicID,
		ChapterIDs: []uuid.UUID{chID, chID},
	})
	if len(chaps.lastIDs) != 1 {
		t.Fatalf("len(chaps.lastIDs) = %d, want 1", len(chaps.lastIDs))
	}
	if err != nil {
		t.Fatalf("MarkRead: %v", err)
	}
}

func TestMarkReadFirstProgressContinueIsHighestNumber(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicID := uuid.New()
	byID := make(map[uuid.UUID]chapters.Chapter, 10)
	ids := make([]uuid.UUID, 10)
	for i := 0; i < 10; i++ {
		id := uuid.New()
		ids[i] = id
		byID[id] = ch(id, comicID, float64(i+1), 20)
	}

	repo := &fakeRepo{}
	tx := &fakeTransactor{}
	svc := markReadSvc(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, &fakeChapters{byID: byID}, tx)

	got, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
		UserID:     userID,
		ComicID:    comicID,
		ChapterIDs: ids,
	})
	if err != nil {
		t.Fatalf("MarkRead: %v", err)
	}

	if tx.calls != 1 {
		t.Errorf("tx.calls = %d, want 1", tx.calls)
	}
	if repo.upsertCalls < 10 {
		t.Errorf("repo.upsertCalls = %d, want >= 10", repo.upsertCalls)
	}

	seenEligible := make(map[uuid.UUID]bool, len(ids))
	for _, up := range repo.upserts {
		if up.Page == 20 {
			seenEligible[up.ChapterID] = true
		}
	}
	for _, id := range ids {
		if !seenEligible[id] {
			t.Errorf("chapter %v was not upserted with Page 20", id)
		}
	}

	lastUpsert := repo.upserts[len(repo.upserts)-1]
	wantWinnerID := ids[9]
	if lastUpsert.ChapterID != wantWinnerID || lastUpsert.Page != 20 {
		t.Errorf("last upsert = %+v, want chapter %v at page 20", lastUpsert, wantWinnerID)
	}

	if len(repo.upserts) > 1 && !lastUpsert.UpdatedAt.After(repo.upserts[0].UpdatedAt) {
		t.Errorf(
			"last upsert UpdatedAt %v not after first upsert UpdatedAt %v",
			lastUpsert.UpdatedAt,
			repo.upserts[0].UpdatedAt,
		)
	}

	if got.Continue == nil || got.Continue.ChapterID != wantWinnerID || got.Continue.Page != 20 {
		t.Errorf("got.Continue = %+v, want chapter %v at page 20", got.Continue, wantWinnerID)
	}
}

func TestMarkReadKeepsEarlierContinuePage(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicID := uuid.New()
	ch50ID := uuid.New()
	ch50 := ch(ch50ID, comicID, 50.0, 20)

	byID := make(map[uuid.UUID]chapters.Chapter, 41)
	byID[ch50ID] = ch50

	ids := make([]uuid.UUID, 40)
	for i := 0; i < 40; i++ {
		id := uuid.New()
		ids[i] = id
		byID[id] = ch(id, comicID, float64(i+1), 10)
	}

	repo := &fakeRepo{
		latest: &readingprogress.Progress{
			ChapterID: ch50ID,
			Page:      3,
			UpdatedAt: time.Unix(10, 0).UTC(),
		},
	}
	tx := &fakeTransactor{}
	svc := markReadSvc(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, &fakeChapters{byID: byID}, tx)

	got, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
		UserID:     userID,
		ComicID:    comicID,
		ChapterIDs: ids,
	})
	if err != nil {
		t.Fatalf("MarkRead: %v", err)
	}

	if len(repo.upserts) < 2 {
		t.Fatalf("len(repo.upserts) = %d, want at least 2", len(repo.upserts))
	}

	for i, up := range repo.upserts[:len(repo.upserts)-1] {
		if up.ChapterID == ch50ID {
			t.Errorf("upsert[%d] has ch50, want only in last upsert", i)
		}
		if up.Page != 10 {
			t.Errorf("upsert[%d] Page = %d, want 10", i, up.Page)
		}
	}

	lastUpsert := repo.upserts[len(repo.upserts)-1]
	if lastUpsert.ChapterID != ch50ID || lastUpsert.Page != 3 {
		t.Errorf("last upsert = %+v, want ch50 at page 3", lastUpsert)
	}
	if !lastUpsert.UpdatedAt.After(repo.upserts[0].UpdatedAt) {
		t.Errorf("last upsert UpdatedAt not after first upsert UpdatedAt")
	}

	if got.Continue == nil || got.Continue.ChapterID != ch50ID || got.Continue.Page != 3 {
		t.Errorf("got.Continue = %+v, want ch50 at page 3", got.Continue)
	}
}

func TestMarkReadKeepsCompletedLaterContinue(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicID := uuid.New()
	ch100ID := uuid.New()
	ch100 := ch(ch100ID, comicID, 100.0, 50)

	byID := make(map[uuid.UUID]chapters.Chapter, 11)
	byID[ch100ID] = ch100

	ids := make([]uuid.UUID, 10)
	for i := 0; i < 10; i++ {
		id := uuid.New()
		ids[i] = id
		byID[id] = ch(id, comicID, float64(i+1), 20)
	}

	repo := &fakeRepo{
		latest: &readingprogress.Progress{
			ChapterID: ch100ID,
			Page:      50,
			UpdatedAt: time.Unix(10, 0).UTC(),
		},
	}
	tx := &fakeTransactor{}
	svc := markReadSvc(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, &fakeChapters{byID: byID}, tx)

	got, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
		UserID:     userID,
		ComicID:    comicID,
		ChapterIDs: ids,
	})
	if err != nil {
		t.Fatalf("MarkRead: %v", err)
	}

	for i, up := range repo.upserts[:len(repo.upserts)-1] {
		if up.ChapterID == ch100ID {
			t.Errorf("non-final upsert[%d] rewrote ch100", i)
		}
	}

	lastUpsert := repo.upserts[len(repo.upserts)-1]
	if lastUpsert.ChapterID != ch100ID || lastUpsert.Page != 50 {
		t.Errorf("last upsert = %+v, want ch100 at page 50", lastUpsert)
	}
	if got.Continue == nil || got.Continue.ChapterID != ch100ID || got.Continue.Page != 50 {
		t.Errorf("got.Continue = %+v, want ch100 at page 50", got.Continue)
	}
}

func TestMarkReadMovesContinueForward(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicID := uuid.New()
	byID := make(map[uuid.UUID]chapters.Chapter, 50)
	ids := make([]uuid.UUID, 50)
	for i := 0; i < 50; i++ {
		id := uuid.New()
		ids[i] = id
		byID[id] = ch(id, comicID, float64(i+1), 15)
	}

	ch5ID := ids[4]
	ch50ID := ids[49]

	repo := &fakeRepo{
		latest: &readingprogress.Progress{
			ChapterID: ch5ID,
			Page:      2,
			UpdatedAt: time.Unix(10, 0).UTC(),
		},
	}
	tx := &fakeTransactor{}
	svc := markReadSvc(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, &fakeChapters{byID: byID}, tx)

	got, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
		UserID:     userID,
		ComicID:    comicID,
		ChapterIDs: ids,
	})
	if err != nil {
		t.Fatalf("MarkRead: %v", err)
	}

	foundCh5Upsert := false
	for _, up := range repo.upserts[:len(repo.upserts)-1] {
		if up.ChapterID == ch5ID && up.Page == 15 {
			foundCh5Upsert = true
		}
	}
	if !foundCh5Upsert {
		t.Errorf("ch5 was not upserted to page 15 in loop")
	}

	lastUpsert := repo.upserts[len(repo.upserts)-1]
	if lastUpsert.ChapterID != ch50ID || lastUpsert.Page != 15 {
		t.Errorf("last upsert = %+v, want ch50 at page 15", lastUpsert)
	}
	if got.Continue == nil || got.Continue.ChapterID != ch50ID || got.Continue.Page != 15 {
		t.Errorf("got.Continue = %+v, want ch50 at page 15", got.Continue)
	}
}

func TestMarkReadSkipsZeroPagesNbAmongEligible(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicID := uuid.New()
	ch1ID := uuid.New()
	ch2ID := uuid.New()
	ch3ID := uuid.New()

	byID := map[uuid.UUID]chapters.Chapter{
		ch1ID: ch(ch1ID, comicID, 1.0, 0),
		ch2ID: ch(ch2ID, comicID, 2.0, 10),
		ch3ID: ch(ch3ID, comicID, 3.0, 10),
	}

	repo := &fakeRepo{}
	tx := &fakeTransactor{}
	svc := markReadSvc(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, &fakeChapters{byID: byID}, tx)

	got, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
		UserID:     userID,
		ComicID:    comicID,
		ChapterIDs: []uuid.UUID{ch1ID, ch2ID, ch3ID},
	})
	if err != nil {
		t.Fatalf("MarkRead: %v", err)
	}

	for _, up := range repo.upserts {
		if up.ChapterID == ch1ID {
			t.Errorf("ch1 (PagesNb: 0) was upserted")
		}
	}

	lastUpsert := repo.upserts[len(repo.upserts)-1]
	if lastUpsert.ChapterID != ch3ID || lastUpsert.Page != 10 {
		t.Errorf("last upsert = %+v, want ch3 at page 10", lastUpsert)
	}
	if got.Continue == nil || got.Continue.ChapterID != ch3ID || got.Continue.Page != 10 {
		t.Errorf("got.Continue = %+v, want ch3 at page 10", got.Continue)
	}
}

func TestMarkReadDuplicateIDsOneWritePerChapter(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicID := uuid.New()
	ch1ID := uuid.New()

	byID := map[uuid.UUID]chapters.Chapter{
		ch1ID: ch(ch1ID, comicID, 1.0, 8),
	}

	repo := &fakeRepo{}
	tx := &fakeTransactor{}
	svc := markReadSvc(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, &fakeChapters{byID: byID}, tx)

	_, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
		UserID:     userID,
		ComicID:    comicID,
		ChapterIDs: []uuid.UUID{ch1ID, ch1ID},
	})
	if err != nil {
		t.Fatalf("MarkRead: %v", err)
	}

	ch1Upserts := 0
	for _, up := range repo.upserts {
		if up.ChapterID == ch1ID {
			ch1Upserts++
		}
	}
	if ch1Upserts > 2 {
		t.Errorf("ch1 upserts = %d, want <= 2", ch1Upserts)
	}
}

func TestMarkReadMissingContinueChapterTreatsAsNone(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicID := uuid.New()
	missingID := uuid.New()
	ch3ID := uuid.New()
	ch7ID := uuid.New()

	byID := map[uuid.UUID]chapters.Chapter{
		ch3ID: ch(ch3ID, comicID, 3.0, 12),
		ch7ID: ch(ch7ID, comicID, 7.0, 12),
	}

	repo := &fakeRepo{
		latest: &readingprogress.Progress{
			ChapterID: missingID,
			Page:      5,
			UpdatedAt: time.Unix(10, 0).UTC(),
		},
	}
	tx := &fakeTransactor{}
	svc := markReadSvc(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, &fakeChapters{byID: byID}, tx)

	got, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
		UserID:     userID,
		ComicID:    comicID,
		ChapterIDs: []uuid.UUID{ch3ID, ch7ID},
	})
	if err != nil {
		t.Fatalf("MarkRead: %v", err)
	}

	for _, up := range repo.upserts {
		if up.ChapterID == missingID {
			t.Errorf("missing continue chapter %v was upserted", missingID)
		}
	}

	lastUpsert := repo.upserts[len(repo.upserts)-1]
	if lastUpsert.ChapterID != ch7ID || lastUpsert.Page != 12 {
		t.Errorf("last upsert = %+v, want ch7 at page 12", lastUpsert)
	}
	if got.Continue == nil || got.Continue.ChapterID != ch7ID || got.Continue.Page != 12 {
		t.Errorf("got.Continue = %+v, want ch7 at page 12", got.Continue)
	}
}

func TestMarkReadTxErrorNoListSideEffect(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chID := uuid.New()
	byID := map[uuid.UUID]chapters.Chapter{
		chID: ch(chID, comicID, 1.0, 10),
	}

	t.Run("upsert error inside tx", func(t *testing.T) {
		t.Parallel()

		upsertErr := errors.New("db error")
		repo := &fakeRepo{upsertErr: upsertErr}
		tx := &fakeTransactor{}
		svc := markReadSvc(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, &fakeChapters{byID: byID}, tx)

		_, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
			UserID:     uuid.New(),
			ComicID:    comicID,
			ChapterIDs: []uuid.UUID{chID},
		})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, upsertErr) {
			t.Errorf("err = %v, want wrapped %v", err, upsertErr)
		}
	})

	t.Run("withinTx returns error", func(t *testing.T) {
		t.Parallel()

		txErr := errors.New("tx commit failed")
		repo := &fakeRepo{}
		tx := &fakeTransactor{err: txErr}
		svc := markReadSvc(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, &fakeChapters{byID: byID}, tx)

		_, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
			UserID:     uuid.New(),
			ComicID:    comicID,
			ChapterIDs: []uuid.UUID{chID},
		})
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, txErr) {
			t.Errorf("err = %v, want wrapped %v", err, txErr)
		}
	})
}

func TestMarkReadTieBreakByUUID(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicID := uuid.New()
	id1 := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	id2 := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	byID := map[uuid.UUID]chapters.Chapter{
		id1: ch(id1, comicID, 1.0, 10),
		id2: ch(id2, comicID, 1.0, 10),
	}

	repo := &fakeRepo{}
	tx := &fakeTransactor{}
	svc := markReadSvc(t, repo, &fakeLibrary{inLib: true}, &fakeComics{exists: true}, &fakeChapters{byID: byID}, tx)

	got, err := svc.MarkRead(context.Background(), readingprogress.MarkReadOpts{
		UserID:     userID,
		ComicID:    comicID,
		ChapterIDs: []uuid.UUID{id2, id1},
	})
	if err != nil {
		t.Fatalf("MarkRead: %v", err)
	}

	lastUpsert := repo.upserts[len(repo.upserts)-1]
	if lastUpsert.ChapterID != id1 {
		t.Errorf("winner = %v, want %v (smallest UUID)", lastUpsert.ChapterID, id1)
	}
	if got.Continue == nil || got.Continue.ChapterID != id1 {
		t.Errorf("got.Continue = %+v, want chapter %v", got.Continue, id1)
	}
}
