// SPDX-License-Identifier: AGPL-3.0-or-later

package domain

import "fmt"

func (o SearchCacheOpts) CacheKey() string {
	return fmt.Sprintf(
		"offset=%d limit=%d search=%q sort=%q order=%q",
		o.Offset, o.Limit, o.Search, o.Sort, o.SortOrder,
	)
}

func (o GetImageURLsByChapterOpts) CacheKey() string {
	return fmt.Sprintf("series=%q chapter=%q", o.SeriesSlug, o.ChapterID)
}
