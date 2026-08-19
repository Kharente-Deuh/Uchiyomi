// SPDX-License-Identifier: AGPL-3.0-or-later

package readersettings

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

var ErrInvalid = errors.New("invalid reader settings")

type ReadingMode string

const (
	ReadingModePagedRTL ReadingMode = "paged-rtl"
	ReadingModePagedLTR ReadingMode = "paged-ltr"
	ReadingModeWebtoon  ReadingMode = "webtoon"
)

type PageScale string

const (
	PageScaleFitWidth  PageScale = "fit-width"
	PageScaleFitHeight PageScale = "fit-height"
	PageScaleFitScreen PageScale = "fit-screen"
)

type Profile struct {
	Type        sources.SeriesType
	ReadingMode ReadingMode
	PageScale   PageScale
	DoublePage  bool
}

type ReplaceOpts struct {
	ReadingMode ReadingMode
	PageScale   PageScale
	Type        sources.SeriesType
	UserID      uuid.UUID
	DoublePage  bool
}

type UpsertOpts struct {
	ReadingMode ReadingMode
	PageScale   PageScale
	Type        sources.SeriesType
	UserID      uuid.UUID
	DoublePage  bool
}

type Repository interface {
	ListByUser(context.Context, uuid.UUID) ([]Profile, error)
	Upsert(context.Context, UpsertOpts) (Profile, error)
}

type ReaderSettingsService interface {
	ListForUser(context.Context, uuid.UUID) ([]Profile, error)
	Replace(context.Context, ReplaceOpts) (Profile, error)
}

func AllTypes() []sources.SeriesType {
	return []sources.SeriesType{
		sources.SeriesTypeManga,
		sources.SeriesTypeManhua,
		sources.SeriesTypeManhwa,
		sources.SeriesTypeMangatoon,
	}
}

func DefaultProfile(t sources.SeriesType) Profile {
	switch t {
	case sources.SeriesTypeManga:
		return Profile{
			Type:        sources.SeriesTypeManga,
			ReadingMode: ReadingModePagedRTL,
			PageScale:   PageScaleFitScreen,
		}
	case sources.SeriesTypeManhua:
		return Profile{
			Type:        sources.SeriesTypeManhua,
			ReadingMode: ReadingModePagedLTR,
			PageScale:   PageScaleFitScreen,
		}
	case sources.SeriesTypeManhwa:
		return Profile{
			Type:        sources.SeriesTypeManhwa,
			ReadingMode: ReadingModeWebtoon,
			PageScale:   PageScaleFitWidth,
		}
	case sources.SeriesTypeMangatoon:
		return Profile{
			Type:        sources.SeriesTypeMangatoon,
			ReadingMode: ReadingModeWebtoon,
			PageScale:   PageScaleFitWidth,
		}
	default:
		return Profile{Type: t}
	}
}

func ParseReadingMode(s string) (ReadingMode, error) {
	switch s {
	case string(ReadingModePagedRTL):
		return ReadingModePagedRTL, nil
	case string(ReadingModePagedLTR):
		return ReadingModePagedLTR, nil
	case string(ReadingModeWebtoon):
		return ReadingModeWebtoon, nil
	default:
		return "", fmt.Errorf("%w: reading mode %q", ErrInvalid, s)
	}
}

func ParsePageScale(s string) (PageScale, error) {
	switch s {
	case string(PageScaleFitWidth):
		return PageScaleFitWidth, nil
	case string(PageScaleFitHeight):
		return PageScaleFitHeight, nil
	case string(PageScaleFitScreen):
		return PageScaleFitScreen, nil
	default:
		return "", fmt.Errorf("%w: page scale %q", ErrInvalid, s)
	}
}
