// SPDX-License-Identifier: AGPL-3.0-or-later

package fncache_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/fncache"
)

const (
	testCacheName = "test"
	callTimeout   = 2 * time.Second
)

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func callWithTimeout(t *testing.T, what string, fn func()) {
	t.Helper()

	done := make(chan struct{})

	go func() {
		defer close(done)

		fn()
	}()

	select {
	case <-done:
	case <-time.After(callTimeout):
		t.Fatalf("%s never returned after %v: deadlock", what, callTimeout)
	}
}

func newTestCache[P any, T any](t *testing.T, fn func(context.Context, P) (*T, error)) *fncache.Cache[P, T] {
	t.Helper()

	c, err := fncache.New(fncache.Config[P, T]{
		Name:          testCacheName,
		Fn:            fn,
		Key:           func(p P) string { return fmt.Sprintf("%v", p) },
		TTL:           time.Hour,
		ErrorTTL:      time.Hour,
		FetchTimeout:  time.Minute,
		CleanInterval: time.Hour,
		MaxEntries:    128,
	}, fncache.Deps{Logger: discardLogger()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return c
}

func TestFnCacheConfigValidate(t *testing.T) {
	t.Parallel()

	fn := func(context.Context, string) (*int, error) { return nil, nil }
	key := func(s string) string { return s }

	valid := fncache.Config[string, int]{
		Name:          testCacheName,
		Fn:            fn,
		Key:           key,
		TTL:           time.Minute,
		ErrorTTL:      time.Minute,
		FetchTimeout:  time.Minute,
		CleanInterval: time.Minute,
		MaxEntries:    16,
	}

	without := func(mutate func(*fncache.Config[string, int])) fncache.Config[string, int] {
		cfg := valid
		mutate(&cfg)

		return cfg
	}

	tests := map[string]struct {
		wantErr string
		cfg     fncache.Config[string, int]
	}{
		"valide": {cfg: valid},
		"TTL exactly one second": {
			cfg: without(func(c *fncache.Config[string, int]) { c.TTL = time.Second }),
		},
		"without Fn": {
			cfg:     without(func(c *fncache.Config[string, int]) { c.Fn = nil }),
			wantErr: "fn is required",
		},
		"without Key": {
			cfg:     without(func(c *fncache.Config[string, int]) { c.Key = nil }),
			wantErr: "key is required",
		},
		"zero TTL": {
			cfg:     without(func(c *fncache.Config[string, int]) { c.TTL = 0 }),
			wantErr: "ttl must be at least 1 second",
		},
		"zero ErrorTTL": {
			cfg:     without(func(c *fncache.Config[string, int]) { c.ErrorTTL = 0 }),
			wantErr: "errorTTL must be at least 1 second",
		},
		"FetchTimeout nul": {
			cfg:     without(func(c *fncache.Config[string, int]) { c.FetchTimeout = 0 }),
			wantErr: "fetchTimeout must be at least 1 second",
		},
		"zero CleanInterval": {
			cfg:     without(func(c *fncache.Config[string, int]) { c.CleanInterval = 0 }),
			wantErr: "cleanInterval must be at least 1 second",
		},
		"MaxEntries nul": {
			cfg:     without(func(c *fncache.Config[string, int]) { c.MaxEntries = 0 }),
			wantErr: "maxEntries must be greater than 0",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := tc.cfg
			err := cfg.Validate()

			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}

				return
			}

			if err == nil || err.Error() != tc.wantErr {
				t.Errorf("Validate() = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestNewFnCacheRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	c, err := fncache.New(fncache.Config[string, int]{}, fncache.Deps{Logger: discardLogger()})
	if err == nil {
		t.Fatal("New must reject empty config")
	}

	if c != nil {
		t.Error("New returned a cache in addition to the error")
	}
}

func TestFnCacheGetReturnsValue(t *testing.T) {
	t.Parallel()

	want := 42
	c := newTestCache(t, func(context.Context, string) (*int, error) { return &want, nil })

	var (
		got *int
		err error
	)

	callWithTimeout(t, "Cache.Get", func() {
		got, err = c.Get(context.Background(), "key")
	})

	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got == nil || *got != want {
		t.Errorf("Get() = %v, want %d", got, want)
	}
}

func TestFnCacheGetCachesResult(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	value := 7
	c := newTestCache(t, func(context.Context, string) (*int, error) {
		calls.Add(1)

		return &value, nil
	})

	for i := range 3 {
		callWithTimeout(t, "Cache.Get", func() {
			if _, err := c.Get(context.Background(), "same-key"); err != nil {
				t.Errorf("Get call %d: %v", i+1, err)
			}
		})
	}

	if got := calls.Load(); got != 1 {
		t.Errorf("Fn called %d times for 3 Get on same key (TTL = 1h), want 1", got)
	}
}

func TestFnCacheGetDistinctKeys(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	c := newTestCache(t, func(_ context.Context, p string) (*string, error) {
		calls.Add(1)
		out := "v:" + p

		return &out, nil
	})

	for _, key := range []string{"a", "b"} {
		callWithTimeout(t, "Cache.Get", func() {
			got, err := c.Get(context.Background(), key)
			if err != nil {
				t.Errorf("Get(%q): %v", key, err)

				return
			}

			if want := "v:" + key; got == nil || *got != want {
				t.Errorf("Get(%q) = %v, want %q", key, got, want)
			}
		})
	}

	if got := calls.Load(); got != 2 {
		t.Errorf("Fn called %d times for 2 distinct keys, want 2", got)
	}
}

func TestFnCacheGetPropagatesError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("upstream down")
	c := newTestCache(t, func(context.Context, string) (*int, error) { return nil, sentinel })

	var err error

	callWithTimeout(t, "Cache.Get", func() {
		_, err = c.Get(context.Background(), "key")
	})

	if !errors.Is(err, sentinel) {
		t.Errorf("Get() = %v, want %v", err, sentinel)
	}
}

func waitForCalls(t *testing.T, calls *atomic.Int32, want int32) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)

	for calls.Load() < want {
		if time.Now().After(deadline) {
			t.Fatalf("Fn called %d times after 2s, want %d", calls.Load(), want)
		}

		time.Sleep(time.Millisecond)
	}
}

func TestFnCacheGetDedupesConcurrentCalls(t *testing.T) {
	t.Parallel()

	const goroutines = 8

	var calls atomic.Int32

	release := make(chan struct{})
	value := 1

	c := newTestCache(t, func(context.Context, string) (*int, error) {
		calls.Add(1)
		<-release

		return &value, nil
	})

	var wg sync.WaitGroup

	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()

			if _, err := c.Get(context.Background(), "k"); err != nil {
				t.Errorf("Get: %v", err)
			}
		}()
	}

	waitForCalls(t, &calls, 1)
	close(release)
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Errorf("Fn called %d times for %d concurrent Get on same key, want 1", got, goroutines)
	}
}

func TestFnCacheGetDetachesFnFromCallerContext(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})
	fnCtxErr := make(chan error, 1)
	value := 1

	c := newTestCache(t, func(ctx context.Context, _ string) (*int, error) {
		close(started)
		time.Sleep(50 * time.Millisecond)
		fnCtxErr <- ctx.Err()

		return &value, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-started
		cancel()
	}()

	var err error

	callWithTimeout(t, "Cache.Get", func() {
		_, err = c.Get(ctx, "k")
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Get() = %v, want context.Canceled (caller must be able to abandon)", err)
	}

	if got := <-fnCtxErr; got != nil {
		t.Errorf("ctx.Err() in Fn = %v, want nil (Fn must survive caller abandonment)", got)
	}
}

func TestFnCacheGetConvertsPanicToError(t *testing.T) {
	t.Parallel()

	c := newTestCache(t, func(context.Context, string) (*int, error) { panic("boom") })

	var err error

	callWithTimeout(t, "Cache.Get", func() {
		_, err = c.Get(context.Background(), "k")
	})

	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("Get() = %v, want error mentioning panic", err)
	}
}

func TestFnCacheGetBoundsFnByFetchTimeout(t *testing.T) {
	t.Parallel()

	fnCtxErr := make(chan error, 1)

	c, err := fncache.New(fncache.Config[string, int]{
		Name: testCacheName,
		Fn: func(ctx context.Context, _ string) (*int, error) {
			<-ctx.Done()
			fnCtxErr <- ctx.Err()

			return nil, ctx.Err()
		},
		Key:           func(s string) string { return s },
		TTL:           time.Hour,
		ErrorTTL:      time.Hour,
		FetchTimeout:  time.Second,
		CleanInterval: time.Hour,
		MaxEntries:    128,
	}, fncache.Deps{Logger: discardLogger()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	var getErr error

	callWithTimeout(t, "Cache.Get", func() {
		_, getErr = c.Get(context.Background(), "k")
	})

	if !errors.Is(getErr, context.DeadlineExceeded) {
		t.Errorf("Get() = %v, want context.DeadlineExceeded", getErr)
	}

	if fnErr := <-fnCtxErr; !errors.Is(fnErr, context.DeadlineExceeded) {
		t.Errorf("ctx.Err() in Fn = %v, want context.DeadlineExceeded (FetchTimeout must bound Fn)", fnErr)
	}
}

func TestFnCacheRunStopsWithContext(t *testing.T) {
	t.Parallel()

	value := 1
	c := newTestCache(t, func(context.Context, string) (*int, error) { return &value, nil })

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var err error

	callWithTimeout(t, "Cache.Run", func() {
		err = c.Run(ctx)
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Run() = %v, want context.DeadlineExceeded", err)
	}
}
