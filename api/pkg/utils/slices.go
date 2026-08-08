// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

func MapSlice[T, U any](s []T, f func(T) U) []U {
	us := make([]U, len(s))
	for i := range s {
		us[i] = f(s[i])
	}

	return us
}

func FilterSlice[T any](s []T, f func(T) bool) []T {
	filtered := []T{}

	for _, i := range s {
		if f(i) {
			filtered = append(filtered, i)
		}
	}

	return filtered
}
