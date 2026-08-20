// SPDX-License-Identifier: AGPL-3.0-or-later

package pgmodels

import (
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readingprogress"
)

type ReadingProgress struct {
	UpdatedAt      time.Time    `gorm:"not null"`
	LibraryEntry   LibraryEntry `gorm:"foreignKey:LibraryEntryID;constraint:OnDelete:CASCADE"`
	Chapter        Chapter      `gorm:"foreignKey:ChapterID;constraint:OnDelete:CASCADE"`
	ID             uuid.UUID    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	LibraryEntryID uuid.UUID    `gorm:"type:uuid;not null;uniqueIndex:idx_reading_progress_entry_chapter,priority:1"`
	ChapterID      uuid.UUID    `gorm:"type:uuid;not null;uniqueIndex:idx_reading_progress_entry_chapter,priority:2"`
	Page           int          `gorm:"not null"`
}

func (ReadingProgress) TableName() string {
	return "reading_progress"
}

func (m *ReadingProgress) Domain() readingprogress.Progress {
	return readingprogress.Progress{
		ChapterID: m.ChapterID,
		Page:      m.Page,
		UpdatedAt: m.UpdatedAt,
	}
}
