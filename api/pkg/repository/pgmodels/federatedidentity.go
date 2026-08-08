// SPDX-License-Identifier: AGPL-3.0-or-later

package pgmodels

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type FederatedIdentity struct {
	LastLoginAt time.Time
	CreatedAt   time.Time    `gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"`
	Claims      Claims       `gorm:"type:jsonb;not null;default:'{}'"`
	Subject     string       `gorm:"type:text;not null;uniqueIndex:idx_fedid_provider_subject,priority:2"`
	Provider    OIDCProvider `gorm:"foreignKey:ProviderID"`
	User        User         `gorm:"foreignKey:UserID"`
	ID          uuid.UUID    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID      uuid.UUID    `gorm:"type:uuid;not null;index"`
	ProviderID  uuid.UUID    `gorm:"type:uuid;not null;uniqueIndex:idx_fedid_provider_subject,priority:1"`
}

func (FederatedIdentity) TableName() string {
	return "federated_identities"
}

type Claims map[string]any

func (c *Claims) Scan(src any) error {
	if src == nil {
		*c = nil

		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("claims: type %T inattendu", src)
	}

	err := json.Unmarshal(b, c)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	return nil
}

func (c Claims) Value() (driver.Value, error) {
	if c == nil {
		return []byte("{}"), nil
	}

	v, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("json.Marshall: %w", err)
	}

	return v, nil
}

func (Claims) GormDataType() string { return "jsonb" }
