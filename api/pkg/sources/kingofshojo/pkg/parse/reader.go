// SPDX-License-Identifier: AGPL-3.0-or-later

package parse

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ParsePageURLs extracts http(s) image URLs from #readerarea.
// A nil result means goquery failed to parse the HTML.
// A non-nil empty slice means the document parsed but contained no usable images.
func ParsePageURLs(html string) []string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil
	}

	urls := make([]string, 0)

	doc.Find("#readerarea img").Each(func(_ int, img *goquery.Selection) {
		url := pageImageURL(img)
		if url == "" {
			return
		}

		urls = append(urls, url)
	})

	return urls
}

func pageImageURL(img *goquery.Selection) string {
	if dataSrc, exists := img.Attr("data-src"); exists && dataSrc != "" {
		if isHTTPURL(dataSrc) {
			return dataSrc
		}
	}

	if lazySrc, exists := img.Attr("data-lazy-src"); exists && lazySrc != "" {
		if isHTTPURL(lazySrc) {
			return lazySrc
		}
	}

	if src, exists := img.Attr("src"); exists && src != "" {
		if isHTTPURL(src) {
			return src
		}
	}

	return ""
}

func isHTTPURL(raw string) bool {
	return strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://")
}
