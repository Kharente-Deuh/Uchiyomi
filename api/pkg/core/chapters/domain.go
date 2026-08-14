// SPDX-License-Identifier: AGPL-3.0-or-later

package chapters

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Chapter struct {
	PublishedAt       time.Time
	EarlyAccessUntil  time.Time
	SourceChapterSlug string
	Title             string
	Number            float64
	PagesNb           int
	Download          int
	ID                uuid.UUID
	ComicID           uuid.UUID
}

type ChaptersRepository interface {
	Create(context.Context, CreateOpts) (*Chapter, error)
	ListByComicID(context.Context, uuid.UUID) ([]Chapter, error)
}

type CreateOpts struct {
	PublishedAt       time.Time
	EarlyAccessUntil  time.Time
	SourceChapterSlug string
	Title             string
	Number            float64
	PagesNb           int
	ComicID           uuid.UUID
}
