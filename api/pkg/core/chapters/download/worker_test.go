// SPDX-License-Identifier: AGPL-3.0-or-later

//nolint:goconst,lll,dupl
package download_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"sync"
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
	mu       sync.Mutex
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

func (f *fakeChaptersRepository) ListResumable(context.Context, time.Time) ([]chapters.Chapter, error) {
	panic("not implemented")
}

func (f *fakeChaptersRepository) ListEarlyAccessUnlocked(context.Context, time.Time) ([]chapters.Chapter, error) {
	panic("not implemented")
}

func (f *fakeChaptersRepository) GetByIds(context.Context, []uuid.UUID) ([]chapters.Chapter, error) {
	panic("GetByIds must not be called")
}

func (f *fakeChaptersRepository) GetByID(_ context.Context, id uuid.UUID) (*chapters.Chapter, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	chapter, ok := f.chapters[id]
	if !ok {
		return nil, errors.New("not found")
	}

	return &chapter, nil
}

func (f *fakeChaptersRepository) UpdateDownload(_ context.Context, id uuid.UUID, value int) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	chapter, ok := f.chapters[id]
	if !ok {
		return errors.New("not found")
	}

	chapter.Download = value
	f.chapters[id] = chapter

	return nil
}

func (f *fakeChaptersRepository) UpdatePagesNb(_ context.Context, id uuid.UUID, pagesNb int) error {
	f.mu.Lock()
	defer f.mu.Unlock()

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

func (f *fakeComicsRepository) GetMany(context.Context, comics.GetManyOpts) (comics.Page, error) {
	panic("not implemented")
}

func (f *fakeComicsRepository) ListByStatuses(context.Context, comics.ListByStatusesOpts) ([]comics.Comic, error) {
	panic("not implemented")
}

func (f *fakeComicsRepository) UpdateStatusAndChapterCount(context.Context, comics.UpdateStatusAndChapterCountOpts) error {
	panic("not implemented")
}

type fakeSource struct {
	pageURLs []string
}

func (f *fakeSource) GetInfosBySlug(context.Context, sources.GetInfosBySlugOpts) (*sources.GetInfosBySlugResponse, error) {
	panic("not implemented")
}

func (f *fakeSource) GetChaptersBySlug(context.Context, sources.GetChaptersBySlugOpts) ([]sources.SourceChapter, error) {
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
		ext := path.Ext(pageURLs[i])
		pageURLs[i] = server.URL + fmt.Sprintf("/page-%d%s", i+1, ext)
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

	worker, _, _ := newTestWorker(
		t,
		chapter,
		comic,
		[]string{""},
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

func TestWorkerCleanupComicDeletesDownloadDir(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chapterID := uuid.New()
	chapter := chapters.Chapter{
		ID:      chapterID,
		ComicID: comicID,
		Number:  1,
	}
	comic := comics.Comic{
		ID:     comicID,
		Source: sources.SourceAsuraScans,
		Slug:   "series-slug",
	}

	worker, dir, _ := newTestWorker(
		t,
		chapter,
		comic,
		[]string{""},
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	chapterDir := filepath.Join(dir, comicID.String(), "1")
	if err := os.MkdirAll(chapterDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll: %v", err)
	}

	if err := os.WriteFile(filepath.Join(chapterDir, "001.webp"), []byte("page-1"), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}

	err := worker.Enqueue(context.Background(), []chapters.Chapter{chapter})
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	err = worker.CleanupComic(context.Background(), comicID, []chapters.Chapter{chapter})
	if err != nil {
		t.Fatalf("CleanupComic: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, comicID.String())); !os.IsNotExist(err) {
		t.Fatalf("comic download dir still exists: %v", err)
	}
}

func TestWorkerCleanupComicStopsInFlightDownloads(t *testing.T) {
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

	downloadStarted := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		select {
		case downloadStarted <- struct{}{}:
		default:
		}

		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("page"))
	})

	worker, dir, chaptersRepo := newTestWorker(
		t,
		chapter,
		comic,
		[]string{"", ""},
		handler,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = worker.Run(ctx)
	}()

	err := worker.Enqueue(context.Background(), []chapters.Chapter{chapter})
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	select {
	case <-downloadStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("download did not start")
	}

	err = worker.CleanupComic(context.Background(), comicID, []chapters.Chapter{chapter})
	if err != nil {
		t.Fatalf("CleanupComic: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, comicID.String())); !os.IsNotExist(err) {
		t.Fatalf("comic download dir still exists after cleanup: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		updated, err := chaptersRepo.GetByID(context.Background(), chapterID)
		if err == nil && updated.Download == 100 {
			t.Fatal("chapter was marked completed after cleanup")
		}

		if _, err := os.Stat(filepath.Join(dir, comicID.String())); err == nil {
			t.Fatal("comic download dir was recreated by in-flight worker")
		}

		time.Sleep(20 * time.Millisecond)
	}
}

func TestWorkerResetAndEnqueueClearsPagesAndResetsProgress(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chapterID := uuid.New()
	chapter := chapters.Chapter{
		ID:                chapterID,
		ComicID:           comicID,
		SourceChapterSlug: "chapter-1",
		Number:            1,
		PagesNb:           1,
		Download:          -1,
	}
	comic := comics.Comic{
		ID:     comicID,
		Source: sources.SourceAsuraScans,
		Slug:   "series-slug",
	}

	worker, dir, chaptersRepo := newTestWorker(
		t,
		chapter,
		comic,
		[]string{""},
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}),
	)

	chapterPath := filepath.Join(dir, comicID.String(), "1")
	if err := os.MkdirAll(chapterPath, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	pagePath := filepath.Join(chapterPath, "001.webp")
	if err := os.WriteFile(pagePath, []byte("page"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := worker.ResetAndEnqueue(context.Background(), chapterID); err != nil {
		t.Fatalf("ResetAndEnqueue: %v", err)
	}

	if _, err := os.Stat(pagePath); !os.IsNotExist(err) {
		t.Errorf("page still exists after ResetAndEnqueue: %v", err)
	}

	updated, err := chaptersRepo.GetByID(context.Background(), chapterID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if updated.Download != 0 {
		t.Errorf("download = %d, want 0", updated.Download)
	}
}

func TestWorkerResumeEnqueuesWithoutClearingPages(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chapterID := uuid.New()
	chapter := chapters.Chapter{
		ID:                chapterID,
		ComicID:           comicID,
		SourceChapterSlug: "chapter-1",
		Number:            1,
		PagesNb:           1,
		Download:          42,
	}
	comic := comics.Comic{
		ID:     comicID,
		Source: sources.SourceAsuraScans,
		Slug:   "series-slug",
	}

	worker, dir, chaptersRepo := newTestWorker(
		t,
		chapter,
		comic,
		[]string{""},
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}),
	)

	chapterPath := filepath.Join(dir, comicID.String(), "1")
	if err := os.MkdirAll(chapterPath, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	pagePath := filepath.Join(chapterPath, "001.webp")
	if err := os.WriteFile(pagePath, []byte("page"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := worker.Resume(context.Background(), chapterID); err != nil {
		t.Fatalf("Resume: %v", err)
	}

	if _, err := os.Stat(pagePath); err != nil {
		t.Errorf("page must remain after Resume: %v", err)
	}

	updated, err := chaptersRepo.GetByID(context.Background(), chapterID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if updated.Download != 42 {
		t.Errorf("download = %d, want 42", updated.Download)
	}
}

func TestWorkerPreservesPNGBytes(t *testing.T) {
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

	pngBytes := createTestPNG(t, 200, 200)

	worker, dir, chaptersRepo := newTestWorker(
		t,
		chapter,
		comic,
		[]string{".png"},
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(pngBytes)
		}),
	)

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
			chapterDir := filepath.Join(dir, comicID.String(), "1")
			pngPath := filepath.Join(chapterDir, "001.png")
			webpPath := filepath.Join(chapterDir, "001.webp")

			data, err := os.ReadFile(pngPath)
			if err != nil {
				t.Fatalf("expected 001.png to exist: %v", err)
			}

			if !bytes.Equal(data, pngBytes) {
				t.Errorf("expected original png bytes to be preserved exactly")
			}

			if _, err := os.Stat(webpPath); !os.IsNotExist(err) {
				t.Errorf("expected 001.webp not to exist")
			}

			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("chapter was not completed")
}

func TestWorkerPreservesJPEGBytes(t *testing.T) {
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

	jpegBytes := createTestJPEG(t, 200, 200)

	worker, dir, chaptersRepo := newTestWorker(
		t,
		chapter,
		comic,
		[]string{".jpg"},
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(jpegBytes)
		}),
	)

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
			chapterDir := filepath.Join(dir, comicID.String(), "1")
			jpgPath := filepath.Join(chapterDir, "001.jpg")
			webpPath := filepath.Join(chapterDir, "001.webp")

			data, err := os.ReadFile(jpgPath)
			if err != nil {
				t.Fatalf("expected 001.jpg to exist: %v", err)
			}

			if !bytes.Equal(data, jpegBytes) {
				t.Errorf("expected original jpeg bytes to be preserved")
			}

			if _, err := os.Stat(webpPath); !os.IsNotExist(err) {
				t.Errorf("expected 001.webp not to exist")
			}

			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("chapter was not completed")
}

func TestWorkerResumesSkippingExistingWebPAndPNG(t *testing.T) {
	t.Parallel()

	comicID := uuid.New()
	chapterID := uuid.New()
	chapter := chapters.Chapter{
		ID:                chapterID,
		ComicID:           comicID,
		SourceChapterSlug: "chapter-1",
		Number:            1,
		PagesNb:           3,
	}
	comic := comics.Comic{
		ID:     comicID,
		Source: sources.SourceAsuraScans,
		Slug:   "series-slug",
	}

	page3PNG := createTestPNG(t, 200, 200)

	var calls atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(page3PNG)
	})

	worker, dir, chaptersRepo := newTestWorker(
		t,
		chapter,
		comic,
		[]string{".png", ".png", ".png"},
		handler,
	)

	chapterDir := filepath.Join(dir, comicID.String(), "1")
	if err := os.MkdirAll(chapterDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll: %v", err)
	}

	existingPage1 := []byte("existing-webp-page1")
	existingPage2 := []byte("existing-png-page2")

	if err := os.WriteFile(filepath.Join(chapterDir, "001.webp"), existingPage1, 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(chapterDir, "002.png"), existingPage2, 0o644); err != nil {
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

			// Verify existing files were untouched
			p1, err := os.ReadFile(filepath.Join(chapterDir, "001.webp"))
			if err != nil || !bytes.Equal(p1, existingPage1) {
				t.Fatalf("001.webp modified or missing: %v", err)
			}

			p2, err := os.ReadFile(filepath.Join(chapterDir, "002.png"))
			if err != nil || !bytes.Equal(p2, existingPage2) {
				t.Fatalf("002.png modified or missing: %v", err)
			}

			// Verify page 3 was downloaded and saved as png
			p3, err := os.ReadFile(filepath.Join(chapterDir, "003.png"))
			if err != nil || !bytes.Equal(p3, page3PNG) {
				t.Fatalf("third page not saved as png: %v", err)
			}

			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("chapter was not completed")
}
