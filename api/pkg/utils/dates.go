package utils

import "time"

func OptionalTime(t *time.Time) *time.Time {
	if t != nil && !t.IsZero() {
		return t
	}

	return nil
}
