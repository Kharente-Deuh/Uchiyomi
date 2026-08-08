// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

func Ternary[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}

	return falseVal
}
