// SPDX-License-Identifier: AGPL-3.0-or-later

package comics

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Comic struct {
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Artist       string
	Type         ComicType
	Description  string
	CoverPath    string
	Source       string
	Author       string
	Status       ComicStatus
	Slug         string
	Title        string
	Genres       []string
	AltTitles    []string
	ChapterCount int
	ID           uuid.UUID
}

type ComicType string

const (
	ComicTypeManga     ComicType = "manga"
	ComicTypeMangatoon ComicType = "mangatoon"
	ComicTypeManhua    ComicType = "manhua"
	ComicTypeManhwa    ComicType = "manhwa"
)

type ComicStatus string

const (
	ComicStatusOngoing   ComicStatus = "ongoing"
	ComicStatusCompleted ComicStatus = "completed"
	ComicStatusHiatus    ComicStatus = "hiatus"
	ComicStatusCancelled ComicStatus = "cancelled"
	ComicStatusDropped   ComicStatus = "dropped"
)

type SourceSlugKey struct {
	Source string
	Slug   string
}

type ComicsRepository interface {
	GetByID(context.Context, uuid.UUID) (*Comic, error)
	GetBySourceSlug(context.Context, SourceSlugKey) (*Comic, error)
	Create(context.Context, CreateComicOpts) (*Comic, error)
	GetBySlugsAndSource(context.Context, string, []string) ([]Comic, error)
	Delete(context.Context, uuid.UUID) error
}

type CreateComicOpts struct {
	Status       ComicStatus
	Type         ComicType
	Description  string
	CoverPath    string
	Source       string
	Artist       string
	Slug         string
	Author       string
	Title        string
	AltTitles    []string
	Genres       []string
	ChapterCount int
}
