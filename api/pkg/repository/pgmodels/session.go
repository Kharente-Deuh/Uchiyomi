// SPDX-License-Identifier: AGPL-3.0-or-later

package pgmodels

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	ExpiresAt   time.Time  `gorm:"not null;index"`
	ProviderSID *string    `gorm:"column:provider_sid"`
	ProviderID  *uuid.UUID `gorm:"type:uuid;index"`
	AuthMethod  string     `gorm:"type:text;not null"`
	TokenHash   []byte     `gorm:"type:bytea;not null;uniqueIndex"`
	User        User       `gorm:"foreignKey:UserID"`
	ID          uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index"`
}

func (Session) TableName() string {
	return "sessions"
}
