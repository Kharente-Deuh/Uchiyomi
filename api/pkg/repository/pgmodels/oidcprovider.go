// SPDX-License-Identifier: AGPL-3.0-or-later

package pgmodels

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type OIDCProvider struct {
	UpdatedAt           time.Time `gorm:"autoUpdateTime"`
	CreatedAt           time.Time `gorm:"autoCreateTime"`
	AdminClaim          *string
	AllowedClaim        *string
	ClientID            string
	UsernameClaim       string
	IssuerURL           string `gorm:"type:text;not null;uniqueIndex"`
	DisplayName         string
	Scopes              pq.StringArray `gorm:"type:text[];not null;default:'{openid,profile}'"`
	ClientSecretEnc     []byte
	AdminValues         pq.StringArray      `gorm:"type:text[]"`
	AllowedValues       pq.StringArray      `gorm:"type:text[]"`
	FederatedIdentities []FederatedIdentity `gorm:"foreignKey:ProviderID;constraint:OnDelete:CASCADE"`
	ID                  uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AutoProvision       bool                `gorm:"not null;default:true"`
}

func (OIDCProvider) TableName() string {
	return "oidc_providers"
}
