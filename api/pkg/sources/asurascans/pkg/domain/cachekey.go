// SPDX-License-Identifier: AGPL-3.0-or-later

package domain

import (
	"fmt"
	"slices"
)

func (o SearchOpts) CacheKey() string {
	genres := slices.Clone(o.Genres)
	slices.Sort(genres)
	genres = slices.Compact(genres)

	return fmt.Sprintf(
		"page=%d search=%q sort=%q order=%q status=%q type=%q artist=%q genres=%q minchapters=%d",
		o.Page, o.Search, o.Sort, o.SortOrder,
		o.Status, o.Type, o.Artist, genres, o.MinChapters,
	)
}

func (o GetImageURLsByChapterOpts) CacheKey() string {
	return fmt.Sprintf("series=%q chapter=%q", o.SeriesSlug, o.ChapterID)
}
