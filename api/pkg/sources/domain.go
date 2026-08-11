// SPDX-License-Identifier: AGPL-3.0-or-later

package sources

import (
	"context"
	"fmt"
	"time"
)

type SourceMap map[SourceName]Source

type SourceName string

const (
	SourceAsuraScans SourceName = "asurascans"
)

func ParseSourceName(s string) (SourceName, error) {
	switch s {
	case string(SourceAsuraScans):
		return SourceAsuraScans, nil
	default:
		return "", fmt.Errorf("invalid source name: %s", s)
	}
}

type SeriesType string

const (
	SeriesTypeManga     SeriesType = "manga"
	SeriesTypeMangatoon SeriesType = "mangatoon"
	SeriesTypeManhua    SeriesType = "manhua"
	SeriesTypeManhwa    SeriesType = "manhwa"
)

func ParseSeriesType(s string) (SeriesType, error) {
	switch s {
	case "manga":
		return SeriesTypeManga, nil
	case "mangatoon":
		return SeriesTypeMangatoon, nil
	case "manhua":
		return SeriesTypeManhua, nil
	case "manhwa":
		return SeriesTypeManhwa, nil
	default:
		return "", fmt.Errorf("invalid series type: %s", s)
	}
}

type SeriesStatus string

const (
	SeriesStatusOngoing   SeriesStatus = "ongoing"
	SeriesStatusCompleted SeriesStatus = "completed"
	SeriesStatusHiatus    SeriesStatus = "hiatus"
	SeriesStatusCancelled SeriesStatus = "cancelled"
	SeriesStatusDropped   SeriesStatus = "dropped"
)

func ParseSeriesStatus(s string) (SeriesStatus, error) {
	switch s {
	case "ongoing":
		return SeriesStatusOngoing, nil
	case "completed":
		return SeriesStatusCompleted, nil
	case "hiatus":
		return SeriesStatusHiatus, nil
	case "cancelled":
		return SeriesStatusCancelled, nil
	case "dropped":
		return SeriesStatusDropped, nil
	default:
		return "", fmt.Errorf("invalid series status: %s", s)
	}
}

type Source interface {
	GetInfosBySlug(context.Context, string) (*GetInfosBySlugResponse, error)
}

type GetInfosBySlugResponse struct {
	LastChapterAt time.Time
	UpdatedAt     time.Time
	CreatedAt     time.Time
	Description   string
	Title         string
	Cover         string
	Status        SeriesStatus
	Type          SeriesType
	Author        string
	Artist        string
	SourceURL     string
	PublicURL     string
	Slug          string
	AltTitles     []string
	Genres        []string
	ChapterCount  int
	Rating        float64
}
