// SPDX-License-Identifier: AGPL-3.0-or-later

package pgmodels

import (
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/lib/pq"
)

const (
	ComicLibraryEntries string = "LibraryEntries"
)

type Comic struct {
	CreatedAt time.Time   `gorm:"autoCreateTime"`
	UpdatedAt time.Time   `gorm:"autoUpdateTime"`
	Author    string      `gorm:"type:text"`
	Status    ComicStatus `gorm:"type:enum('ongoing','completed','hiatus','cancelled','dropped')"`
	//nolint:lll
	Source         sources.SourceName `gorm:"type:enum('asurascans');not null;uniqueIndex:idx_comic_source_slug,priority:1;"`
	Description    string             `gorm:"type:text"`
	Artist         string             `gorm:"type:text"`
	Slug           string             `gorm:"type:text;not null;uniqueIndex:idx_comic_source_slug,priority:2"`
	Title          string             `gorm:"type:text;not null"`
	ComicType      ComicType          `gorm:"column:comic_type;type:enum('manga','mangatoon','manhua','manwha')"`
	Genres         pq.StringArray     `gorm:"type:text[];not null;default:'{}'"`
	AltTitles      pq.StringArray     `gorm:"type:text[];not null;default:'{}'"`
	LibraryEntries []LibraryEntry     `gorm:"foreignKey:ComicID"`
	ChapterCount   int
	ID             uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
}

func (Comic) TableName() string {
	return "comics"
}

func (c *Comic) Domain() comics.Comic {
	return comics.Comic{
		ID:           c.ID,
		Source:       c.Source,
		Slug:         c.Slug,
		Title:        c.Title,
		Status:       c.Status.Domain(),
		Type:         c.ComicType.Domain(),
		Genres:       c.Genres,
		ChapterCount: c.ChapterCount,
		Author:       c.Author,
		Artist:       c.Artist,
		Description:  c.Description,
		AltTitles:    c.AltTitles,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
}

type ComicType string

const (
	ComicTypeManga     ComicType = "manga"
	ComicTypeMangatoon ComicType = "mangatoon"
	ComicTypeManhua    ComicType = "manhua"
	ComicTypeManhwa    ComicType = "manhwa"
)

func (t *ComicType) Domain() sources.SeriesType {
	return sources.SeriesType(*t)
}

func ComicTypeFromDomain(t sources.SeriesType) ComicType {
	return ComicType(t)
}

type ComicStatus string

const (
	ComicStatusOngoing   ComicStatus = "ongoing"
	ComicStatusCompleted ComicStatus = "completed"
	ComicStatusHiatus    ComicStatus = "hiatus"
	ComicStatusCancelled ComicStatus = "cancelled"
	ComicStatusDropped   ComicStatus = "dropped"
)

func (t *ComicStatus) Domain() sources.SeriesStatus {
	return sources.SeriesStatus(*t)
}

func ComicStatusFromDomain(t sources.SeriesStatus) ComicStatus {
	return ComicStatus(t)
}
