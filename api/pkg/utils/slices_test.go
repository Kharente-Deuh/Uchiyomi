// SPDX-License-Identifier: AGPL-3.0-or-later

package utils_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
)

func TestMapSlice(t *testing.T) {
	t.Parallel()

	got := utils.MapSlice([]int{1, 2, 3}, func(i int) string { return strings.Repeat("x", i) })
	want := []string{"x", "xx", "xxx"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("utils.MapSlice() = %v, want %v", got, want)
	}
}

func TestMapSliceChangesType(t *testing.T) {
	t.Parallel()

	type user struct{ name string }

	got := utils.MapSlice([]user{{name: "bob"}, {name: "alice"}}, func(u user) string { return u.name })
	if !reflect.DeepEqual(got, []string{"bob", "alice"}) {
		t.Errorf("utils.MapSlice() = %v", got)
	}
}

func TestMapSliceEmpty(t *testing.T) {
	t.Parallel()

	if got := utils.MapSlice([]int{}, func(i int) int { return i }); len(got) != 0 {
		t.Errorf("utils.MapSlice(empty) = %v, want empty", got)
	}

	got := utils.MapSlice(nil, func(i int) int { return i })
	if got == nil {
		t.Error("utils.MapSlice(nil) = nil, want empty slice")
	}

	if len(got) != 0 {
		t.Errorf("utils.MapSlice(nil) = %v, want empty", got)
	}
}

func TestMapSlicePreservesOrder(t *testing.T) {
	t.Parallel()

	got := utils.MapSlice([]int{5, 1, 4, 2}, func(i int) int { return i * 10 })
	if !reflect.DeepEqual(got, []int{50, 10, 40, 20}) {
		t.Errorf("utils.MapSlice() = %v, order not preserved", got)
	}
}

func TestFilterSlice(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		in   []int
		keep func(int) bool
		want []int
	}{
		"keeps evens": {in: []int{1, 2, 3, 4}, keep: func(i int) bool { return i%2 == 0 }, want: []int{2, 4}},
		"tout garder": {in: []int{1, 2}, keep: func(int) bool { return true }, want: []int{1, 2}},
		"tout jeter":  {in: []int{1, 2}, keep: func(int) bool { return false }, want: []int{}},
		"empty slice": {in: []int{}, keep: func(int) bool { return true }, want: []int{}},
		"slice nil":   {in: nil, keep: func(int) bool { return true }, want: []int{}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := utils.FilterSlice(tc.in, tc.keep)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("utils.FilterSlice() = %v, want %v", got, tc.want)
			}

			if got == nil {
				t.Error("FilterSlice returned nil, want empty slice")
			}
		})
	}
}

func TestFilterSliceDoesNotMutateInput(t *testing.T) {
	t.Parallel()

	in := []int{1, 2, 3, 4}
	_ = utils.FilterSlice(in, func(i int) bool { return i%2 == 0 })

	if !reflect.DeepEqual(in, []int{1, 2, 3, 4}) {
		t.Errorf("input slice was modified: %v", in)
	}
}
