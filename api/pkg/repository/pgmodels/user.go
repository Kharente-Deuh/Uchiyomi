// SPDX-License-Identifier: AGPL-3.0-or-later

package pgmodels

import (
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
)

type User struct {
	CreatedAt           time.Time           `gorm:"autoCreateTime"`
	UpdatedAt           time.Time           `gorm:"autoUpdateTime"`
	Password            *PasswordCreds      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Name                string              `gorm:"type:text;not null;uniqueIndex"`
	FederatedIdentities []FederatedIdentity `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Sessions            []Session           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	ID                  uuid.UUID           `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	IsAdmin             bool
}

func (User) TableName() string {
	return "users"
}

func (u *User) Domain() users.User {
	return users.User{
		ID:        u.ID,
		Name:      u.Name,
		IsAdmin:   u.IsAdmin,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
