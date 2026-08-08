// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build webui

package webui

import (
	"embed"
	"io/fs"
	"net/http"
)

//nolint:gochecknoglobals // embed impose une variable de paquet.
//go:embed all:dist
var assets embed.FS

func Handler() (http.Handler, bool) {
	sub, err := fs.Sub(assets, "dist")
	if err != nil {
		panic(err)
	}

	return newSPA(sub), true
}
