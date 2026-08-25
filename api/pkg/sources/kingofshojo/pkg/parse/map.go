// SPDX-License-Identifier: AGPL-3.0-or-later

package parse

import (
	"errors"
	"strconv"
	"strings"

	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

var ErrUnsupportedType = errors.New("unsupported series type")

func MapSeriesType(raw string) (sources.SeriesType, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "manga", "mangatoon":
		return sources.SeriesTypeMangatoon, nil
	case "manhwa":
		return sources.SeriesTypeManhwa, nil
	case "manhua":
		return sources.SeriesTypeManhua, nil
	case "comic", "novel":
		return "", ErrUnsupportedType
	default:
		return "", ErrUnsupportedType
	}
}

func MapSeriesStatus(raw string) sources.SeriesStatus {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "completed":
		return sources.SeriesStatusCompleted
	case "hiatus":
		return sources.SeriesStatusHiatus
	default:
		return sources.SeriesStatusOngoing
	}
}

func CleanPerson(raw string) string {
	s := strings.TrimSpace(raw)
	if strings.EqualFold(s, "n/a") {
		return ""
	}

	return s
}

func CleanAltTitles(raw string) []string {
	s := strings.TrimSpace(raw)
	if s == "" || strings.EqualFold(s, "unknown") {
		return nil
	}

	return []string{s}
}

func SourceChapterSlug(seriesSlug string, number float64) string {
	n := strconv.FormatFloat(number, 'f', -1, 64)
	n = strings.ReplaceAll(n, ".", "-")

	return seriesSlug + "-chapter-" + n
}
