// SPDX-License-Identifier: AGPL-3.0-or-later

package fncache

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

type Config[P any, T any] struct {
	Fn            func(context.Context, P) (*T, error)
	Key           func(P) string
	Name          string
	TTL           time.Duration
	ErrorTTL      time.Duration
	FetchTimeout  time.Duration
	CleanInterval time.Duration
	MaxEntries    int
}

func (cfg *Config[P, T]) Validate() error {
	if cfg.Name == "" {
		return errors.New("name is required")
	}

	if cfg.Fn == nil {
		return errors.New("fn is required")
	}

	if cfg.Key == nil {
		return errors.New("key is required")
	}

	if cfg.TTL < time.Second {
		return errors.New("ttl must be at least 1 second")
	}

	if cfg.ErrorTTL < time.Second {
		return errors.New("errorTTL must be at least 1 second")
	}

	if cfg.FetchTimeout < time.Second {
		return errors.New("fetchTimeout must be at least 1 second")
	}

	if cfg.CleanInterval < time.Second {
		return errors.New("cleanInterval must be at least 1 second")
	}

	if cfg.MaxEntries <= 0 {
		return errors.New("maxEntries must be greater than 0")
	}

	return nil
}

type Deps struct {
	Logger *slog.Logger
}

func (deps *Deps) Validate() error {
	if deps.Logger == nil {
		return errors.New("logger is required")
	}

	return nil
}

type item[T any] struct {
	FetchedAt time.Time
	Error     error
	Result    *T
}

type Cache[P any, T any] struct {
	sf    singleflight.Group
	store map[string]item[T]
	deps  Deps
	cfg   Config[P, T]
	mtx   sync.Mutex
}

func New[P any, T any](cfg Config[P, T], deps Deps) (*Cache[P, T], error) {
	var err error
	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err = deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	deps.Logger = deps.Logger.With("component", fmt.Sprintf("fncache.%s", cfg.Name))

	c := &Cache[P, T]{
		cfg:   cfg,
		deps:  deps,
		store: make(map[string]item[T]),
	}

	return c, nil
}

func (c *Cache[P, T]) fresh(it item[T], now time.Time) bool {
	ttl := c.cfg.TTL
	if it.Error != nil {
		ttl = c.cfg.ErrorTTL
	}

	return now.Before(it.FetchedAt.Add(ttl))
}

func (c *Cache[P, T]) storeLocked(key string, it item[T]) {
	if _, exists := c.store[key]; !exists && len(c.store) >= c.cfg.MaxEntries {
		now := time.Now()

		var (
			staleKey, freshKey string
			staleAt, freshAt   time.Time
			hasStale, hasFresh bool
		)

		for k, v := range c.store {
			if !c.fresh(v, now) {
				if !hasStale || v.FetchedAt.Before(staleAt) {
					staleKey, staleAt, hasStale = k, v.FetchedAt, true
				}

				continue
			}

			if !hasFresh || v.FetchedAt.Before(freshAt) {
				freshKey, freshAt, hasFresh = k, v.FetchedAt, true
			}
		}

		switch {
		case hasStale:
			delete(c.store, staleKey)
		case hasFresh:
			delete(c.store, freshKey)
		}
	}

	c.store[key] = it
}

func cacheable(err error) bool {
	return !errors.Is(err, context.Canceled)
}

func (c *Cache[P, T]) Run(ctx context.Context) error {
	errG, errCtx := errgroup.WithContext(ctx)

	errG.Go(func() error {
		//nolint:wrapcheck
		return utils.Loop(errCtx, utils.LoopOpts{
			Interval: c.cfg.CleanInterval,
			Fn:       c.cleanStore,
		})
	})

	c.deps.Logger.Info("started")

	//nolint:wrapcheck
	return errG.Wait()
}

func (c *Cache[P, T]) Get(ctx context.Context, opts P) (*T, error) {
	return c.load(ctx, opts, false)
}

func (c *Cache[P, T]) Fetch(ctx context.Context, opts P) (*T, error) {
	return c.load(ctx, opts, true)
}

func (c *Cache[P, T]) load(ctx context.Context, opts P, fresh bool) (*T, error) {
	key := c.cfg.Key(opts)

	if !fresh {
		c.mtx.Lock()
		value, ok := c.store[key]
		c.mtx.Unlock()

		if ok && c.fresh(value, time.Now()) {
			return value.Result, value.Error
		}
	}

	ch := c.sf.DoChan(key, func() (any, error) {
		res, err := c.fetch(ctx, key, opts)

		return res, err
	})

	select {
	case <-ctx.Done():
		//nolint:wrapcheck
		return nil, ctx.Err()
	case res := <-ch:
		value, _ := res.Val.(*T)

		return value, res.Err
	}
}

func (c *Cache[P, T]) fetch(ctx context.Context, key string, opts P) (*T, error) {
	fetchCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), c.cfg.FetchTimeout)
	defer cancel()

	res, err := c.call(fetchCtx, opts)

	if cacheable(err) {
		c.mtx.Lock()
		c.storeLocked(key, item[T]{Error: err, Result: res, FetchedAt: time.Now()})
		c.mtx.Unlock()
	}

	return res, err
}

func (c *Cache[P, T]) call(ctx context.Context, opts P) (res *T, err error) {
	defer func() {
		if r := recover(); r != nil {
			res, err = nil, fmt.Errorf("panic dans Fn: %v", r)
		}
	}()

	return c.cfg.Fn(ctx, opts)
}

func (c *Cache[P, T]) cleanStore(_ context.Context) error {
	now := time.Now()

	c.mtx.Lock()
	for k, v := range c.store {
		if !c.fresh(v, now) {
			delete(c.store, k)
		}
	}
	c.mtx.Unlock()

	return nil
}
