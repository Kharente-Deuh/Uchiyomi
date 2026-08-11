// SPDX-License-Identifier: AGPL-3.0-or-later

package pgmodels

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Comic struct {
	CreatedAt        time.Time      `gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"`
	Author           string         `gorm:"type:text"`
	Status           string         `gorm:"type:text"`
	Source           string         `gorm:"type:text;not null;uniqueIndex:idx_comic_source_slug,priority:1"`
	Description      string         `gorm:"type:text"`
	LocalCoverPath   string         `gorm:"type:text"`
	ExternalCoverURL string         `gorm:"type:text"`
	SourceURL        string         `gorm:"type:text"`
	Artist           string         `gorm:"type:text"`
	Slug             string         `gorm:"type:text;not null;uniqueIndex:idx_comic_source_slug,priority:2"`
	Title            string         `gorm:"type:text;not null"`
	ComicType        string         `gorm:"column:comic_type;type:text"`
	Genres           pq.StringArray `gorm:"type:text[];not null;default:'{}'"`
	AltTitles        pq.StringArray `gorm:"type:text[];not null;default:'{}'"`
	LibraryEntries   []LibraryEntry `gorm:"foreignKey:ComicID"`
	ChapterCount     int
	ReleaseYear      int
	Rating           float64
	ID               uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
}

func (Comic) TableName() string {
	return "comics"
}
