// SPDX-License-Identifier: AGPL-3.0-or-later

package utils_test

import (
	"testing"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
)

func TestOptionalTime(t *testing.T) {
	t.Parallel()

	if got := utils.OptionalTime(nil); got != nil {
		t.Errorf("OptionalTime(nil) = %v, want nil", got)
	}

	zero := time.Time{}
	if got := utils.OptionalTime(&zero); got != nil {
		t.Errorf("OptionalTime(zero) = %v, want nil", got)
	}

	now := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	got := utils.OptionalTime(&now)
	if got == nil || !got.Equal(now) {
		t.Errorf("OptionalTime(now) = %v, want %v", got, now)
	}

	if got != &now {
		t.Error("OptionalTime must return the same pointer for a non-zero time")
	}
}
