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
	ListResumable(context.Context, time.Time) ([]Chapter, error)
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

type ChapterNeighbor struct {
	Title  string
	ID     uuid.UUID
	Number float64
}

type ChapterDetail struct {
	Previous *ChapterNeighbor
	Next     *ChapterNeighbor
	Chapter  Chapter
}

type PageStore interface {
	OpenPage(comicID uuid.UUID, chapterNumber float64, index int) (diskPath, contentType string, err error)
}

type ServePageOpts struct {
	UserID    uuid.UUID
	ChapterID uuid.UUID
	Index     int
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
	ListForLibrary(context.Context, ListForLibraryOpts) ([]Chapter, error)
	GetForLibrary(context.Context, GetForLibraryOpts) (*Chapter, error)
	GetDetailForLibrary(context.Context, GetForLibraryOpts) (*ChapterDetail, error)
	ServePage(context.Context, ServePageOpts) (diskPath, contentType string, err error)
}

type ComicLookup interface {
	Exists(context.Context, uuid.UUID) (bool, error)
}

type RetryDownloadOpts struct {
	UserID    uuid.UUID
	ChapterID uuid.UUID
}

type GetByIdsOpts struct {
	IDs    []uuid.UUID
	UserID uuid.UUID
}

type ListForLibraryOpts struct {
	UserID  uuid.UUID
	ComicID uuid.UUID
}

type GetForLibraryOpts struct {
	UserID    uuid.UUID
	ChapterID uuid.UUID
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
