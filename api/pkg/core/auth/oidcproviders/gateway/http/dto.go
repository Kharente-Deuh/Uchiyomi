// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"errors"
	"strings"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/httputils"
)

type CreateProviderRequest struct {
	RoleClaim     *string                 `json:"roleClaim"`
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
	RoleClaim     *string                 `json:"roleClaim"`
	DisplayName   httputils.TrimmedString `json:"displayName" validate:"required,max=64"`
	IssuerURL     httputils.TrimmedString `json:"issuerUrl" validate:"required,url"`
	ClientID      httputils.TrimmedString `json:"clientId" validate:"required"`
	UsernameClaim httputils.TrimmedString `json:"usernameClaim" validate:"required"`
	Scopes        []string                `json:"scopes" validate:"required,min=1,dive,required"`
	AdminValues   []string                `json:"adminValues"`
	AllowedValues []string                `json:"allowedValues"`
	AutoProvision bool                    `json:"autoProvision"`
}

var errRoleClaimRequired = errors.New("roleClaim is required when adminValues or allowedValues is set")

func (r CreateProviderRequest) validate() error {
	return validateRoleClaim(r.RoleClaim, r.AdminValues, r.AllowedValues)
}

func (r UpdateProviderRequest) validate() error {
	return validateRoleClaim(r.RoleClaim, r.AdminValues, r.AllowedValues)
}

func validateRoleClaim(claim *string, adminValues, allowedValues []string) error {
	if claim != nil && strings.TrimSpace(*claim) != "" {
		return nil
	}

	if len(adminValues) > 0 || len(allowedValues) > 0 {
		return errRoleClaimRequired
	}

	return nil
}

type ProbeRequest struct {
	IssuerURL httputils.TrimmedString `json:"issuerUrl" validate:"required,url"`
}

type LightProviderResponse struct {
	CreatedAt   time.Time `json:"createdAt"`
	ID          string    `json:"id"`
	DisplayName string    `json:"displayName"`
	UserCount   int64     `json:"userCount"`
}

type ProviderResponse struct {
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	RoleClaim     *string   `json:"roleClaim"`
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

type ProviderUserResponse struct {
	LinkedAt time.Time `json:"linkedAt"`
	ID       string    `json:"id"`
	Username string    `json:"username"`
	IsAdmin  bool      `json:"isAdmin"`
}

type ProviderDetailsResponse struct {
	Users []ProviderUserResponse `json:"users"`
	ProviderResponse
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
