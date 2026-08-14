// SPDX-License-Identifier: AGPL-3.0-or-later

package pgmodels

import (
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
)

type Chapter struct {
	PublishedAt      time.Time `gorm:"not null"`
	EarlyAccessUntil time.Time `gorm:"not null"`
	//nolint:lll
	SourceChapterSlug string    `gorm:"column:source_chapter_slug;type:text;not null;uniqueIndex:idx_chapter_comic_source_slug,priority:2"`
	Title             string    `gorm:"type:text"`
	Comic             Comic     `gorm:"foreignKey:ComicID;constraint:OnDelete:CASCADE"`
	Number            float64   `gorm:"not null"`
	PagesNb           int       `gorm:"column:pages_nb;not null;default:0"`
	Download          int       `gorm:"not null;default:0"`
	ID                uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ComicID           uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_chapter_comic_source_slug,priority:1;index"`
}

func (Chapter) TableName() string {
	return "chapters"
}

func (c *Chapter) Domain() chapters.Chapter {
	return chapters.Chapter{
		ID:                c.ID,
		ComicID:           c.ComicID,
		SourceChapterSlug: c.SourceChapterSlug,
		Number:            c.Number,
		Title:             c.Title,
		PagesNb:           c.PagesNb,
		PublishedAt:       c.PublishedAt,
		EarlyAccessUntil:  c.EarlyAccessUntil,
		Download:          c.Download,
	}
}
