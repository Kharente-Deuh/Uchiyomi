package http

import (
	"time"

	"github.com/google/uuid"
)

type postListBody struct {
	IDs []uuid.UUID `json:"ids" validate:"required,min=1,dive,required"`
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
