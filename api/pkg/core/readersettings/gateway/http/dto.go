// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readersettings"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

type profileResponse struct {
	Type        sources.SeriesType         `json:"type"`
	ReadingMode readersettings.ReadingMode `json:"readingMode"`
	PageScale   readersettings.PageScale   `json:"pageScale"`
	DoublePage  bool                       `json:"doublePage"`
}

type listResponse struct {
	Items []profileResponse `json:"items"`
}

type replaceRequest struct {
	DoublePage  *bool  `json:"doublePage" validate:"required"`
	ReadingMode string `json:"readingMode" validate:"required"`
	PageScale   string `json:"pageScale" validate:"required"`
}

func profileFromDomain(p readersettings.Profile) profileResponse {
	return profileResponse{
		Type:        p.Type,
		ReadingMode: p.ReadingMode,
		PageScale:   p.PageScale,
		DoublePage:  p.DoublePage,
	}
}

func listFromDomain(profiles []readersettings.Profile) listResponse {
	items := make([]profileResponse, 0, len(profiles))
	for _, p := range profiles {
		items = append(items, profileFromDomain(p))
	}

	return listResponse{Items: items}
}
