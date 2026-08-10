// SPDX-License-Identifier: AGPL-3.0-or-later

package imgcache

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

const (
	testDir         = "/tmp/x"
	testImageName   = "001.jpg"
	testNestedImage = "serie/ch1/001.jpg"
	testEscapedPath = "etc/passwd"
)

func fetchString(payload string) func(context.Context, string) (io.ReadCloser, error) {
	return func(context.Context, string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(payload)), nil
	}
}

func startCache(t *testing.T, cfg Config) *Cache {
	t.Helper()

	ic, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)

	go func() { runErr <- ic.Run(ctx) }()

	t.Cleanup(func() {
		cancel()

		select {
		case <-runErr:
		case <-time.After(2 * time.Second):
			t.Error("Run did not return 2s after context cancellation")
		}
	})

	deadline := time.Now().Add(2 * time.Second)
	for !ic.IsReady() {
		if time.Now().After(deadline) {
			t.Fatal("cache never became ready")
		}

		time.Sleep(time.Millisecond)
	}

	return ic
}

func TestImageCacheConfigValidate(t *testing.T) {
	t.Parallel()

	fetch := fetchString("x")

	tests := map[string]struct {
		wantErr string
		cfg     Config
	}{
		"valide": {
			cfg: Config{Dir: testDir, ErrorCacheTTL: time.Minute, MinInterval: time.Millisecond, FetchFn: fetch},
		},
		"without dir": {
			cfg:     Config{ErrorCacheTTL: time.Minute, MinInterval: time.Millisecond, FetchFn: fetch},
			wantErr: "dir is required",
		},
		"TTL too short": {
			cfg:     Config{Dir: testDir, ErrorCacheTTL: time.Second, MinInterval: time.Millisecond, FetchFn: fetch},
			wantErr: "ErrorCacheTTL must be at least 1 minute",
		},
		"without interval": {
			cfg:     Config{Dir: testDir, ErrorCacheTTL: time.Minute, FetchFn: fetch},
			wantErr: "MinInterval is required",
		},
		"negative interval": {
			cfg:     Config{Dir: testDir, ErrorCacheTTL: time.Minute, MinInterval: -time.Second, FetchFn: fetch},
			wantErr: "MinInterval is required",
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

func TestImageCacheConfigValidateRequiresFetchFn(t *testing.T) {
	t.Parallel()

	cfg := Config{Dir: testDir, ErrorCacheTTL: time.Minute, MinInterval: time.Millisecond}

	err := cfg.Validate()
	if err == nil || err.Error() != "FetchFn is required" {
		t.Errorf("Validate() = %v, want %q", err, "FetchFn is required")
	}
}

func TestImageCacheUsesInjectedLogger(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	ic := startCache(t, Config{
		Dir:           t.TempDir(),
		ErrorCacheTTL: time.Minute,
		MinInterval:   time.Millisecond,
		Logger:        logger,
		FetchFn:       fetchString("x"),
	})

	if _, err := ic.Ensure(context.Background(), "a.jpg"); err != nil {
		t.Fatalf("Ensure: %v", err)
	}

	if !strings.Contains(buf.String(), `"component":"imgcache"`) {
		t.Errorf("injected logger was not used, logs: %s", buf.String())
	}
}

func TestImageCacheEnsureBeforeRun(t *testing.T) {
	t.Parallel()

	ic, err := New(Config{
		Dir:           t.TempDir(),
		ErrorCacheTTL: time.Minute,
		MinInterval:   time.Millisecond,
		FetchFn:       fetchString("data"),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if _, err := ic.Ensure(context.Background(), "a.jpg"); !errors.Is(err, ErrNotReady) {
		t.Errorf("Ensure avant Run = %v, want ErrNotReady", err)
	}

	if ic.IsReady() {
		t.Error("IsReady() = true avant Run")
	}
}

func TestImageCacheEnsureDownloadsAndCaches(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	dir := t.TempDir()
	ic := startCache(t, Config{
		Dir:           dir,
		ErrorCacheTTL: time.Minute,
		MinInterval:   time.Millisecond,
		FetchFn: func(context.Context, string) (io.ReadCloser, error) {
			calls.Add(1)

			return io.NopCloser(strings.NewReader("image-bytes")), nil
		},
	})

	path, err := ic.Ensure(context.Background(), "one-piece/ch1/001.jpg")
	if err != nil {
		t.Fatalf("Ensure: %v", err)
	}

	if want := filepath.Join(dir, "one-piece", "ch1", testImageName); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file announced by Ensure is unreadable: %v", err)
	}

	if string(content) != "image-bytes" {
		t.Errorf("contenu = %q, want %q", content, "image-bytes")
	}

	if _, err := ic.Ensure(context.Background(), "one-piece/ch1/001.jpg"); err != nil {
		t.Fatalf("Ensure (2nd call): %v", err)
	}

	if got := calls.Load(); got != 1 {
		t.Errorf("FetchFn called %d times for 2 Ensure on same image, want 1", got)
	}
}

func TestImageCacheEnsureNoTempFileLeftBehind(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	ic := startCache(t, Config{
		Dir:           dir,
		ErrorCacheTTL: time.Minute,
		MinInterval:   time.Millisecond,
		FetchFn:       fetchString("payload"),
	})

	if _, err := ic.Ensure(context.Background(), "a/b.jpg"); err != nil {
		t.Fatalf("Ensure: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(dir, "a"))
	if err != nil {
		t.Fatalf("os.ReadDir: %v", err)
	}

	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".dl-") {
			t.Errorf("temp file left on disk: %s", e.Name())
		}
	}
}

func TestImageCacheEnsureCoalescesConcurrentCalls(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	started := make(chan struct{})
	release := make(chan struct{})

	ic := startCache(t, Config{
		Dir:           t.TempDir(),
		ErrorCacheTTL: time.Minute,
		MinInterval:   time.Millisecond,
		FetchFn: func(context.Context, string) (io.ReadCloser, error) {
			if calls.Add(1) == 1 {
				close(started)
				<-release
			}

			return io.NopCloser(strings.NewReader("same")), nil
		},
	})

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		if _, err := ic.Ensure(context.Background(), "shared.jpg"); err != nil {
			t.Errorf("Ensure (1): %v", err)
		}
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("FetchFn never started")
	}

	for range 3 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if _, err := ic.Ensure(context.Background(), "shared.jpg"); err != nil {
				t.Errorf("Ensure (n): %v", err)
			}
		}()
	}

	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Errorf("FetchFn called %d times for 4 concurrent Ensure on same key, want 1", got)
	}
}

func TestImageCacheEnsureCachesErrors(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	sentinel := errors.New("404 not found")
	ic := startCache(t, Config{
		Dir:           t.TempDir(),
		ErrorCacheTTL: time.Minute,
		MinInterval:   time.Millisecond,
		FetchFn: func(context.Context, string) (io.ReadCloser, error) {
			calls.Add(1)

			return nil, sentinel
		},
	})

	_, err := ic.Ensure(context.Background(), "missing.jpg")
	if !errors.Is(err, sentinel) {
		t.Fatalf("Ensure = %v, want %v", err, sentinel)
	}

	_, err = ic.Ensure(context.Background(), "missing.jpg")
	if !errors.Is(err, sentinel) {
		t.Errorf("Ensure (2nd call) = %v, want %v", err, sentinel)
	}

	if got := calls.Load(); got != 1 {
		t.Errorf("FetchFn called %d times, want 1 (error cache must absorb 2nd call)", got)
	}
}

func TestImageCacheEnsureRejectsInvalidNames(t *testing.T) {
	t.Parallel()

	ic := startCache(t, Config{
		Dir:           t.TempDir(),
		ErrorCacheTTL: time.Minute,
		MinInterval:   time.Millisecond,
		FetchFn:       fetchString("x"),
	})

	for _, name := range []string{"", "/", ".", "..", "a\x00b"} {
		if _, err := ic.Ensure(context.Background(), name); err == nil {
			t.Errorf("Ensure(%q) = nil, want error", name)
		}
	}
}

func TestImageCacheEscapesAreContained(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	ic := startCache(t, Config{
		Dir:           dir,
		ErrorCacheTTL: time.Minute,
		MinInterval:   time.Millisecond,
		FetchFn:       fetchString("pwned"),
	})

	path, err := ic.Ensure(context.Background(), "../../../etc/passwd")
	if err != nil {
		t.Fatalf("Ensure: %v", err)
	}

	if !strings.HasPrefix(path, dir+string(os.PathSeparator)) {
		t.Errorf("write outside cache directory: %q (cache: %q)", path, dir)
	}
}

func TestImageCacheRunCreatesNestedDir(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "a", "b", "c")

	ic := startCache(t, Config{
		Dir:           dir,
		ErrorCacheTTL: time.Minute,
		MinInterval:   time.Millisecond,
		FetchFn:       fetchString("payload"),
	})

	path, err := ic.Ensure(context.Background(), "serie/001.jpg")
	if err != nil {
		t.Fatalf("Ensure in freshly created directory: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("os.Stat(%q): %v", path, err)
	}
}

func TestSafeKey(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		in      string
		want    string
		wantErr bool
	}{
		"simple path":             {in: testImageName, want: testImageName},
		"sous-dossiers":           {in: testNestedImage, want: testNestedImage},
		"slash initial":           {in: "/serie/001.jpg", want: "serie/001.jpg"},
		"simple traversal":        {in: "../001.jpg", want: testImageName},
		"multiple traversal":      {in: "../../../etc/passwd", want: testEscapedPath},
		"traversal in the middle": {in: "serie/../../etc/passwd", want: testEscapedPath},
		"antislashs windows":      {in: `serie\ch1\001.jpg`, want: testNestedImage},
		"backslash traversal":     {in: `..\..\etc\passwd`, want: testEscapedPath},
		"point courant":           {in: "./001.jpg", want: testImageName},
		"double slash":            {in: "serie//001.jpg", want: "serie/001.jpg"},
		"empty":                   {in: "", wantErr: true},
		"octet nul":               {in: "a\x00b.jpg", wantErr: true},
		"racine":                  {in: "/", wantErr: true},
		"point":                   {in: ".", wantErr: true},
		"double point":            {in: "..", wantErr: true},
		"traversal only":          {in: "../..", wantErr: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := safeKey(tc.in)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("safeKey(%q) = %q, want error", tc.in, got)
				}

				return
			}

			if err != nil {
				t.Fatalf("safeKey(%q): %v", tc.in, err)
			}

			if got != tc.want {
				t.Errorf("safeKey(%q) = %q, want %q", tc.in, got, tc.want)
			}

			if strings.HasPrefix(got, "/") || strings.Contains(got, "..") {
				t.Errorf("safeKey(%q) = %q: key not contained", tc.in, got)
			}
		})
	}
}

func TestImageCacheRateLimitsFetches(t *testing.T) {
	t.Parallel()

	const minInterval = 100 * time.Millisecond

	ic := startCache(t, Config{
		Dir:           t.TempDir(),
		ErrorCacheTTL: time.Minute,
		MinInterval:   minInterval,
		FetchFn:       fetchString("x"),
	})

	start := time.Now()

	for _, name := range []string{"a.jpg", "b.jpg", "c.jpg"} {
		if _, err := ic.Ensure(context.Background(), name); err != nil {
			t.Fatalf("Ensure(%q): %v", name, err)
		}
	}

	if elapsed := time.Since(start); elapsed < 2*minInterval {
		t.Errorf("3 downloads in %v, want at least %v", elapsed, 2*minInterval)
	}
}

func TestImageCacheProcessSkipsRateLimitWhenAlreadyOnDisk(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	dir := t.TempDir()

	ic, err := New(Config{
		Dir:           dir,
		ErrorCacheTTL: time.Minute,
		MinInterval:   time.Hour,
		FetchFn: func(context.Context, string) (io.ReadCloser, error) {
			calls.Add(1)

			return io.NopCloser(strings.NewReader("x")), nil
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	diskPath := filepath.Join(dir, testImageName)
	if err := os.WriteFile(diskPath, []byte("already there"), 0o600); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}

	j := &job{key: testImageName, diskPath: diskPath, done: make(chan struct{})}

	lim := rate.NewLimiter(rate.Every(time.Hour), 1)
	if !lim.Allow() {
		t.Fatal("initial rate limiter token should be available")
	}

	finished := make(chan struct{})

	go func() {
		defer close(finished)

		ic.process(context.Background(), j, lim)
	}()

	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("process waited on rate limiter for cached file")
	}

	if j.err != nil {
		t.Errorf("job.err = %v, want nil", j.err)
	}

	if got := calls.Load(); got != 0 {
		t.Errorf("FetchFn called %d times for cached file, want 0", got)
	}
}

func TestImageCacheProcessDoesNotCacheContextErrors(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		newCtx   func() (context.Context, context.CancelFunc)
		checkErr func(*testing.T, error)
	}{
		"annulation pendant l'attente": {
			newCtx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())

				go func() {
					time.Sleep(100 * time.Millisecond)
					cancel()
				}()

				return ctx, cancel
			},
			checkErr: func(t *testing.T, err error) {
				t.Helper()

				if !errors.Is(err, context.Canceled) {
					t.Errorf("job.err = %v, want context.Canceled", err)
				}
			},
		},
		"attente plus longue que la deadline": {
			newCtx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 500*time.Millisecond)
			},
			checkErr: func(t *testing.T, err error) {
				t.Helper()

				if err == nil {
					t.Fatal("job.err = nil, want rate limiter error")
				}

				if errors.Is(err, context.DeadlineExceeded) {
					t.Errorf("job.err = %v, want non-sentinel error", err)
				}
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			dir := t.TempDir()

			ic, err := New(Config{
				Dir:           dir,
				ErrorCacheTTL: time.Minute,
				MinInterval:   time.Hour,
				Logger:        slog.New(slog.NewJSONHandler(&buf, nil)),
				FetchFn:       fetchString("x"),
			})
			if err != nil {
				t.Fatalf("New: %v", err)
			}

			lim := rate.NewLimiter(rate.Every(time.Hour), 1)
			if !lim.Allow() {
				t.Fatal("initial rate limiter token should be available")
			}

			ctx, cancel := tc.newCtx()
			defer cancel()

			j := &job{
				key:      testImageName,
				diskPath: filepath.Join(dir, testImageName),
				done:     make(chan struct{}),
			}

			ic.process(ctx, j, lim)

			tc.checkErr(t, j.err)

			if !strings.Contains(buf.String(), "job failed") {
				t.Fatalf("process did not run through defer, logs: %s", buf.String())
			}

			ic.mtx.Lock()
			defer ic.mtx.Unlock()

			if len(ic.errorCache) != 0 {
				t.Errorf("error cache not empty (%d entry/entries): these errors say nothing about the image",
					len(ic.errorCache))
			}
		})
	}
}
