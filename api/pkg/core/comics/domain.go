// SPDX-License-Identifier: AGPL-3.0-or-later

package comics

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

type Comic struct {
	UpdatedAt    time.Time
	CreatedAt    time.Time
	Source       sources.SourceName
	Artist       string
	Type         sources.SeriesType
	Description  string
	Author       string
	Status       sources.SeriesStatus
	Slug         string
	Title        string
	Genres       []string
	AltTitles    []string
	ChapterCount int
	ID           uuid.UUID
}

type GetBySourceSlugOpts struct {
	Source sources.SourceName
	Slug   string
	UserID uuid.UUID
}

type FindBySourceSlugOpts struct {
	Source sources.SourceName
	Slug   string
}

type LocalCoverStore interface {
	ObtainLocal(ctx context.Context, comicID uuid.UUID, source, slug string) error
	ServeLocal(ctx context.Context, comicID uuid.UUID) (diskPath, contentType string, err error)
	RemoveLocal(comicID uuid.UUID) error
}

type ComicsRepository interface {
	GetByID(context.Context, GetByIDOpts) (*Comic, error)
	FindByID(context.Context, uuid.UUID) (*Comic, error)
	GetBySourceSlug(context.Context, GetBySourceSlugOpts) (*Comic, error)
	FindBySourceSlug(context.Context, FindBySourceSlugOpts) (*Comic, error)
	Create(context.Context, CreateComicOpts) (*Comic, error)
	GetBySlugsAndSource(context.Context, GetBySlugsAndSource) ([]Comic, error)
	Delete(context.Context, uuid.UUID) error
	GetMany(context.Context, GetManyOpts) (Page, error)
	ListByStatuses(context.Context, ListByStatusesOpts) ([]Comic, error)
	UpdateStatusAndChapterCount(context.Context, UpdateStatusAndChapterCountOpts) error
	UpdateType(context.Context, UpdateTypeOpts) error
}

type CreateComicOpts struct {
	Status       sources.SeriesStatus
	Type         sources.SeriesType
	Description  string
	Source       sources.SourceName
	Artist       string
	Slug         string
	Author       string
	Title        string
	AltTitles    []string
	Genres       []string
	ChapterCount int
	ID           uuid.UUID
}

type GetBySlugsAndSource struct {
	Source sources.SourceName
	Slugs  []string
	UserID uuid.UUID
}

var ErrSourceUnavailable = errors.New("source unavailable")

type RefreshComicOpts struct {
	UserID uuid.UUID
	ID     uuid.UUID
}

type RetryChaptersOpts struct {
	ChapterIDs []uuid.UUID
	UserID     uuid.UUID
	ComicID    uuid.UUID
}

type ComicsService interface {
	Create(context.Context, CreateOpts) (*Comic, error)
	GetByID(context.Context, GetByIDOpts) (*Comic, error)
	GetMany(context.Context, GetManyOpts) (Page, error)
	Delete(context.Context, DeleteOpts) error
	RefreshChapterLists(context.Context) error
	RefreshComic(context.Context, RefreshComicOpts) (*Comic, error)
	RetryChapters(context.Context, RetryChaptersOpts) error
	ServeCover(ctx context.Context, opts GetByIDOpts) (diskPath, contentType string, err error)
	UpdateType(context.Context, UpdateTypeOpts) (*Comic, error)
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

type Page struct {
	Items []Comic
	Total int64
}

type ListSort string

const (
	ListSortTitle   ListSort = "title"
	ListSortAddedAt ListSort = "addedAt"
)

func ParseListSort(s string) (ListSort, error) {
	switch s {
	case string(ListSortTitle):
		return ListSortTitle, nil
	case string(ListSortAddedAt):
		return ListSortAddedAt, nil
	default:
		return "", fmt.Errorf("invalid sort: %s", s)
	}
}

type ListOrder string

const (
	ListOrderAsc  ListOrder = "asc"
	ListOrderDesc ListOrder = "desc"
)

func ParseListOrder(s string) (ListOrder, error) {
	switch s {
	case string(ListOrderAsc):
		return ListOrderAsc, nil
	case string(ListOrderDesc):
		return ListOrderDesc, nil
	default:
		return "", fmt.Errorf("invalid order: %s", s)
	}
}

type GetManyOpts struct {
	UserID *uuid.UUID
	Source *sources.SourceName
	Type   *sources.SeriesType
	Status *sources.SeriesStatus
	Search string
	Sort   ListSort
	Order  ListOrder
	Limit  int
	Offset int
}

type DeleteOpts struct {
	UserID uuid.UUID
	ID     uuid.UUID
}

type ListByStatusesOpts struct {
	Source   sources.SourceName
	Statuses []sources.SeriesStatus
}

type UpdateStatusAndChapterCountOpts struct {
	Status       sources.SeriesStatus
	ID           uuid.UUID
	ChapterCount int
}

type UpdateTypeOpts struct {
	Type   sources.SeriesType
	UserID uuid.UUID
	ID     uuid.UUID
}
