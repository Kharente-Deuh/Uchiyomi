// SPDX-License-Identifier: AGPL-3.0-or-later

package readersettings

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

var _ ReaderSettingsService = (*Service)(nil)

type Deps struct {
	Repository Repository
}

func (deps *Deps) Validate() error {
	if deps.Repository == nil {
		return errors.New("repository is required")
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

func (s *Service) ListForUser(ctx context.Context, userID uuid.UUID) ([]Profile, error) {
	stored, err := s.deps.Repository.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("s.deps.Repository.ListByUser: %w", err)
	}

	byType := make(map[sources.SeriesType]Profile, len(stored))
	for _, profile := range stored {
		if !knownType(profile.Type) {
			continue
		}

		byType[profile.Type] = profile
	}

	types := AllTypes()
	out := make([]Profile, 0, len(types))
	for _, typ := range types {
		if profile, ok := byType[typ]; ok {
			out = append(out, profile)

			continue
		}

		out = append(out, DefaultProfile(typ))
	}

	return out, nil
}

func (s *Service) Replace(ctx context.Context, opts ReplaceOpts) (Profile, error) {
	if !knownType(opts.Type) {
		return Profile{}, ErrInvalid
	}

	if opts.ReadingMode == ReadingModeWebtoon && opts.DoublePage {
		return Profile{}, ErrInvalid
	}

	profile, err := s.deps.Repository.Upsert(ctx, UpsertOpts(opts))
	if err != nil {
		return Profile{}, fmt.Errorf("s.deps.Repository.Upsert: %w", err)
	}

	return profile, nil
}

func knownType(t sources.SeriesType) bool {
	for _, typ := range AllTypes() {
		if typ == t {
			return true
		}
	}

	return false
}
