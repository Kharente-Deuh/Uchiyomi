// SPDX-License-Identifier: AGPL-3.0-or-later

package comics

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

type Comic struct {
	ID           uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Artist       string
	Type         sources.SeriesType
	Description  string
	Source       sources.SourceName
	Author       string
	Status       sources.SeriesStatus
	Slug         string
	Title        string
	Genres       []string
	AltTitles    []string
	ChapterCount int
}

type GetBySourceSlugOpts struct {
	UserID uuid.UUID
	Source sources.SourceName
	Slug   string
}

type ComicsRepository interface {
	GetByID(context.Context, GetByIDOpts) (*Comic, error)
	GetBySourceSlug(context.Context, GetBySourceSlugOpts) (*Comic, error)
	Create(context.Context, CreateComicOpts) (*Comic, error)
	GetBySlugsAndSource(context.Context, sources.SourceName, []string) ([]Comic, error)
	Delete(context.Context, uuid.UUID) error
	GetMany(context.Context, GetManyOpts) ([]Comic, error)
}

type CreateComicOpts struct {
	Status       sources.SeriesStatus
	Type         sources.SeriesType
	Description  string
	CoverPath    string
	Source       sources.SourceName
	Artist       string
	Slug         string
	Author       string
	Title        string
	AltTitles    []string
	Genres       []string
	ChapterCount int
}

type ComicsService interface {
	Create(context.Context, CreateOpts) (*Comic, error)
	GetByID(context.Context, GetByIDOpts) (*Comic, error)
	GetMany(context.Context, GetManyOpts) ([]Comic, error)
	Delete(context.Context, DeleteOpts) error
}

type CreateOpts struct {
	Source sources.SourceName
	Slug   string
	UserID uuid.UUID
}

type GetByIDOpts struct {
	UserID uuid.UUID
	ID     uuid.UUID
}

type GetManyOpts struct {
	UserID *uuid.UUID
	Source *sources.SourceName
	Type   *sources.SeriesType
	Status *sources.SeriesStatus
	Limit  int
	Offset int
}

type DeleteOpts struct {
	UserID uuid.UUID
	ID     uuid.UUID
}
