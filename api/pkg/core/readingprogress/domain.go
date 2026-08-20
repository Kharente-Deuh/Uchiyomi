// SPDX-License-Identifier: AGPL-3.0-or-later

package readingprogress

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
)

var ErrInvalid = errors.New("invalid reading progress")

type Progress struct {
	UpdatedAt time.Time
	ChapterID uuid.UUID
	Page      int
}

type Continue struct {
	ChapterID uuid.UUID
	Page      int
}

type ListResult struct {
	Continue *Continue
	Chapters []Progress
}

type ListOpts struct {
	UserID  uuid.UUID
	ComicID uuid.UUID
}

type SaveOpts struct {
	UserID    uuid.UUID
	ComicID   uuid.UUID
	ChapterID uuid.UUID
	Page      int
}

type GetOpts struct {
	UserID    uuid.UUID
	ComicID   uuid.UUID
	ChapterID uuid.UUID
}

type UpsertOpts struct {
	UpdatedAt time.Time
	UserID    uuid.UUID
	ComicID   uuid.UUID
	ChapterID uuid.UUID
	Page      int
}

type Repository interface {
	ListByUserAndComic(context.Context, ListOpts) ([]Progress, error)
	Get(context.Context, GetOpts) (*Progress, error)
	Upsert(context.Context, UpsertOpts) (Progress, error)
}

type ReadingProgressService interface {
	List(context.Context, ListOpts) (ListResult, error)
	Save(context.Context, SaveOpts) (Progress, error)
}

type LibraryMembership interface {
	ExistsByUserAndComic(context.Context, uuid.UUID, uuid.UUID) (bool, error)
}

type ComicLookup interface {
	Exists(context.Context, uuid.UUID) (bool, error)
}

type ChapterLookup interface {
	GetByID(context.Context, uuid.UUID) (*chapters.Chapter, error)
	GetByIds(context.Context, []uuid.UUID) ([]chapters.Chapter, error)
}

func ClampPage(page, pagesNb int) int {
	if pagesNb > 0 && page > pagesNb {
		return pagesNb
	}

	return page
}

func MergePage(stored *int, incoming, pagesNb int) (int, error) {
	if incoming < 1 {
		return 0, fmt.Errorf("%w: page must be at least 1", ErrInvalid)
	}

	if pagesNb > 0 && incoming > pagesNb {
		return 0, fmt.Errorf("%w: page exceeds chapter length", ErrInvalid)
	}

	page := incoming
	if stored != nil && *stored > page {
		page = *stored
	}

	return ClampPage(page, pagesNb), nil
}
