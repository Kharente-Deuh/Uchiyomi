// SPDX-License-Identifier: AGPL-3.0-or-later

package comics

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Comic struct {
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Artist           string
	Type             string
	Description      string
	LocalCoverPath   string
	ExternalCoverURL string
	SourceURL        string
	Source           string
	Author           string
	Status           string
	Slug             string
	Title            string
	Genres           []string
	AltTitles        []string
	ChapterCount     int
	ReleaseYear      int
	Rating           float64
	ID               uuid.UUID
}

type SourceSlugKey struct {
	Source string
	Slug   string
}

type ComicsRepository interface {
	GetByID(context.Context, uuid.UUID) (*Comic, error)
	GetBySourceSlug(context.Context, SourceSlugKey) (*Comic, error)
	Create(context.Context, CreateComicOpts) (*Comic, error)
}

type CreateComicOpts struct {
	Status           string
	Type             string
	Description      string
	LocalCoverPath   string
	ExternalCoverURL string
	SourceURL        string
	Source           string
	Artist           string
	Slug             string
	Author           string
	Title            string
	AltTitles        []string
	Genres           []string
	ChapterCount     int
	ReleaseYear      int
	Rating           float64
}
