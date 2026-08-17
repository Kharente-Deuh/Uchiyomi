// SPDX-License-Identifier: AGPL-3.0-or-later

package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/feed"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
	"gorm.io/gorm"
)

var _ feed.FeedRepository = (*PGFeedRepository)(nil)

const availabilitySQL = `CASE
  WHEN chapters.early_access_until > TIMESTAMPTZ '0001-01-01 00:00:00+00'
   AND chapters.early_access_until <= ?
  THEN chapters.early_access_until
  ELSE chapters.published_at
END`

type Deps struct {
	DB *gorm.DB
}

func (deps *Deps) Validate() error {
	if deps.DB == nil {
		return errors.New("db is required")
	}

	return nil
}

type PGFeedRepository struct {
	deps Deps
}

func New(deps Deps) (*PGFeedRepository, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	r := &PGFeedRepository{deps: deps}

	return r, nil
}

type pageRow struct {
	Title  string
	Slug   string
	Source sources.SourceName
	Status sources.SeriesStatus
	Type   sources.SeriesType `gorm:"column:comic_type"`
	ID     uuid.UUID
}

type chapterRow struct {
	PublishedAt      time.Time
	EarlyAccessUntil time.Time
	Title            string
	Number           float64
	Download         int
	ID               uuid.UUID
	ComicID          uuid.UUID
}

func (r *PGFeedRepository) ListPage(ctx context.Context, opts feed.ListPageOpts) (feed.Page, error) {
	countSQL, countArgs := listPageCountSQL(opts)

	var total int64

	err := pgtx.From(ctx, r.deps.DB).Raw(countSQL, countArgs...).Scan(&total).Error
	if err != nil {
		return feed.Page{}, fmt.Errorf("count: %w", err)
	}

	pageSQL, pageArgs := listPageSQL(opts)

	var rows []pageRow

	err = pgtx.From(ctx, r.deps.DB).Raw(pageSQL, pageArgs...).Scan(&rows).Error
	if err != nil {
		return feed.Page{}, fmt.Errorf("page: %w", err)
	}

	items := make([]feed.Item, len(rows))
	for i := range rows {
		items[i] = feed.Item{
			ID:     rows[i].ID,
			Title:  rows[i].Title,
			Slug:   rows[i].Slug,
			Source: rows[i].Source,
			Status: rows[i].Status,
			Type:   rows[i].Type,
		}
	}

	return feed.Page{Items: items, Total: total}, nil
}

func (r *PGFeedRepository) ListUnlockedChapters(
	ctx context.Context, opts feed.ListChaptersOpts,
) ([]feed.LatestChapter, error) {
	if len(opts.ComicIDs) == 0 {
		return nil, nil
	}

	const sql = `
SELECT id, comic_id, title, number, published_at, early_access_until, download
FROM chapters
WHERE comic_id IN ? AND early_access_until <= ?
`

	var rows []chapterRow

	err := pgtx.From(ctx, r.deps.DB).Raw(sql, opts.ComicIDs, opts.Now).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("chapters: %w", err)
	}

	ret := make([]feed.LatestChapter, len(rows))
	for i := range rows {
		ret[i] = feed.LatestChapter{
			ID:               rows[i].ID,
			ComicID:          rows[i].ComicID,
			Title:            rows[i].Title,
			Number:           rows[i].Number,
			PublishedAt:      rows[i].PublishedAt,
			EarlyAccessUntil: rows[i].EarlyAccessUntil,
			Download:         rows[i].Download,
		}
	}

	return ret, nil
}

func listPageCountSQL(opts feed.ListPageOpts) (string, []any) {
	sql := `
SELECT COUNT(DISTINCT comics.id)
FROM comics
INNER JOIN library_entries ON library_entries.comic_id = comics.id AND library_entries.user_id = ?
INNER JOIN chapters ON chapters.comic_id = comics.id AND chapters.early_access_until <= ?
`

	args := make([]any, 0, 5)
	args = append(args, opts.UserID, opts.Now)

	return appendPageFilters(sql, args, opts)
}

func listPageSQL(opts feed.ListPageOpts) (string, []any) {
	sql := `
SELECT comics.id, comics.source, comics.slug, comics.title, comics.status, comics.comic_type
FROM comics
INNER JOIN library_entries ON library_entries.comic_id = comics.id AND library_entries.user_id = ?
INNER JOIN chapters ON chapters.comic_id = comics.id AND chapters.early_access_until <= ?
`

	args := make([]any, 0, 8)
	args = append(args, opts.UserID, opts.Now)
	sql, args = appendPageFilters(sql, args, opts)

	sql += `
GROUP BY comics.id, comics.source, comics.slug, comics.title, comics.status, comics.comic_type
ORDER BY MAX(` + availabilitySQL + `) DESC, comics.title ASC
LIMIT ? OFFSET ?
`
	args = append(args, opts.Now, opts.Limit, opts.Offset)

	return sql, args
}

func appendPageFilters(sql string, args []any, opts feed.ListPageOpts) (string, []any) {
	if opts.Source != nil {
		sql += " AND comics.source = ?"
		args = append(args, *opts.Source)
	}

	if opts.Type != nil {
		sql += " AND comics.comic_type = ?"
		args = append(args, *opts.Type)
	}

	if opts.Status != nil {
		sql += " AND comics.status = ?"
		args = append(args, *opts.Status)
	}

	return sql, args
}
