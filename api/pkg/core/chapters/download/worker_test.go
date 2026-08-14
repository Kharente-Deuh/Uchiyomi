// SPDX-License-Identifier: AGPL-3.0-or-later

package download_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters/download"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

type fakeChaptersRepository struct {
	chapters map[uuid.UUID]chapters.Chapter
}

func (f *fakeChaptersRepository) Create(context.Context, chapters.CreateOpts) (*chapters.Chapter, error) {
	panic("not implemented")
}

func (f *fakeChaptersRepository) CreateMany(context.Context, []chapters.CreateOpts) ([]chapters.Chapter, error) {
	panic("not implemented")
}

func (f *fakeChaptersRepository) ListByComicID(context.Context, uuid.UUID) ([]chapters.Chapter, error) {
	panic("not implemented")
}

func (f *fakeChaptersRepository) GetByID(_ context.Context, id uuid.UUID) (*chapters.Chapter, error) {
	chapter, ok := f.chapters[id]
	if !ok {
		return nil, errors.New("not found")
	}

	return &chapter, nil
}

func (f *fakeChaptersRepository) UpdateDownload(_ context.Context, id uuid.UUID, value int) error {
	chapter, ok := f.chapters[id]
	if !ok {
		return errors.New("not found")
	}

	chapter.Download = value
	f.chapters[id] = chapter

	return nil
}

func (f *fakeChaptersRepository) UpdatePagesNb(_ context.Context, id uuid.UUID, pagesNb int) error {
	chapter, ok := f.chapters[id]
	if !ok {
		return errors.New("not found")
	}

	chapter.PagesNb = pagesNb
	f.chapters[id] = chapter

	return nil
}

type fakeComicsRepository struct {
	comics map[uuid.UUID]comics.Comic
}

func (f *fakeComicsRepository) GetByID(context.Context, comics.GetByIDOpts) (*comics.Comic, error) {
	panic("not implemented")
}

func (f *fakeComicsRepository) FindByID(_ context.Context, id uuid.UUID) (*comics.Comic, error) {
	comic, ok := f.comics[id]
	if !ok {
		return nil, errors.New("not found")
	}

	return &comic, nil
}

func (f *fakeComicsRepository) GetBySourceSlug(context.Context, comics.GetBySourceSlugOpts) (*comics.Comic, error) {
	panic("not implemented")
}

func (f *fakeComicsRepository) FindBySourceSlug(context.Context, comics.FindBySourceSlugOpts) (*comics.Comic, error) {
	panic("not implemented")
}

func (f *fakeComicsRepository) Create(context.Context, comics.CreateComicOpts) (*comics.Comic, error) {
	panic("not implemented")
}

func (f *fakeComicsRepository) GetBySlugsAndSource(context.Context, comics.GetBySlugsAndSource) ([]comics.Comic, error) {
	panic("not implemented")
}

func (f *fakeComicsRepository) Delete(context.Context, uuid.UUID) error {
	panic("not implemented")
}

func (f *fakeComicsRepository) GetMany(context.Context, comics.GetManyOpts) ([]comics.Comic, error) {
	panic("not implemented")
}

type fakeSource struct {
	pageURLs []string
}

func (f *fakeSource) GetInfosBySlug(context.Context, sources.GetInfosBySlugOpts) (*sources.GetInfosBySlugResponse, error) {
	panic("not implemented")
}

func (f *fakeSource) GetChaptersBySlug(context.Context, string) ([]sources.SourceChapter, error) {
	panic("not implemented")
}

func (f *fakeSource) GetPageURLsByChapter(context.Context, sources.GetPageURLsByChapterOpts) ([]string, error) {
	return f.pageURLs, nil
}

func newTestWorkerWithConfig(
	t *testing.T,
	chapter chapters.Chapter,
	comic comics.Comic,
	pageURLs []string,
	httpHandler http.Handler,
	cfg download.Config,
) (*download.Worker, string, *fakeChaptersRepository) {
	t.Helper()

	server := httptest.NewServer(httpHandler)
	t.Cleanup(server.Close)

	for i := range pageURLs {
		pageURLs[i] = server.URL + fmt.Sprintf("/page-%d", i+1)
	}

	if cfg.Dir == "" {
		cfg.Dir = t.TempDir()
	}

	chaptersRepo := &fakeChaptersRepository{chapters: map[uuid.UUID]chapters.Chapter{
		chapter.ID: chapter,
	}}
	comicsRepo := &fakeComicsRepository{comics: map[uuid.UUID]comics.Comic{
		comic.ID: comic,
	}}

	worker, err := download.New(
		cfg,
		download.Deps{
			ChaptersRepository: chaptersRepo,
			ComicsRepository:   comicsRepo,
			Sources: sources.SourceMap{
				comic.Source: &fakeSource{pageURLs: pageURLs},
			},
			HTTPClient: server.Client(),
		},
	)
	if err != nil {
		t.Fatalf("download.New: %v", err)
	}

	return worker, cfg.Dir, chaptersRepo
}

func newTestWorker(
	t *testing.T,
	chapter chapters.Chapter,
	comic comics.Comic,
	pageURLs []string,
	httpHandler http.Handler,
) (*download.Worker, string, *fakeChaptersRepository) {
	t.Helper()

	return newTestWorkerWithConfig(
		t,
		chapter,
		comic,
		pageURLs,
		httpHandler,
		download.Config{
			Dir:       t.TempDir(),
			RateLimit: time.Millisecond,
		},
	)
}

func TestWorkerEnqueueIsIdempotent(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chapterID := uuid.New()
	chapter := chapters.Chapter{
		ID:                chapterID,
		ComicID:           comicID,
		SourceChapterSlug: "chapter-1",
		Number:            1,
		PagesNb:           1,
	}
	comic := comics.Comic{
		ID:     comicID,
		Source: sources.SourceAsuraScans,
		Slug:   "series-slug",
	}

	worker, _, _ := newTestWorker(t, chapter, comic, []string{""}, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = worker.Run(ctx)
	}()

	err := worker.Enqueue(context.Background(), []chapters.Chapter{chapter, chapter})
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	err = worker.Enqueue(context.Background(), []chapters.Chapter{chapter})
	if err != nil {
		t.Fatalf("Enqueue second call: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
}

func TestWorkerRetriesPageUntilExhausted(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chapterID := uuid.New()
	chapter := chapters.Chapter{
		ID:                chapterID,
		ComicID:           comicID,
		SourceChapterSlug: "chapter-1",
		Number:            1,
		PagesNb:           1,
	}
	comic := comics.Comic{
		ID:     comicID,
		Source: sources.SourceAsuraScans,
		Slug:   "series-slug",
	}

	var calls atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	})

	worker, _, chaptersRepo := newTestWorker(t, chapter, comic, []string{""}, handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = worker.Run(ctx)
	}()

	err := worker.Enqueue(context.Background(), []chapters.Chapter{chapter})
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		updated, err := chaptersRepo.GetByID(context.Background(), chapterID)
		if err == nil && updated.Download == -1 {
			if calls.Load() != 3 {
				t.Fatalf("download attempts = %d, want 3", calls.Load())
			}

			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("chapter was not marked as failed")
}

func TestWorkerResumesIncrementally(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chapterID := uuid.New()
	chapter := chapters.Chapter{
		ID:                chapterID,
		ComicID:           comicID,
		SourceChapterSlug: "chapter-1",
		Number:            1,
		PagesNb:           2,
	}
	comic := comics.Comic{
		ID:     comicID,
		Source: sources.SourceAsuraScans,
		Slug:   "series-slug",
	}

	var calls atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("page-2"))
	})

	worker, dir, chaptersRepo := newTestWorker(
		t,
		chapter,
		comic,
		[]string{"", ""},
		handler,
	)

	chapterDir := filepath.Join(dir, comicID.String(), "1")
	if err := os.MkdirAll(chapterDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll: %v", err)
	}

	if err := os.WriteFile(filepath.Join(chapterDir, "001.webp"), []byte("page-1"), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = worker.Run(ctx)
	}()

	err := worker.Enqueue(context.Background(), []chapters.Chapter{chapter})
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		updated, err := chaptersRepo.GetByID(context.Background(), chapterID)
		if err == nil && updated.Download == 100 {
			if calls.Load() != 1 {
				t.Fatalf("download attempts for missing page = %d, want 1", calls.Load())
			}

			if _, err := os.Stat(filepath.Join(chapterDir, "002.webp")); err != nil {
				t.Fatalf("second page not saved: %v", err)
			}

			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("chapter was not completed")
}
