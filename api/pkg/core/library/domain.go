// SPDX-License-Identifier: AGPL-3.0-or-later

package library

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Entry struct {
	AddedAt time.Time
	ID      uuid.UUID
	UserID  uuid.UUID
	ComicID uuid.UUID
}

type LibraryRepository interface {
	Create(context.Context, CreateOpts) (*Entry, error)
	Delete(context.Context, DeleteOpts) error
}

type CreateOpts struct {
	UserID  uuid.UUID
	ComicID uuid.UUID
}

type DeleteOpts struct {
	UserID  uuid.UUID
	ComicID uuid.UUID
}
