// SPDX-License-Identifier: AGPL-3.0-or-later

package feed

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
)

var _ FeedService = (*Service)(nil)

const maxLatestChapters = 3

type Deps struct {
	FeedRepository FeedRepository
	Now            func() time.Time
}

func (deps *Deps) Validate() error {
	if deps.FeedRepository == nil {
		return errors.New("feed repository is required")
	}

	if deps.Now == nil {
		return errors.New("now is required")
	}

	return nil
}

type Service struct {
	deps Deps
}

func NewService(deps Deps) (*Service, error) {
	if err := deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	return &Service{deps: deps}, nil
}

func (s *Service) Get(ctx context.Context, opts GetOpts) (Page, error) {
	now := s.deps.Now()

	page, err := s.deps.FeedRepository.ListPage(ctx, ListPageOpts{
		Now:    now,
		Source: opts.Source,
		Type:   opts.Type,
		UserID: opts.UserID,
		Limit:  opts.Limit,
		Offset: opts.Offset,
	})
	if err != nil {
		return Page{}, fmt.Errorf("s.deps.FeedRepository.ListPage: %w", err)
	}

	if len(page.Items) == 0 {
		if page.Items == nil {
			page.Items = []Item{}
		}

		return page, nil
	}

	ids := make([]uuid.UUID, len(page.Items))
	for i := range page.Items {
		ids[i] = page.Items[i].ID
	}

	chapters, err := s.deps.FeedRepository.ListUnlockedChapters(ctx, ListChaptersOpts{
		Now:      now,
		ComicIDs: ids,
		UserID:   opts.UserID,
	})
	if err != nil {
		return Page{}, fmt.Errorf("s.deps.FeedRepository.ListUnlockedChapters: %w", err)
	}

	byComic := map[uuid.UUID][]LatestChapter{}
	for i := range chapters {
		ch := chapters[i]
		byComic[ch.ComicID] = append(byComic[ch.ComicID], ch)
	}

	for i := range page.Items {
		chs := byComic[page.Items[i].ID]
		sort.SliceStable(chs, func(a, b int) bool {
			aa := AvailabilityAt(chs[a].PublishedAt, chs[a].EarlyAccessUntil, now)
			bb := AvailabilityAt(chs[b].PublishedAt, chs[b].EarlyAccessUntil, now)
			if !aa.Equal(bb) {
				return aa.After(bb)
			}

			return chs[a].Number > chs[b].Number
		})

		if len(chs) > maxLatestChapters {
			chs = chs[:maxLatestChapters]
		}

		if chs == nil {
			chs = []LatestChapter{}
		}

		page.Items[i].LatestChapters = chs
	}

	return page, nil
}
