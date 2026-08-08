// SPDX-License-Identifier: AGPL-3.0-or-later

package imgcache

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

var (
	ErrClosed   = errors.New("imgcache: closed")
	ErrNotReady = errors.New("imgcache: cache not ready")
)

type job struct {
	err      error
	done     chan struct{}
	key      string
	diskPath string
}

type jobError struct {
	err     error
	errorAt time.Time
}

type Config struct {
	FetchFn       func(context.Context, string) (io.ReadCloser, error)
	Logger        *slog.Logger
	Dir           string
	ErrorCacheTTL time.Duration
	MinInterval   time.Duration
}

func (cfg *Config) Validate() error {
	if cfg.Dir == "" {
		return errors.New("dir is required")
	}

	if cfg.FetchFn == nil {
		return errors.New("FetchFn is required")
	}

	if cfg.ErrorCacheTTL < time.Minute {
		return errors.New("ErrorCacheTTL must be at least 1 minute")
	}

	if cfg.MinInterval <= 0 {
		return errors.New("MinInterval is required")
	}

	return nil
}

type Cache struct {
	queue      chan *job
	stopped    chan struct{}
	errorCache map[string]*jobError
	processing map[string]*job
	logger     *slog.Logger
	cfg        Config
	mtx        sync.Mutex
	isReady    atomic.Bool
}

func New(cfg Config) (*Cache, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	ic := &Cache{
		cfg:        cfg,
		logger:     logger.With("component", "imgcache"),
		errorCache: make(map[string]*jobError),
		queue:      make(chan *job),
		processing: make(map[string]*job),
		stopped:    make(chan struct{}),
	}

	return ic, nil
}

func (ic *Cache) Run(ctx context.Context) error {
	err := utils.EnsureDir(ic.cfg.Dir)
	if err != nil {
		return fmt.Errorf("utils.EnsureDir: %w", err)
	}

	lim := rate.NewLimiter(rate.Every(ic.cfg.MinInterval), 1)

	errG, errCtx := errgroup.WithContext(ctx)

	errG.Go(func() error {
		for {
			select {
			case j := <-ic.queue:
				ic.process(errCtx, j, lim)

				continue
			default:
			}

			select {
			case j := <-ic.queue:
				ic.process(errCtx, j, lim)
			case <-errCtx.Done():
				return nil
			}
		}
	})

	errG.Go(func() error {
		//nolint:wrapcheck
		return utils.Loop(errCtx, utils.LoopOpts{
			Interval: ic.cfg.ErrorCacheTTL,
			Fn:       ic.clearErrorMap,
		})
	})

	ic.isReady.Store(true)
	defer close(ic.stopped)
	defer ic.isReady.Store(false)

	//nolint:wrapcheck
	return errG.Wait()
}

func (ic *Cache) IsReady() bool {
	return ic.isReady.Load()
}

func (ic *Cache) Ensure(ctx context.Context, name string) (string, error) {
	if !ic.isReady.Load() {
		return "", ErrNotReady
	}

	key, err := safeKey(name)
	if err != nil {
		return "", err
	}

	ic.logger.DebugContext(ctx, "ensure file", "key", key)

	diskPath := filepath.Join(ic.cfg.Dir, filepath.FromSlash(key))
	if st, err := os.Stat(diskPath); err == nil && st.Mode().IsRegular() {
		ic.logger.DebugContext(ctx, "already in cache", "key", key)

		return diskPath, nil
	}

	ic.mtx.Lock()
	jobErr, ok := ic.errorCache[key]
	if ok {
		if jobErr.errorAt.Add(ic.cfg.ErrorCacheTTL).After(time.Now()) {
			defer ic.mtx.Unlock()

			ic.logger.DebugContext(ctx, "served from error cache", "key", key, "err", jobErr.err)

			return "", jobErr.err
		} else {
			delete(ic.errorCache, key)
		}
	}

	ic.mtx.Unlock()

	job, err := ic.enqueue(ctx, key, diskPath)
	if err != nil {
		return "", fmt.Errorf("ic.enqueue: %w", err)
	}

	select {
	case <-job.done:
		if job.err != nil {
			return "", job.err
		}

		return diskPath, nil

	case <-ctx.Done():
		//nolint:wrapcheck
		return "", ctx.Err()
	}
}

func (ic *Cache) enqueue(ctx context.Context, key string, diskPath string) (*job, error) {
	ic.mtx.Lock()
	if existing, ok := ic.processing[key]; ok {
		ic.mtx.Unlock()

		return existing, nil
	}

	j := &job{
		key:      key,
		diskPath: diskPath,
		done:     make(chan struct{}),
	}

	ic.processing[key] = j

	ic.mtx.Unlock()

	select {
	case ic.queue <- j:
		return j, nil
	case <-ic.stopped:
		ic.endJob(j, ErrClosed)

		return nil, ErrClosed
	case <-ctx.Done():
		ic.endJob(j, ctx.Err())

		//nolint:wrapcheck
		return nil, ctx.Err()
	}
}

func (ic *Cache) process(ctx context.Context, j *job, lim *rate.Limiter) {
	ic.logger.DebugContext(ctx, "processing job", "key", j.key)

	var err error
	if err = ctx.Err(); err != nil {
		ic.endJob(j, err)

		return
	}

	fetchAttempted := false

	defer func() {
		if err != nil {
			ic.logger.WarnContext(ctx, "job failed", "key", j.key, "err", err)

			if fetchAttempted && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
				ic.mtx.Lock()
				ic.errorCache[j.key] = &jobError{err: err, errorAt: time.Now()}
				ic.mtx.Unlock()
			}
		}

		ic.endJob(j, err)
	}()

	if st, err := os.Stat(j.diskPath); err == nil && st.Mode().IsRegular() {
		ic.logger.DebugContext(ctx, "already in cache", "key", j.key)

		return
	}

	if err = lim.Wait(ctx); err != nil {
		err = fmt.Errorf("lim.Wait: %w", err)

		return
	}

	fetchAttempted = true

	rc, err := ic.cfg.FetchFn(ctx, j.key)
	if err != nil {
		err = fmt.Errorf("failed to fetch image: %w", err)

		return
	}

	defer rc.Close()
	if err = os.MkdirAll(filepath.Dir(j.diskPath), 0o755); err != nil {
		err = fmt.Errorf("os.MkdirAll %s: %w", j.diskPath, err)

		return
	}

	tmp, err := os.CreateTemp(filepath.Dir(j.diskPath), ".dl-*")
	if err != nil {
		err = fmt.Errorf("os.CreateTemp %s: %w", j.diskPath, err)

		return
	}

	tmpName := tmp.Name()
	defer func() {
		if err == nil {
			return
		}

		tmp.Close()
		os.Remove(tmpName)
	}()

	if _, err = io.Copy(tmp, rc); err != nil {
		err = fmt.Errorf("io.Copy %s: %w", tmpName, err)

		return
	}

	if err = tmp.Sync(); err != nil {
		err = fmt.Errorf("tmp.Sync %s: %w", tmpName, err)

		return
	}

	if err = tmp.Close(); err != nil {
		err = fmt.Errorf("tmp.Close %s: %w", tmpName, err)

		return
	}

	if err = os.Rename(tmpName, j.diskPath); err != nil {
		err = fmt.Errorf("os.Rename %s: %w", tmpName, err)
	}
}

func (ic *Cache) endJob(j *job, err error) {
	ic.mtx.Lock()
	if ic.processing[j.key] == j {
		delete(ic.processing, j.key)
	}

	ic.mtx.Unlock()

	j.err = err
	close(j.done)
}

func (ic *Cache) clearErrorMap(_ context.Context) error {
	limit := time.Now().Add(-ic.cfg.ErrorCacheTTL)

	ic.mtx.Lock()
	for k, v := range ic.errorCache {
		if v.errorAt.Before(limit) {
			delete(ic.errorCache, k)
		}
	}

	ic.mtx.Unlock()

	return nil
}

func safeKey(name string) (string, error) {
	if name == "" {
		return "", errors.New("file name is empty")
	}

	if strings.ContainsRune(name, 0) {
		return "", errors.New("invalid file name")
	}

	key := strings.TrimPrefix(path.Clean("/"+strings.ReplaceAll(name, `\`, "/")), "/")
	if key == "" || key == "." {
		return "", fmt.Errorf("invalid file name: %q", name)
	}

	return key, nil
}
