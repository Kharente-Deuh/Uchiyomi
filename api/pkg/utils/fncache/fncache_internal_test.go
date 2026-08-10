// SPDX-License-Identifier: AGPL-3.0-or-later

package fncache

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"
)

func newInternalTestCache(t *testing.T, fn func(context.Context, string) (*int, error)) *Cache[string, int] {
	t.Helper()

	c, err := New(Config[string, int]{
		Name:          "test",
		Fn:            fn,
		Key:           func(s string) string { return s },
		TTL:           time.Hour,
		ErrorTTL:      time.Hour,
		FetchTimeout:  time.Minute,
		CleanInterval: time.Hour,
		MaxEntries:    128,
	}, Deps{Logger: slog.New(slog.DiscardHandler)})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return c
}

func TestFnCacheGetRefetchesAfterTTL(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	value := 1
	c := newInternalTestCache(t, func(context.Context, string) (*int, error) {
		calls.Add(1)

		return &value, nil
	})

	if _, err := c.Get(context.Background(), "k"); err != nil {
		t.Fatalf("Get: %v", err)
	}

	c.mtx.Lock()
	entry := c.store[`k`]
	entry.FetchedAt = time.Now().Add(-2 * time.Hour)
	c.store[`k`] = entry
	c.mtx.Unlock()

	if _, err := c.Get(context.Background(), "k"); err != nil {
		t.Fatalf("Get (after expiration): %v", err)
	}

	if got := calls.Load(); got != 2 {
		t.Errorf("Fn called %d times, want 2 (expired entry must be reloaded)", got)
	}
}

func TestFnCacheCleanStoreDropsOldEntries(t *testing.T) {
	t.Parallel()

	value := 1
	c := newInternalTestCache(t, func(context.Context, string) (*int, error) { return &value, nil })

	if _, err := c.Get(context.Background(), "k"); err != nil {
		t.Fatalf("Get: %v", err)
	}

	c.mtx.Lock()
	entry := c.store[`k`]
	entry.FetchedAt = time.Now().Add(-2 * time.Hour)
	c.store[`k`] = entry
	c.mtx.Unlock()

	if err := c.cleanStore(context.Background()); err != nil {
		t.Fatalf("cleanStore: %v", err)
	}

	c.mtx.Lock()
	_, still := c.store[`k`]
	c.mtx.Unlock()

	if still {
		t.Error("expired entry still in store after cleanStore")
	}
}

func TestFnCacheGetUsesErrorTTLForFailures(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	sentinel := errors.New("upstream down")
	c := newInternalTestCache(t, func(context.Context, string) (*int, error) {
		calls.Add(1)

		return nil, sentinel
	})

	c.cfg.ErrorTTL = time.Second

	if _, err := c.Get(context.Background(), "k"); !errors.Is(err, sentinel) {
		t.Fatalf("Get: %v", err)
	}

	c.mtx.Lock()
	entry := c.store["k"]
	entry.FetchedAt = time.Now().Add(-2 * time.Second)
	c.store["k"] = entry
	c.mtx.Unlock()

	if _, err := c.Get(context.Background(), "k"); !errors.Is(err, sentinel) {
		t.Fatalf("Get (after ErrorTTL): %v", err)
	}

	if got := calls.Load(); got != 2 {
		t.Errorf("Fn called %d times, want 2 (error expires per ErrorTTL)", got)
	}
}

func TestFnCacheGetDoesNotCacheContextCanceled(t *testing.T) {
	t.Parallel()

	c := newInternalTestCache(t, func(context.Context, string) (*int, error) {
		return nil, context.Canceled
	})

	if _, err := c.Get(context.Background(), "k"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Get: %v", err)
	}

	c.mtx.Lock()
	_, stored := c.store["k"]
	c.mtx.Unlock()

	if stored {
		t.Error("context.Canceled was stored")
	}
}

func TestFnCacheGetCachesContextDeadlineExceeded(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	c := newInternalTestCache(t, func(context.Context, string) (*int, error) {
		calls.Add(1)

		return nil, context.DeadlineExceeded
	})

	if _, err := c.Get(context.Background(), "k"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Get: %v", err)
	}

	if _, err := c.Get(context.Background(), "k"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Get (2nd call): %v", err)
	}

	if got := calls.Load(); got != 1 {
		t.Errorf("Fn called %d times, want 1 (DeadlineExceeded must be cached)", got)
	}
}

func TestFnCacheCleanStoreKeepsFreshEntriesRegardlessOfInterval(t *testing.T) {
	t.Parallel()

	value := 1
	c := newInternalTestCache(t, func(context.Context, string) (*int, error) { return &value, nil })

	c.cfg.CleanInterval = time.Second

	if _, err := c.Get(context.Background(), "frais"); err != nil {
		t.Fatalf("Get: %v", err)
	}

	if err := c.cleanStore(context.Background()); err != nil {
		t.Fatalf("cleanStore: %v", err)
	}

	c.mtx.Lock()
	_, still := c.store["frais"]
	c.mtx.Unlock()

	if !still {
		t.Error("cleanStore removed entry still within TTL")
	}
}

func TestFnCacheCleanStoreDropsExpiredErrorEntries(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("upstream down")
	c := newInternalTestCache(t, func(context.Context, string) (*int, error) { return nil, sentinel })

	c.cfg.ErrorTTL = time.Second

	if _, err := c.Get(context.Background(), "k"); !errors.Is(err, sentinel) {
		t.Fatalf("Get: %v", err)
	}

	c.mtx.Lock()
	entry := c.store["k"]
	entry.FetchedAt = time.Now().Add(-2 * time.Second)
	c.store["k"] = entry
	c.mtx.Unlock()

	if err := c.cleanStore(context.Background()); err != nil {
		t.Fatalf("cleanStore: %v", err)
	}

	c.mtx.Lock()
	_, still := c.store["k"]
	c.mtx.Unlock()

	if still {
		t.Error("cleanStore kept error entry expired per ErrorTTL")
	}
}

func TestFnCacheGetEnforcesMaxEntries(t *testing.T) {
	t.Parallel()

	value := 1
	c := newInternalTestCache(t, func(context.Context, string) (*int, error) { return &value, nil })

	c.cfg.MaxEntries = 2

	for _, key := range []string{"a", "b"} {
		if _, err := c.Get(context.Background(), key); err != nil {
			t.Fatalf("Get(%q): %v", key, err)
		}
	}

	c.mtx.Lock()
	entry := c.store["a"]
	entry.FetchedAt = time.Now().Add(-time.Minute)
	c.store["a"] = entry
	c.mtx.Unlock()

	if _, err := c.Get(context.Background(), "c"); err != nil {
		t.Fatalf("Get(\"c\"): %v", err)
	}

	c.mtx.Lock()
	size := len(c.store)
	_, hasA := c.store["a"]
	_, hasC := c.store["c"]
	c.mtx.Unlock()

	if size != 2 {
		t.Errorf("len(store) = %d, want 2 (MaxEntries)", size)
	}

	if hasA {
		t.Error("oldest entry still present")
	}

	if !hasC {
		t.Error("last inserted entry was evicted")
	}
}

func TestFnCacheGetRefreshDoesNotEvict(t *testing.T) {
	t.Parallel()

	value := 1
	c := newInternalTestCache(t, func(context.Context, string) (*int, error) { return &value, nil })

	c.cfg.MaxEntries = 2

	for _, key := range []string{"a", "b"} {
		if _, err := c.Get(context.Background(), key); err != nil {
			t.Fatalf("Get(%q): %v", key, err)
		}
	}

	c.mtx.Lock()
	for key, age := range map[string]time.Duration{"a": 3 * time.Hour, "b": 2 * time.Hour} {
		entry := c.store[key]
		entry.FetchedAt = time.Now().Add(-age)
		c.store[key] = entry
	}
	c.mtx.Unlock()

	if _, err := c.Get(context.Background(), "b"); err != nil {
		t.Fatalf("Get(\"b\") after expiration: %v", err)
	}

	c.mtx.Lock()
	size := len(c.store)
	_, hasA := c.store["a"]
	c.mtx.Unlock()

	if size != 2 {
		t.Errorf("len(store) = %d, want 2", size)
	}

	if !hasA {
		t.Error("reloading existing key evicted oldest entry")
	}
}

func TestFnCacheGetEvictsStaleEntryOverOlderLiveOne(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("upstream down")
	value := 1

	responses := map[string]struct {
		val *int
		err error
	}{
		"ok": {val: &value},
		"ko": {err: sentinel},
	}

	c := newInternalTestCache(t, func(_ context.Context, key string) (*int, error) {
		r := responses[key]

		return r.val, r.err
	})

	c.cfg.MaxEntries = 2
	c.cfg.ErrorTTL = time.Second

	if _, err := c.Get(context.Background(), "ok"); err != nil {
		t.Fatalf("Get(\"ok\"): %v", err)
	}

	if _, err := c.Get(context.Background(), "ko"); !errors.Is(err, sentinel) {
		t.Fatalf("Get(\"ko\"): %v", err)
	}

	c.mtx.Lock()
	okEntry := c.store["ok"]
	okEntry.FetchedAt = time.Now().Add(-30 * time.Minute)
	c.store["ok"] = okEntry

	koEntry := c.store["ko"]
	koEntry.FetchedAt = time.Now().Add(-2 * time.Second)
	c.store["ko"] = koEntry
	c.mtx.Unlock()

	if _, err := c.Get(context.Background(), "new"); err != nil {
		t.Fatalf("Get(\"new\"): %v", err)
	}

	c.mtx.Lock()
	size := len(c.store)
	_, hasOK := c.store["ok"]
	_, hasKO := c.store["ko"]
	_, hasNew := c.store["new"]
	c.mtx.Unlock()

	if size != 2 {
		t.Errorf("len(store) = %d, want 2 (MaxEntries)", size)
	}

	if hasKO {
		t.Error("expired error entry still present")
	}

	if !hasOK {
		t.Error("fresh success was evicted although expired entry was available")
	}

	if !hasNew {
		t.Error("new entry was not inserted")
	}
}
