// SPDX-License-Identifier: AGPL-3.0-or-later

package utils_test

import (
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
)

func TestTernary(t *testing.T) {
	t.Parallel()

	if got := utils.Ternary(true, "a", "b"); got != "a" {
		t.Errorf("utils.Ternary(true) = %q, want %q", got, "a")
	}

	if got := utils.Ternary(false, "a", "b"); got != "b" {
		t.Errorf("utils.Ternary(false) = %q, want %q", got, "b")
	}

	if got := utils.Ternary(true, 1, 2); got != 1 {
		t.Errorf("utils.Ternary(true, 1, 2) = %d, want 1", got)
	}
}
