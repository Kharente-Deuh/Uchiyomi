package http

import (
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
)

type postListBody struct {
	IDs []uuid.UUID `json:"ids" validate:"required,min=1,dive,required"`
}

type postListResponse struct {
	Chapters []postListResponseChapter `json:"chapters"`
}

func postListResponseFromDomain(domain []chapters.Chapter) postListResponse {
	chapters := make([]postListResponseChapter, 0, len(domain))
	for _, chapter := range domain {
		var earlyAccessUntil *time.Time = nil
		if !chapter.EarlyAccessUntil.IsZero() {
			earlyAccessUntil = &chapter.EarlyAccessUntil
		}

		chapters = append(chapters, postListResponseChapter{
			PublishedAt:       chapter.PublishedAt,
			EarlyAccessUntil:  earlyAccessUntil,
			SourceChapterSlug: chapter.SourceChapterSlug,
			Title:             chapter.Title,
			Number:            chapter.Number,
			PagesNb:           chapter.PagesNb,
			Download:          chapter.Download,
			ID:                chapter.ID,
			ComicID:           chapter.ComicID,
		})
	}

	return postListResponse{
		Chapters: chapters,
	}
}

type postListResponseChapter struct {
	PublishedAt       time.Time  `json:"publishedAt"`
	EarlyAccessUntil  *time.Time `json:"earlyAccessUntil"`
	SourceChapterSlug string     `json:"sourceChapterSlug"`
	Title             string     `json:"title"`
	Number            float64    `json:"number"`
	PagesNb           int        `json:"pagesNb"`
	Download          int        `json:"download"`
	ID                uuid.UUID  `json:"id"`
	ComicID           uuid.UUID  `json:"comicId"`
}
