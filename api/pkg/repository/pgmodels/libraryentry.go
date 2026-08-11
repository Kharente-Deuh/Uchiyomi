// SPDX-License-Identifier: AGPL-3.0-or-later

package pgmodels

import (
	"time"

	"github.com/google/uuid"
)

type LibraryEntry struct {
	AddedAt time.Time `gorm:"not null"`
	User    User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Comic   Comic     `gorm:"foreignKey:ComicID"`
	ID      uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_library_entry_user_comic,priority:1;index"`
	ComicID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_library_entry_user_comic,priority:2;index"`
}

func (LibraryEntry) TableName() string {
	return "library_entries"
}
