// SPDX-License-Identifier: AGPL-3.0-or-later

package pgmodels

import (
	"time"

	"github.com/google/uuid"
)

type PasswordCreds struct {
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	Hash      string
	User      User      `gorm:"foreignKey:UserID"`
	UserID    uuid.UUID `gorm:"primaryKey;type:uuid"`
}

func (PasswordCreds) TableName() string {
	return "password_credentials"
}
