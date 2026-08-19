// SPDX-License-Identifier: AGPL-3.0-or-later

package pgmodels

import (
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readersettings"
)

type ReaderSettings struct {
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	ReadingMode string    `gorm:"type:text;not null"`
	PageScale   string    `gorm:"type:text;not null"`
	//nolint:lll
	ComicType  ComicType `gorm:"column:comic_type;type:text;not null;uniqueIndex:idx_reader_settings_user_type,priority:2"`
	User       User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_reader_settings_user_type,priority:1"`
	DoublePage bool      `gorm:"not null"`
}

func (ReaderSettings) TableName() string {
	return "reader_settings"
}

func (m *ReaderSettings) Domain() readersettings.Profile {
	return readersettings.Profile{
		Type:        m.ComicType.Domain(),
		ReadingMode: readersettings.ReadingMode(m.ReadingMode),
		PageScale:   readersettings.PageScale(m.PageScale),
		DoublePage:  m.DoublePage,
	}
}
