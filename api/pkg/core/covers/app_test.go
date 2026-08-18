// SPDX-License-Identifier: AGPL-3.0-or-later

package covers_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/covers"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/imgcache"
)

func TestNewAppRequiresCache(t *testing.T) {
	t.Parallel()

	app, err := covers.NewApp(covers.AppDeps{})
	if err == nil {
		t.Fatal("NewApp without cache must fail")
	}

	if app != nil {
		t.Error("NewApp returned an app in addition to the error")
	}
}

func TestAppRunStopsWhenContextCanceled(t *testing.T) {
	t.Parallel()

	cache, err := imgcache.New(imgcache.Config{
		Dir:           t.TempDir(),
		FetchFn:       covers.NewFetchFn(nil),
		ErrorCacheTTL: time.Minute,
		MinInterval:   time.Millisecond,
		Logger:        slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("imgcache.New: %v", err)
	}

	app, err := covers.NewApp(covers.AppDeps{Cache: cache})
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Run(ctx)
	}()

	cancel()

	if err := <-errCh; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("Run: %v", err)
	}
}
