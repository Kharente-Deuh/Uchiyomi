// SPDX-License-Identifier: AGPL-3.0-or-later

package refresh_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics/refresh"
)

type fakeComicsService struct {
	refreshErr   error
	started      chan struct{}
	release      chan struct{}
	refreshCalls int
	mu           sync.Mutex
}

func (f *fakeComicsService) Create(context.Context, comics.CreateOpts) (*comics.Comic, error) {
	panic("Create must not be called")
}

func (f *fakeComicsService) GetByID(context.Context, comics.GetByIDOpts) (*comics.Comic, error) {
	panic("GetByID must not be called")
}

func (f *fakeComicsService) GetMany(context.Context, comics.GetManyOpts) ([]comics.Comic, error) {
	panic("GetMany must not be called")
}

func (f *fakeComicsService) Delete(context.Context, comics.DeleteOpts) error {
	panic("Delete must not be called")
}

func (f *fakeComicsService) RefreshChapterLists(ctx context.Context) error {
	f.mu.Lock()
	f.refreshCalls++
	started := f.started
	release := f.release
	f.mu.Unlock()

	if started != nil {
		select {
		case <-started:
		default:
			close(started)
		}
	}

	if release != nil {
		select {
		case <-release:
		case <-ctx.Done():
			return fmt.Errorf("refresh interrupted: %w", ctx.Err())
		}
	}

	return f.refreshErr
}

func (f *fakeComicsService) ServeCover(context.Context, comics.GetByIDOpts) (string, string, error) {
	panic("ServeCover must not be called")
}

func TestNewRejectsShortInterval(t *testing.T) {
	t.Parallel()

	app, err := refresh.New(refresh.Config{Interval: time.Second}, refresh.Deps{
		ComicsService: &fakeComicsService{},
	})
	if err == nil {
		t.Fatal("New with 1s interval must fail")
	}

	if app != nil {
		t.Error("New returned an app in addition to the error")
	}
}

func TestRunRefreshesOnStartup(t *testing.T) {
	t.Parallel()

	svc := &fakeComicsService{started: make(chan struct{})}
	app, err := refresh.New(
		refresh.Config{Interval: time.Hour},
		refresh.Deps{ComicsService: svc},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Run(ctx)
	}()

	select {
	case <-svc.started:
	case <-time.After(time.Second):
		t.Fatal("RefreshChapterLists did not run on startup")
	}

	cancel()

	if err := <-errCh; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("Run: %v", err)
	}

	svc.mu.Lock()
	calls := svc.refreshCalls
	svc.mu.Unlock()

	if calls != 1 {
		t.Errorf("RefreshChapterLists called %d times, want 1", calls)
	}
}

func TestRunDoesNotAbortWhenRefreshFails(t *testing.T) {
	t.Parallel()

	svc := &fakeComicsService{
		refreshErr: errors.New("source down"),
		started:    make(chan struct{}),
	}
	app, err := refresh.New(
		refresh.Config{Interval: time.Hour},
		refresh.Deps{ComicsService: svc},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Run(ctx)
	}()

	select {
	case <-svc.started:
	case <-time.After(time.Second):
		t.Fatal("refresh did not start")
	}

	cancel()

	if err := <-errCh; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("Run: %v, want context cancel not refresh error", err)
	}
}

func TestSkipIfAlreadyRunning(t *testing.T) {
	t.Parallel()

	svc := &fakeComicsService{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	app, err := refresh.New(
		refresh.Config{Interval: time.Hour},
		refresh.Deps{ComicsService: svc},
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 2)

	go func() { errCh <- app.Run(ctx) }()

	select {
	case <-svc.started:
	case <-time.After(time.Second):
		t.Fatal("first refresh did not start")
	}

	go func() { errCh <- app.Run(ctx) }()

	time.Sleep(50 * time.Millisecond)

	close(svc.release)
	cancel()

	<-errCh
	<-errCh

	svc.mu.Lock()
	calls := svc.refreshCalls
	svc.mu.Unlock()

	if calls != 1 {
		t.Errorf("RefreshChapterLists called %d times, want 1 (second run skipped)", calls)
	}
}
