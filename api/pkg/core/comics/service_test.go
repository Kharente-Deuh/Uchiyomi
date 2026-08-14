// SPDX-License-Identifier: AGPL-3.0-or-later

package comics_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/library"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
)

const (
	testSlug   = "solo-leveling"
	testSource = sources.SourceAsuraScans
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

type fakeComicsRepository struct {
	getByIDErr            error
	getBySourceSlugErr    error
	deleteErr             error
	createErr             error
	getManyErr            error
	getByIDResult         *comics.Comic
	getBySourceSlugResult *comics.Comic
	createResult          *comics.Comic
	getManyResult         []comics.Comic
	lastCreateOpts        comics.CreateComicOpts
	getBySourceSlugCalls  int
	createCalls           int
	getByIDCalls          int
	getManyCalls          int
	deleteCalls           int
	lastDeleteID          uuid.UUID
}

func (f *fakeComicsRepository) GetByID(_ context.Context, _ comics.GetByIDOpts) (*comics.Comic, error) {
	f.getByIDCalls++

	return f.getByIDResult, f.getByIDErr
}

func (f *fakeComicsRepository) GetBySourceSlug(_ context.Context, _ comics.GetBySourceSlugOpts) (*comics.Comic, error) {
	f.getBySourceSlugCalls++

	return f.getBySourceSlugResult, f.getBySourceSlugErr
}

func (f *fakeComicsRepository) Create(_ context.Context, opts comics.CreateComicOpts) (*comics.Comic, error) {
	f.createCalls++
	f.lastCreateOpts = opts

	if f.createErr != nil {
		return nil, f.createErr
	}

	if f.createResult != nil {
		return f.createResult, nil
	}

	return &comics.Comic{
		ID:     uuid.New(),
		Slug:   opts.Slug,
		Source: opts.Source,
		Title:  opts.Title,
		Type:   opts.Type,
		Status: opts.Status,
	}, nil
}

//nolint:lll
func (f *fakeComicsRepository) GetBySlugsAndSource(context.Context, comics.GetBySlugsAndSource) ([]comics.Comic, error) {
	panic("GetBySlugsAndSource must not be called by the comics service")
}

func (f *fakeComicsRepository) Delete(_ context.Context, id uuid.UUID) error {
	f.deleteCalls++
	f.lastDeleteID = id

	return f.deleteErr
}

func (f *fakeComicsRepository) GetMany(_ context.Context, _ comics.GetManyOpts) ([]comics.Comic, error) {
	f.getManyCalls++

	return f.getManyResult, f.getManyErr
}

type fakeLibraryRepository struct {
	createErr   error
	deleteErr   error
	createCalls int
	deleteCalls int
	lastCreate  library.CreateOpts
	lastDelete  library.DeleteOpts
}

func (f *fakeLibraryRepository) Create(_ context.Context, opts library.CreateOpts) (*library.Entry, error) {
	f.createCalls++
	f.lastCreate = opts

	if f.createErr != nil {
		return nil, f.createErr
	}

	return &library.Entry{
		ID:      uuid.New(),
		UserID:  opts.UserID,
		ComicID: opts.ComicID,
	}, nil
}

func (f *fakeLibraryRepository) Delete(_ context.Context, opts library.DeleteOpts) error {
	f.deleteCalls++
	f.lastDelete = opts

	return f.deleteErr
}

type fakeSource struct {
	err      error
	infos    *sources.GetInfosBySlugResponse
	lastSlug string
	calls    int
}

func (f *fakeSource) GetInfosBySlug(_ context.Context, slug string) (*sources.GetInfosBySlugResponse, error) {
	f.calls++
	f.lastSlug = slug

	if f.err != nil {
		return nil, f.err
	}

	return f.infos, nil
}

func validServiceDeps() comics.Deps {
	return comics.Deps{
		ComicsRepository:  &fakeComicsRepository{},
		LibraryRepository: &fakeLibraryRepository{},
		Transactor:        &fakeTransactor{},
		Sources: sources.SourceMap{
			testSource: &fakeSource{
				infos: &sources.GetInfosBySlugResponse{
					Slug:         testSlug,
					Title:        "Solo Leveling",
					Status:       sources.SeriesStatusCompleted,
					Type:         sources.SeriesTypeManhwa,
					ChapterCount: 200,
				},
			},
		},
	}
}

func TestNewServiceRejectsMissingDeps(t *testing.T) {
	t.Parallel()

	svc, err := comics.NewService(comics.Deps{})
	if err == nil {
		t.Fatal("NewService without deps must fail")
	}

	if svc != nil {
		t.Errorf("NewService returned a service (%v) in addition to the error", svc)
	}

	if got := err.Error(); got != "deps.Validate: comics repository is required" {
		t.Errorf("err = %q, want %q", got, "deps.Validate: comics repository is required")
	}
}

func TestCreateReturnsExistingComic(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	existing := &comics.Comic{ID: uuid.New(), Slug: testSlug, Source: testSource}

	comicsRepo := &fakeComicsRepository{
		getBySourceSlugResult: existing,
	}
	source := &fakeSource{}
	deps := validServiceDeps()
	deps.ComicsRepository = comicsRepo
	deps.Sources = sources.SourceMap{testSource: source}

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	got, err := svc.Create(context.Background(), comics.CreateOpts{
		UserID: userID,
		Source: testSource,
		Slug:   testSlug,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if got != existing {
		t.Errorf("Create() = %+v, want existing comic %+v", got, existing)
	}

	if comicsRepo.getBySourceSlugCalls != 1 {
		t.Errorf("GetBySourceSlug called %d times, want 1", comicsRepo.getBySourceSlugCalls)
	}

	if source.calls != 0 {
		t.Errorf("source.GetInfosBySlug called %d times, want 0", source.calls)
	}
}

func TestCreateFetchesSourceAndPersistsLibraryEntry(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicsRepo := &fakeComicsRepository{getBySourceSlugErr: domain.ErrNotFound}
	libraryRepo := &fakeLibraryRepository{}
	tx := &fakeTransactor{}
	source := &fakeSource{
		infos: &sources.GetInfosBySlugResponse{
			Slug:         testSlug,
			Title:        "Solo Leveling",
			Status:       sources.SeriesStatusCompleted,
			Type:         sources.SeriesTypeManhwa,
			ChapterCount: 200,
			Author:       "Chugong",
		},
	}

	deps := validServiceDeps()
	deps.ComicsRepository = comicsRepo
	deps.LibraryRepository = libraryRepo
	deps.Transactor = tx
	deps.Sources = sources.SourceMap{testSource: source}

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	got, err := svc.Create(context.Background(), comics.CreateOpts{
		UserID: userID,
		Source: testSource,
		Slug:   testSlug,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if got.Slug != testSlug || got.Source != testSource {
		t.Errorf("Create() = %+v", got)
	}

	if tx.calls != 1 {
		t.Errorf("WithinTx called %d times, want 1", tx.calls)
	}

	if comicsRepo.createCalls != 1 {
		t.Errorf("ComicsRepository.Create called %d times, want 1", comicsRepo.createCalls)
	}

	if comicsRepo.lastCreateOpts.Slug != testSlug || comicsRepo.lastCreateOpts.Author != "Chugong" {
		t.Errorf("CreateComicOpts = %+v", comicsRepo.lastCreateOpts)
	}

	if libraryRepo.createCalls != 1 {
		t.Errorf("LibraryRepository.Create called %d times, want 1", libraryRepo.createCalls)
	}

	if libraryRepo.lastCreate.UserID != userID || libraryRepo.lastCreate.ComicID != got.ID {
		t.Errorf("library CreateOpts = %+v, comic ID = %s", libraryRepo.lastCreate, got.ID)
	}

	if source.lastSlug != testSlug {
		t.Errorf("source slug = %q, want %q", source.lastSlug, testSlug)
	}
}

func TestCreateUnknownSource(t *testing.T) {
	t.Parallel()

	deps := validServiceDeps()
	deps.ComicsRepository = &fakeComicsRepository{getBySourceSlugErr: domain.ErrNotFound}

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	_, err = svc.Create(context.Background(), comics.CreateOpts{
		UserID: uuid.New(),
		Source: "unknown",
		Slug:   testSlug,
	})
	if err == nil {
		t.Fatal("Create with unknown source must fail")
	}

	if want := "source unknown not found"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
}

func TestCreatePropagatesSourceError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("upstream unavailable")
	source := &fakeSource{err: sentinel}
	deps := validServiceDeps()
	deps.ComicsRepository = &fakeComicsRepository{getBySourceSlugErr: domain.ErrNotFound}
	deps.Sources = sources.SourceMap{testSource: source}

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	_, err = svc.Create(context.Background(), comics.CreateOpts{
		UserID: uuid.New(),
		Source: testSource,
		Slug:   testSlug,
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("Create = %v, original error no longer reachable", err)
	}
}

func TestGetByIDDelegatesToRepository(t *testing.T) {
	t.Parallel()

	comic := &comics.Comic{ID: uuid.New()}
	repo := &fakeComicsRepository{getByIDResult: comic}
	deps := validServiceDeps()
	deps.ComicsRepository = repo

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	got, err := svc.GetByID(context.Background(), comics.GetByIDOpts{
		UserID: uuid.New(),
		ID:     comic.ID,
	})
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got != comic {
		t.Errorf("GetByID() = %+v, want %+v", got, comic)
	}

	if repo.getByIDCalls != 1 {
		t.Errorf("GetByID called %d times, want 1", repo.getByIDCalls)
	}
}

func TestGetManyDelegatesToRepository(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	items := []comics.Comic{{ID: uuid.New(), Slug: testSlug}}
	repo := &fakeComicsRepository{getManyResult: items}
	deps := validServiceDeps()
	deps.ComicsRepository = repo

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	got, err := svc.GetMany(context.Background(), comics.GetManyOpts{UserID: &userID})
	if err != nil {
		t.Fatalf("GetMany: %v", err)
	}

	if len(got) != 1 || got[0].ID != items[0].ID {
		t.Errorf("GetMany() = %+v, want %+v", got, items)
	}
}

func TestDeleteRemovesLibraryEntryAndComic(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	comicID := uuid.New()
	comicsRepo := &fakeComicsRepository{}
	libraryRepo := &fakeLibraryRepository{}
	tx := &fakeTransactor{}

	deps := validServiceDeps()
	deps.ComicsRepository = comicsRepo
	deps.LibraryRepository = libraryRepo
	deps.Transactor = tx

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	err = svc.Delete(context.Background(), comics.DeleteOpts{
		UserID: userID,
		ID:     comicID,
	})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if tx.calls != 1 {
		t.Errorf("WithinTx called %d times, want 1", tx.calls)
	}

	if libraryRepo.deleteCalls != 1 {
		t.Errorf("LibraryRepository.Delete called %d times, want 1", libraryRepo.deleteCalls)
	}

	if libraryRepo.lastDelete.UserID != userID || libraryRepo.lastDelete.ComicID != comicID {
		t.Errorf("library DeleteOpts = %+v", libraryRepo.lastDelete)
	}

	if comicsRepo.deleteCalls != 1 || comicsRepo.lastDeleteID != comicID {
		t.Errorf("ComicsRepository.Delete comic ID = %s, want %s", comicsRepo.lastDeleteID, comicID)
	}
}

func TestDeleteReturnsWhenLibraryEntryMissing(t *testing.T) {
	t.Parallel()

	deps := validServiceDeps()
	deps.LibraryRepository = &fakeLibraryRepository{deleteErr: domain.ErrNotFound}
	deps.ComicsRepository = &fakeComicsRepository{}

	svc, err := comics.NewService(deps)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	err = svc.Delete(context.Background(), comics.DeleteOpts{
		UserID: uuid.New(),
		ID:     uuid.New(),
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete = %v, want domain.ErrNotFound", err)
	}
}
