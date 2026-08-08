// SPDX-License-Identifier: AGPL-3.0-or-later

package http

type GetMeResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"isAdmin"`
}
