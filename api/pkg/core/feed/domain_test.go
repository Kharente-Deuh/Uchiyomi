// SPDX-License-Identifier: AGPL-3.0-or-later

package feed_test

import (
	"testing"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/feed"
)

func TestIsUnlocked(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 16, 12, 0, 0, 0, time.UTC)
	aug20 := time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC)

	tests := map[string]struct {
		until time.Time
		now   time.Time
		want  bool
	}{
		"zero is unlocked":          {until: time.Time{}, now: now, want: true},
		"future early access locked": {until: aug20, now: now, want: false},
		"equal now is unlocked":     {until: now, now: now, want: true},
		"past early access unlocked": {until: aug20, now: time.Date(2026, time.August, 21, 0, 0, 0, 0, time.UTC), want: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := feed.IsUnlocked(tc.until, tc.now); got != tc.want {
				t.Errorf("IsUnlocked() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAvailabilityAtAugustScenario(t *testing.T) {
	t.Parallel()

	published10 := time.Date(2026, time.August, 10, 0, 0, 0, 0, time.UTC)
	early20 := time.Date(2026, time.August, 20, 0, 0, 0, 0, time.UTC)
	published1 := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)

	aug16 := time.Date(2026, time.August, 16, 12, 0, 0, 0, time.UTC)
	aug21 := time.Date(2026, time.August, 21, 0, 0, 0, 0, time.UTC)

	if got := feed.AvailabilityAt(published1, time.Time{}, aug16); !got.Equal(published1) {
		t.Errorf("ch10 on 16 Aug = %v, want publishedAt", got)
	}

	if got := feed.AvailabilityAt(published10, early20, aug16); !got.Equal(published10) {
		t.Errorf("ch11 still locked on 16 Aug availability = %v, want publishedAt (caller must not use this for ranking locked chapters)", got)
	}

	if got := feed.AvailabilityAt(published10, early20, aug21); !got.Equal(early20) {
		t.Errorf("ch11 on 21 Aug = %v, want earlyAccessUntil 20 Aug", got)
	}
}
