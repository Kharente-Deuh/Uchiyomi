// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/httputils"
)

type GetStatusResponse struct {
	Required bool `json:"required"`
}

type DoSetupRequest struct {
	Username httputils.TrimmedString `json:"username" validate:"required,alphanum"`
	Password httputils.TrimmedString `json:"password" validate:"required,min=8,printascii"`
}
