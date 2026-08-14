// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"context"
	"fmt"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"golang.org/x/time/rate"
)

type sourceThrottle struct {
	limiter *rate.Limiter
}

func newSourceThrottle(interval time.Duration) *sourceThrottle {
	return &sourceThrottle{
		limiter: rate.NewLimiter(rate.Every(interval), 1),
	}
}

func (t *sourceThrottle) wait(ctx context.Context) error {
	if err := t.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("t.limiter.Wait: %w", err)
	}

	return nil
}

func buildSourceThrottles(
	sourceMap sources.SourceMap,
	defaultInterval time.Duration,
	overrides map[sources.SourceName]time.Duration,
) map[sources.SourceName]*sourceThrottle {
	throttles := make(map[sources.SourceName]*sourceThrottle, len(sourceMap))

	for name := range sourceMap {
		interval := defaultInterval
		if override, ok := overrides[name]; ok && override > 0 {
			interval = override
		}

		throttles[name] = newSourceThrottle(interval)
	}

	return throttles
}
