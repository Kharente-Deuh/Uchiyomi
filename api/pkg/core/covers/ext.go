// SPDX-License-Identifier: AGPL-3.0-or-later

package covers

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
)

func ExtensionFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	ext := path.Ext(u.Path)
	if ext == "" || ext == "." {
		return ""
	}

	return ext
}

func ExtensionFromContentType(contentType string) string {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return ""
	}

	exts, err := mime.ExtensionsByType(mediaType)
	if err != nil || len(exts) == 0 {
		return ""
	}

	return exts[0]
}

func MIMEForExtension(ext string) string {
	mediaType := mime.TypeByExtension(ext)
	if mediaType == "" {
		return "application/octet-stream"
	}

	return mediaType
}

func ResolveAbsoluteURL(base, cover string) string {
	if strings.HasPrefix(cover, "http://") || strings.HasPrefix(cover, "https://") {
		return cover
	}

	base = strings.TrimSuffix(base, "/")
	if strings.HasPrefix(cover, "/") {
		return base + cover
	}

	return base + "/" + cover
}

func cacheKey(source, slug, ext string) string {
	return source + "/" + slug + ext
}

func parseCacheKey(key string) (source, slug, ext string, err error) {
	parts := strings.SplitN(key, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", fmt.Errorf("invalid cache key %q", key)
	}

	source = parts[0]
	ext = path.Ext(parts[1])
	if ext == "" {
		slug = parts[1]

		return source, slug, ext, nil
	}

	slug = strings.TrimSuffix(parts[1], ext)

	return source, slug, ext, nil
}

func probeExtension(ctx context.Context, client *http.Client, rawURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("client.Do: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return "", ErrSeriesNotFound
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("unexpected status %d", res.StatusCode)
	}

	ext := ExtensionFromContentType(res.Header.Get("Content-Type"))
	if ext == "" {
		ext = ".bin"
	}

	return ext, nil
}
