// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readingprogress"
)

type progressResponse struct {
	UpdatedAt time.Time `json:"updatedAt"`
	ChapterID uuid.UUID `json:"chapterId"`
	Page      int       `json:"page"`
}

type continueResponse struct {
	ChapterID uuid.UUID `json:"chapterId"`
	Page      int       `json:"page"`
}

type listResponse struct {
	Continue *continueResponse  `json:"continue"`
	Chapters []progressResponse `json:"chapters"`
}

type saveRequest struct {
	ChapterID uuid.UUID `json:"chapterId" validate:"required"`
	Page      *int      `json:"page" validate:"required"`
}

func progressFromDomain(p readingprogress.Progress) progressResponse {
	return progressResponse{
		UpdatedAt: p.UpdatedAt,
		ChapterID: p.ChapterID,
		Page:      p.Page,
	}
}

func listFromDomain(result readingprogress.ListResult) listResponse {
	chapters := make([]progressResponse, 0, len(result.Chapters))
	for _, p := range result.Chapters {
		chapters = append(chapters, progressFromDomain(p))
	}

	var cont *continueResponse
	if result.Continue != nil {
		cont = &continueResponse{
			ChapterID: result.Continue.ChapterID,
			Page:      result.Continue.Page,
		}
	}

	return listResponse{
		Continue: cont,
		Chapters: chapters,
	}
}
