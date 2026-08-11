// SPDX-License-Identifier: AGPL-3.0-or-later

package library

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
)

type Entry struct {
	AddedAt time.Time
	ID      uuid.UUID
	UserID  uuid.UUID
	ComicID uuid.UUID
}

type EntryWithComic struct {
	Entry Entry
	Comic comics.Comic
}

type LibraryRepository interface {
	GetByID(context.Context, uuid.UUID) (*Entry, error)
	GetByUserAndComic(context.Context, uuid.UUID, uuid.UUID) (*Entry, error)
	ListByUser(context.Context, uuid.UUID) ([]EntryWithComic, error)
	Create(context.Context, CreateEntryOpts) (*Entry, error)
	Delete(context.Context, uuid.UUID) error
}

type CreateEntryOpts struct {
	AddedAt time.Time
	UserID  uuid.UUID
	ComicID uuid.UUID
}
