// SPDX-License-Identifier: AGPL-3.0-or-later

package feed

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

type LatestChapter struct {
	PublishedAt      time.Time
	EarlyAccessUntil time.Time
	Title            string
	Number           float64
	Download         int
	ID               uuid.UUID
	ComicID          uuid.UUID
}

type Item struct {
	Title          string
	Slug           string
	Source         sources.SourceName
	Status         sources.SeriesStatus
	Type           sources.SeriesType
	LatestChapters []LatestChapter
	ID             uuid.UUID
}

type Page struct {
	Items []Item
	Total int64
}

type GetOpts struct {
	Source *sources.SourceName
	Type   *sources.SeriesType
	UserID uuid.UUID
	Limit  int
	Offset int
}

type ListPageOpts struct {
	Now    time.Time
	Source *sources.SourceName
	Type   *sources.SeriesType
	UserID uuid.UUID
	Limit  int
	Offset int
}

type ListChaptersOpts struct {
	Now      time.Time
	ComicIDs []uuid.UUID
}

type FeedRepository interface {
	ListPage(context.Context, ListPageOpts) (Page, error)
	ListUnlockedChapters(context.Context, ListChaptersOpts) ([]LatestChapter, error)
}

type FeedService interface {
	Get(context.Context, GetOpts) (Page, error)
}

func IsUnlocked(earlyAccessUntil, now time.Time) bool {
	return !earlyAccessUntil.After(now)
}

func AvailabilityAt(publishedAt, earlyAccessUntil, now time.Time) time.Time {
	if !earlyAccessUntil.IsZero() && !earlyAccessUntil.After(now) {
		return earlyAccessUntil
	}

	return publishedAt
}
