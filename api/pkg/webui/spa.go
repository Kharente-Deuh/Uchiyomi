// SPDX-License-Identifier: AGPL-3.0-or-later

package webui

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

const indexFile = "index.html"

const hashedDir = "_nuxt/"

const immutableCache = "public, max-age=31536000, immutable"

func newSPA(fsys fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

			return
		}

		name := resolve(fsys, r.URL.Path)
		if name == "" {
			http.NotFound(w, r)

			return
		}

		if strings.HasPrefix(name, hashedDir) {
			w.Header().Set("Cache-Control", immutableCache)
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}

		http.ServeFileFS(w, r, fsys, name)
	})
}

func resolve(fsys fs.FS, urlPath string) string {
	name := strings.TrimPrefix(path.Clean("/"+urlPath), "/")
	if name == "" {
		name = indexFile
	}

	if info, err := fs.Stat(fsys, name); err == nil && !info.IsDir() {
		return name
	}

	if path.Ext(name) != "" {
		return ""
	}

	if _, err := fs.Stat(fsys, indexFile); err != nil {
		return ""
	}

	return indexFile
}
