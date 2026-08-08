// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !webui

package webui

import "net/http"

func Handler() (http.Handler, bool) {
	return nil, false
}
