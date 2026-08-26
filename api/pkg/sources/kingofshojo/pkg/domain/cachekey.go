// SPDX-License-Identifier: AGPL-3.0-or-later

package domain

import "fmt"

func (o SearchCacheOpts) CacheKey() string {
	return fmt.Sprintf("page=%d search=%q sort=%q order=%q", o.Page, o.Search, o.Sort, o.SortOrder)
}

func (o GetImageURLsByChapterOpts) CacheKey() string {
	return fmt.Sprintf("series=%q chapter=%q", o.SeriesSlug, o.ChapterID)
}
