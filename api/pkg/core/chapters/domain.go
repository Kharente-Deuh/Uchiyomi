// SPDX-License-Identifier: AGPL-3.0-or-later

package chapters

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

type Chapter struct {
	PublishedAt       time.Time
	EarlyAccessUntil  *time.Time
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
	CreateMany(context.Context, []CreateOpts) ([]Chapter, error)
	ListByComicID(context.Context, uuid.UUID) ([]Chapter, error)
	ListResumable(context.Context) ([]Chapter, error)
	ListEarlyAccessUnlocked(context.Context, time.Time) ([]Chapter, error)
	GetByID(context.Context, uuid.UUID) (*Chapter, error)
	UpdateDownload(context.Context, uuid.UUID, int) error
	UpdatePagesNb(context.Context, uuid.UUID, int) error
	GetByIds(context.Context, []uuid.UUID) ([]Chapter, error)
}

type ChapterDownloader interface {
	Enqueue(context.Context, []Chapter) error
	CleanupComic(context.Context, uuid.UUID, []Chapter) error
	ResetAndEnqueue(context.Context, uuid.UUID) error
	Resume(context.Context, uuid.UUID) error
}

type ChaptersService interface {
	CreateAll(context.Context, uuid.UUID, []sources.SourceChapter) ([]Chapter, error)
	ListByComicID(context.Context, uuid.UUID) ([]Chapter, error)
	EnqueueDownloadable(context.Context, []Chapter) error
	EnqueueResumable(context.Context) error
	ScanEarlyAccess(context.Context) error
	CleanupComic(context.Context, uuid.UUID, []Chapter) error
	RetryDownload(context.Context, RetryDownloadOpts) error
	GetByIds(context.Context, GetByIdsOpts) ([]Chapter, error)
}

type RetryDownloadOpts struct {
	UserID    uuid.UUID
	ChapterID uuid.UUID
}

type GetByIdsOpts struct {
	IDs    []uuid.UUID
	UserID uuid.UUID
}

type CreateOpts struct {
	PublishedAt       time.Time
	EarlyAccessUntil  *time.Time
	SourceChapterSlug string
	Title             string
	Number            float64
	PagesNb           int
	ComicID           uuid.UUID
}
