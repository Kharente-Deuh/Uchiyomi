// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions_test

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
)

type cleanupRepository struct {
	deleteExpired chan struct{}
	err           error
	mu            sync.Mutex
	deleted       int64
	calls         int
}

func (r *cleanupRepository) Insert(context.Context, sessions.InsertSessionOpts) (*sessions.Session, error) {
	panic("Insert must not be called")
}

func (r *cleanupRepository) GetByTokenHash(context.Context, []byte) (*sessions.Session, *users.User, error) {
	panic("GetByTokenHash must not be called")
}

func (r *cleanupRepository) UpdateExpiry(context.Context, uuid.UUID, time.Time) error {
	panic("UpdateExpiry must not be called")
}

func (r *cleanupRepository) DeleteByTokenHash(context.Context, []byte) error {
	panic("DeleteByTokenHash must not be called")
}

func (r *cleanupRepository) DeleteByUserID(context.Context, uuid.UUID) error {
	panic("DeleteByUserID must not be called")
}

func (r *cleanupRepository) DeleteByUserAndProvider(context.Context, uuid.UUID, uuid.UUID) error {
	panic("DeleteByUserAndProvider must not be called")
}

func (r *cleanupRepository) DeleteByProviderAndSID(context.Context, uuid.UUID, string) error {
	panic("DeleteByProviderAndSID must not be called")
}

func (r *cleanupRepository) DeleteExpired(context.Context, time.Time) (int64, error) {
	r.mu.Lock()
	r.calls++
	started := r.deleteExpired
	r.mu.Unlock()

	if started != nil {
		select {
		case <-started:
		default:
			close(started)
		}
	}

	return r.deleted, r.err
}

func TestAppNewRejectsShortInterval(t *testing.T) {
	t.Parallel()

	app, err := sessions.New(sessions.Config{RemoveExpiredSessionsInterval: time.Second}, sessions.Deps{
		SessionsRepository: &cleanupRepository{},
		Logger:             slog.New(slog.DiscardHandler),
	})
	if err == nil {
		t.Fatal("New with 1s interval must fail")
	}

	if app != nil {
		t.Error("New returned an app in addition to the error")
	}
}

func TestAppRunRemovesExpiredSessions(t *testing.T) {
	t.Parallel()

	repo := &cleanupRepository{deleteExpired: make(chan struct{}), deleted: 3}
	app, err := sessions.New(
		sessions.Config{RemoveExpiredSessionsInterval: time.Minute},
		sessions.Deps{SessionsRepository: repo, Logger: slog.New(slog.DiscardHandler)},
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
	case <-repo.deleteExpired:
	case <-time.After(time.Second):
		t.Fatal("DeleteExpired did not run on startup")
	}

	cancel()

	if err := <-errCh; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("Run: %v", err)
	}

	repo.mu.Lock()
	calls := repo.calls
	repo.mu.Unlock()

	if calls < 1 {
		t.Errorf("DeleteExpired called %d times, want at least 1", calls)
	}
}

func TestAppRunKeepsGoingWhenCleanupFails(t *testing.T) {
	t.Parallel()

	repo := &cleanupRepository{
		err:           errors.New("db down"),
		deleteExpired: make(chan struct{}),
	}
	app, err := sessions.New(
		sessions.Config{RemoveExpiredSessionsInterval: time.Minute},
		sessions.Deps{SessionsRepository: repo, Logger: slog.New(slog.DiscardHandler)},
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
	case <-repo.deleteExpired:
	case <-time.After(time.Second):
		t.Fatal("cleanup did not start")
	}

	cancel()

	if err := <-errCh; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("Run: %v, want context cancel not cleanup error", err)
	}
}
