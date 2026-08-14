// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"context"
	"testing"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"golang.org/x/time/rate"
)

func TestSourceThrottleRespectsRateLimit(t *testing.T) {
	t.Parallel()

	const minInterval = 50 * time.Millisecond

	throttle := newSourceThrottle(minInterval)
	start := time.Now()

	for range 3 {
		if err := throttle.wait(context.Background()); err != nil {
			t.Fatalf("throttle.wait: %v", err)
		}
	}

	if elapsed := time.Since(start); elapsed < 2*minInterval {
		t.Fatalf("3 waits in %v, want at least %v", elapsed, 2*minInterval)
	}
}

func TestBuildSourceThrottlesUsesPerSourceOverrides(t *testing.T) {
	t.Parallel()

	sourcesMap := sources.SourceMap{
		"fast": nil,
		"slow": nil,
	}

	throttles := buildSourceThrottles(
		sourcesMap,
		500*time.Millisecond,
		map[sources.SourceName]time.Duration{
			"slow": time.Second,
		},
	)

	if got := throttles["fast"].limiter.Limit(); got != rate.Limit(2) {
		t.Fatalf("fast source limit = %v, want 2 req/s", got)
	}

	if got := throttles["slow"].limiter.Limit(); got != rate.Limit(1) {
		t.Fatalf("slow source limit = %v, want 1 req/s", got)
	}
}
