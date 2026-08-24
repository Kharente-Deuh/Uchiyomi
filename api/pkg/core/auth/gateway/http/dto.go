// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import "github.com/kharente-deuh/uchiyomi-server/pkg/utils/httputils"

type LoginWithPwdRequest struct {
	Username httputils.TrimmedString `json:"username" validate:"required"`
	Password httputils.TrimmedString `json:"password" validate:"required"`
}

type LoginWithPwdResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"isAdmin"`
}

type ProviderSummaryResponse struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`
}

type LogoutResponse struct {
	EndSessionURL string `json:"endSessionUrl,omitempty"`
}
