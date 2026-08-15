// SPDX-License-Identifier: AGPL-3.0-or-later

package domain

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrForbidden     = errors.New("forbidden")
	ErrConflict      = errors.New("conflict")
)
