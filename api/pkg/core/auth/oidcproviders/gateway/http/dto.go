// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/httputils"
)

type CreateProviderRequest struct {
	AdminClaim    *string                 `json:"adminClaim"`
	AllowedClaim  *string                 `json:"allowedClaim"`
	DisplayName   httputils.TrimmedString `json:"displayName" validate:"required,max=64"`
	IssuerURL     httputils.TrimmedString `json:"issuerUrl" validate:"required,url"`
	ClientID      httputils.TrimmedString `json:"clientId" validate:"required"`
	ClientSecret  httputils.TrimmedString `json:"clientSecret" validate:"required"`
	UsernameClaim httputils.TrimmedString `json:"usernameClaim" validate:"required"`
	Scopes        []string                `json:"scopes" validate:"required,min=1,dive,required"`
	AdminValues   []string                `json:"adminValues"`
	AllowedValues []string                `json:"allowedValues"`
	AutoProvision bool                    `json:"autoProvision"`
}

type UpdateProviderRequest struct {
	AdminClaim    *string                  `json:"adminClaim"`
	AllowedClaim  *string                  `json:"allowedClaim"`
	ClientSecret  *httputils.TrimmedString `json:"clientSecret" validate:"omitempty,min=1"`
	DisplayName   httputils.TrimmedString  `json:"displayName" validate:"required,max=64"`
	IssuerURL     httputils.TrimmedString  `json:"issuerUrl" validate:"required,url"`
	ClientID      httputils.TrimmedString  `json:"clientId" validate:"required"`
	UsernameClaim httputils.TrimmedString  `json:"usernameClaim" validate:"required"`
	Scopes        []string                 `json:"scopes" validate:"required,min=1,dive,required"`
	AdminValues   []string                 `json:"adminValues"`
	AllowedValues []string                 `json:"allowedValues"`
	AutoProvision bool                     `json:"autoProvision"`
}

type ProbeRequest struct {
	IssuerURL httputils.TrimmedString `json:"issuerUrl" validate:"required,url"`
}

type LightProviderResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type ProviderResponse struct {
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	AdminClaim    *string   `json:"adminClaim"`
	AllowedClaim  *string   `json:"allowedClaim"`
	ID            string    `json:"id"`
	DisplayName   string    `json:"displayName"`
	IssuerURL     string    `json:"issuerUrl"`
	ClientID      string    `json:"clientId"`
	UsernameClaim string    `json:"usernameClaim"`
	Scopes        []string  `json:"scopes"`
	AdminValues   []string  `json:"adminValues"`
	AllowedValues []string  `json:"allowedValues"`
	AutoProvision bool      `json:"autoProvision"`
}

type ProbeResponse struct {
	Issuer                    string `json:"issuer"`
	AuthorizationEndpoint     string `json:"authorizationEndpoint"`
	TokenEndpoint             string `json:"tokenEndpoint"`
	UserInfoEndpoint          string `json:"userInfoEndpoint"`
	EndSessionEndpoint        string `json:"endSessionEndpoint"`
	RedirectURI               string `json:"redirectUri"`
	SupportsRPInitiatedLogout bool   `json:"supportsRpInitiatedLogout"`
}
