// SPDX-License-Identifier: AGPL-3.0-or-later

//nolint:goconst,lll
package download_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters/download"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

type fakeChaptersService struct {
	enqueueResumableErr   error
	scanEarlyAccessErr    error
	scanDone              chan struct{}
	enqueueResumableCalls int
	scanEarlyAccessCalls  int
}

func (f *fakeChaptersService) CreateAll(context.Context, uuid.UUID, []sources.SourceChapter) ([]chapters.Chapter, error) {
	panic("CreateAll must not be called")
}

func (f *fakeChaptersService) ListByComicID(context.Context, uuid.UUID) ([]chapters.Chapter, error) {
	panic("ListByComicID must not be called")
}

func (f *fakeChaptersService) EnqueueDownloadable(context.Context, []chapters.Chapter) error {
	panic("EnqueueDownloadable must not be called")
}

func (f *fakeChaptersService) EnqueueResumable(context.Context) error {
	f.enqueueResumableCalls++

	return f.enqueueResumableErr
}

func (f *fakeChaptersService) ScanEarlyAccess(context.Context) error {
	f.scanEarlyAccessCalls++

	if f.scanDone != nil {
		close(f.scanDone)
	}

	return f.scanEarlyAccessErr
}

func (f *fakeChaptersService) CleanupComic(context.Context, uuid.UUID, []chapters.Chapter) error {
	panic("CleanupComic must not be called")
}

func (f *fakeChaptersService) RetryDownload(context.Context, chapters.RetryDownloadOpts) error {
	panic("RetryDownload must not be called")
}

func (f *fakeChaptersService) GetByIds(context.Context, chapters.GetByIdsOpts) ([]chapters.Chapter, error) {
	panic("GetByIds must not be called")
}

func (f *fakeChaptersService) ListForLibrary(context.Context, chapters.ListForLibraryOpts) ([]chapters.Chapter, error) {
	panic("ListForLibrary must not be called")
}

func (f *fakeChaptersService) GetForLibrary(context.Context, chapters.GetForLibraryOpts) (*chapters.Chapter, error) {
	panic("GetForLibrary must not be called")
}

func (f *fakeChaptersService) GetDetailForLibrary(
	context.Context, chapters.GetForLibraryOpts,
) (*chapters.ChapterDetail, error) {
	panic("GetDetailForLibrary must not be called")
}

func (f *fakeChaptersService) ServePage(context.Context, chapters.ServePageOpts) (string, string, error) {
	panic("ServePage must not be called")
}

type blockingWorker struct {
	started chan struct{}
}

func (b *blockingWorker) Run(ctx context.Context) error {
	close(b.started)

	<-ctx.Done()

	return context.Canceled
}

func TestAppRunEnqueuesResumableOnBoot(t *testing.T) {
	t.Parallel()

	chaptersSvc := &fakeChaptersService{}
	worker := &blockingWorker{started: make(chan struct{})}

	app, err := download.NewApp(
		download.AppConfig{ScanInterval: time.Hour},
		download.AppDeps{
			Worker:          worker,
			ChaptersService: chaptersSvc,
		},
	)
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Run(ctx)
	}()

	select {
	case <-worker.started:
	case <-time.After(time.Second):
		t.Fatal("worker did not start")
	}

	if chaptersSvc.enqueueResumableCalls != 1 {
		t.Errorf("EnqueueResumable called %d times, want 1", chaptersSvc.enqueueResumableCalls)
	}

	cancel()

	if err := <-errCh; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("Run: %v", err)
	}
}

func TestAppRunScansEarlyAccessOnStartup(t *testing.T) {
	t.Parallel()

	chaptersSvc := &fakeChaptersService{scanDone: make(chan struct{})}
	worker := &blockingWorker{started: make(chan struct{})}

	app, err := download.NewApp(
		download.AppConfig{ScanInterval: time.Hour},
		download.AppDeps{
			Worker:          worker,
			ChaptersService: chaptersSvc,
		},
	)
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Run(ctx)
	}()

	select {
	case <-chaptersSvc.scanDone:
	case <-time.After(time.Second):
		t.Fatal("early-access scan did not run on startup")
	}

	if chaptersSvc.scanEarlyAccessCalls != 1 {
		t.Errorf("ScanEarlyAccess called %d times, want 1 on startup", chaptersSvc.scanEarlyAccessCalls)
	}

	cancel()

	if err := <-errCh; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("Run: %v", err)
	}
}

func TestAppRunReturnsEnqueueResumableError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("enqueue failed")
	chaptersSvc := &fakeChaptersService{enqueueResumableErr: wantErr}
	worker := &blockingWorker{started: make(chan struct{})}

	app, err := download.NewApp(
		download.AppConfig{ScanInterval: time.Hour},
		download.AppDeps{
			Worker:          worker,
			ChaptersService: chaptersSvc,
		},
	)
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}

	err = app.Run(context.Background())
	if err == nil {
		t.Fatal("Run: error expected")
	}

	if !errors.Is(err, wantErr) {
		t.Errorf("Run error = %v, want it to wrap %v", err, wantErr)
	}
}
