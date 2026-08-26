// SPDX-License-Identifier: AGPL-3.0-or-later

package pg

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/comics"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgmodels"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ comics.ComicsRepository = (*PGComicsRepository)(nil)

type Deps struct {
	DB *gorm.DB
}

func (deps *Deps) Validate() error {
	if deps.DB == nil {
		return errors.New("db is required")
	}

	return nil
}

type PGComicsRepository struct {
	deps Deps
}

func New(deps Deps) (*PGComicsRepository, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	r := &PGComicsRepository{deps: deps}

	return r, nil
}

func (r *PGComicsRepository) db(ctx context.Context) gorm.Interface[pgmodels.Comic] {
	return gorm.G[pgmodels.Comic](pgtx.From(ctx, r.deps.DB))
}

func (r *PGComicsRepository) GetByID(ctx context.Context, opts comics.GetByIDOpts) (*comics.Comic, error) {
	model, err := r.db(ctx).
		Where(`comics.id = ? AND EXISTS (
			SELECT 1 FROM library_entries
			WHERE library_entries.comic_id = comics.id AND library_entries.user_id = ?
		)`, opts.ID, opts.UserID).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := model.Domain()

	return &ret, nil
}

func (r *PGComicsRepository) FindByID(ctx context.Context, id uuid.UUID) (*comics.Comic, error) {
	model, err := r.db(ctx).Where("comics.id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := model.Domain()

	return &ret, nil
}

//nolint:lll
func (r *PGComicsRepository) GetBySourceSlug(ctx context.Context, opts comics.GetBySourceSlugOpts) (*comics.Comic, error) {
	model, err := r.db(ctx).
		Where(`comics.slug = ? AND comics.source = ? AND EXISTS (
			SELECT 1 FROM library_entries
			WHERE library_entries.comic_id = comics.id AND library_entries.user_id = ?
		)`, opts.Slug, opts.Source, opts.UserID).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := model.Domain()

	return &ret, nil
}

func (r *PGComicsRepository) FindBySourceSlug(
	ctx context.Context, opts comics.FindBySourceSlugOpts,
) (*comics.Comic, error) {
	model, err := r.db(ctx).
		Where("comics.slug = ? AND comics.source = ?", opts.Slug, opts.Source).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := model.Domain()

	return &ret, nil
}

func (r *PGComicsRepository) Create(ctx context.Context, opts comics.CreateComicOpts) (*comics.Comic, error) {
	now := time.Now()

	model := &pgmodels.Comic{
		ID:           opts.ID,
		Source:       opts.Source,
		Slug:         opts.Slug,
		Title:        opts.Title,
		Status:       pgmodels.ComicStatusFromDomain(opts.Status),
		ComicType:    pgmodels.ComicTypeFromDomain(opts.Type),
		Genres:       pq.StringArray(opts.Genres),
		ChapterCount: opts.ChapterCount,
		Author:       opts.Author,
		Artist:       opts.Artist,
		Description:  opts.Description,
		AltTitles:    pq.StringArray(opts.AltTitles),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err := r.db(ctx).Create(ctx, model)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, domain.ErrAlreadyExists
		}

		return nil, fmt.Errorf("r.db(ctx).Create: %w", err)
	}

	ret := model.Domain()

	return &ret, nil
}

func likeContains(q string) string {
	q = strings.ReplaceAll(q, `\`, `\\`)
	q = strings.ReplaceAll(q, `%`, `\%`)
	q = strings.ReplaceAll(q, `_`, `\_`)

	return `%` + q + `%`
}

func (r *PGComicsRepository) getManyQuery(ctx context.Context, opts comics.GetManyOpts) *gorm.DB {
	db := pgtx.From(ctx, r.deps.DB).Model(&pgmodels.Comic{})

	if opts.UserID != nil {
		db = db.Joins(
			"JOIN library_entries ON library_entries.comic_id = comics.id AND library_entries.user_id = ?",
			*opts.UserID,
		)
	}

	if opts.Source != nil {
		db = db.Where("comics.source = ?", *opts.Source)
	}

	if opts.Type != nil {
		db = db.Where("comics.comic_type = ?", *opts.Type)
	}

	if opts.Status != nil {
		db = db.Where("comics.status = ?", *opts.Status)
	}

	if opts.Search != "" {
		pattern := likeContains(opts.Search)
		//nolint:lll
		db = db.Where(
			`(comics.title ILIKE ? ESCAPE '\' OR EXISTS (
				SELECT 1 FROM unnest(comics.alt_titles) AS alt(title)
				WHERE alt.title ILIKE ? ESCAPE '\'
			))`,
			pattern,
			pattern,
		)
	}

	return db
}

func getManyOrderSQL(opts comics.GetManyOpts) string {
	dir := "ASC"
	if opts.Order == comics.ListOrderDesc {
		dir = "DESC"
	}

	if opts.Sort == comics.ListSortAddedAt {
		return "library_entries.added_at " + dir
	}

	return "LOWER(comics.title) " + dir
}

func (r *PGComicsRepository) GetMany(ctx context.Context, opts comics.GetManyOpts) (comics.Page, error) {
	var total int64

	if err := r.getManyQuery(ctx, opts).Count(&total).Error; err != nil {
		return comics.Page{}, fmt.Errorf("r.getManyQuery.Count: %w", err)
	}

	var models []pgmodels.Comic

	err := r.getManyQuery(ctx, opts).
		Order(getManyOrderSQL(opts)).
		Offset(opts.Offset).
		Limit(opts.Limit).
		Find(&models).Error
	if err != nil {
		return comics.Page{}, fmt.Errorf("r.getManyQuery.Find: %w", err)
	}

	ret := make([]comics.Comic, len(models))
	for i, model := range models {
		ret[i] = model.Domain()
	}

	return comics.Page{Items: ret, Total: total}, nil
}

func (r *PGComicsRepository) ListByStatuses(
	ctx context.Context,
	opts comics.ListByStatusesOpts,
) ([]comics.Comic, error) {
	if len(opts.Statuses) == 0 {
		return nil, nil
	}

	models, err := r.db(ctx).
		Where("comics.source = ? AND comics.status IN ?", opts.Source, opts.Statuses).
		Order("comics.created_at ASC").
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := make([]comics.Comic, len(models))
	for i, model := range models {
		ret[i] = model.Domain()
	}

	return ret, nil
}

func (r *PGComicsRepository) UpdateStatusAndChapterCount(
	ctx context.Context,
	opts comics.UpdateStatusAndChapterCountOpts,
) error {
	values := map[string]any{
		"status":        pgmodels.ComicStatusFromDomain(opts.Status),
		"chapter_count": opts.ChapterCount,
		"updated_at":    time.Now(),
	}

	rows, err := r.db(ctx).Where("id = ?", opts.ID).Set(clause.Assignments(values)).Update(ctx)
	if err != nil {
		return fmt.Errorf("r.db(ctx).Updates: %w", err)
	}

	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *PGComicsRepository) UpdateType(
	ctx context.Context,
	opts comics.UpdateTypeOpts,
) error {
	values := map[string]any{
		"comic_type": pgmodels.ComicTypeFromDomain(opts.Type),
		"updated_at": time.Now(),
	}

	rows, err := r.db(ctx).Where("id = ?", opts.ID).Set(clause.Assignments(values)).Update(ctx)
	if err != nil {
		return fmt.Errorf("r.db(ctx).Update: %w", err)
	}

	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// nolint:lll
func (r *PGComicsRepository) GetBySlugsAndSource(ctx context.Context, opts comics.GetBySlugsAndSource) ([]comics.Comic, error) {
	if len(opts.Slugs) == 0 {
		return nil, nil
	}

	models, err := r.db(ctx).
		Where(
			`comics.source = ? AND comics.slug IN ? AND EXISTS (
				SELECT 1 FROM library_entries
				WHERE library_entries.comic_id = comics.id AND library_entries.user_id = ?
			)`,
			opts.Source, opts.Slugs, opts.UserID,
		).Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	ret := make([]comics.Comic, len(models))
	for i, model := range models {
		ret[i] = model.Domain()
	}

	return ret, nil
}

func (r *PGComicsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	affected, err := r.db(ctx).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("r.db(ctx).Where: %w", err)
	}

	if affected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
