// SPDX-License-Identifier: AGPL-3.0-or-later

package domain

import (
	"context"
	"slices"
	"time"
)

type ApiClient interface {
	Search(context.Context, SearchOpts) (*SearchResult, error)
	GetInfosBySlug(context.Context, string) (*GetInfosBySlugResponse, error)
	GetChaptersListBySerie(context.Context, string) (*[]Chapter, error)
	GetImageURLsByChapter(context.Context, GetImageURLsByChapterOpts) (*[]string, error)
}

type SeriesType string

const (
	SeriesTypeManga     SeriesType = "manga"
	SeriesTypeMangatoon SeriesType = "mangatoon"
	SeriesTypeManhua    SeriesType = "manhua"
	SeriesTypeManhwa    SeriesType = "manhwa"
	SeriesTypeNone      SeriesType = ""
)

func IsSeriesType(s string) bool {
	values := []SeriesType{
		SeriesTypeManga,
		SeriesTypeMangatoon,
		SeriesTypeManhua,
		SeriesTypeManhwa,
		SeriesTypeNone,
	}

	return slices.Contains(values, SeriesType(s))
}

type SeriesGenre string

const (
	SeriesGenreAction        SeriesGenre = "action"
	SeriesGenreAdventure     SeriesGenre = "adventure"
	SeriesGenreComedy        SeriesGenre = "comedy"
	SeriesGenreCrazyMc       SeriesGenre = "crazy-mc"
	SeriesGenreDarkFantasy   SeriesGenre = "dark-fantasy"
	SeriesGenreDemon         SeriesGenre = "demon"
	SeriesGenreDrama         SeriesGenre = "drama"
	SeriesGenreDungeons      SeriesGenre = "dungeons"
	SeriesGenreFantasy       SeriesGenre = "fantasy"
	SeriesGenreGame          SeriesGenre = "game"
	SeriesGenreGeniusMc      SeriesGenre = "genius-mc"
	SeriesGenreIsekai        SeriesGenre = "isekai"
	SeriesGenreKuchikuchi    SeriesGenre = "kuchikuchi"
	SeriesGenreMagic         SeriesGenre = "magic"
	SeriesGenreMartialArts   SeriesGenre = "martial-arts"
	SeriesGenreMurim         SeriesGenre = "murim"
	SeriesGenreMystery       SeriesGenre = "mystery"
	SeriesGenreNecromancer   SeriesGenre = "necromancer"
	SeriesGenreOverpowered   SeriesGenre = "overpowered"
	SeriesGenrePsychological SeriesGenre = "psychological"
	SeriesGenreRegression    SeriesGenre = "regression"
	SeriesGenreReincarnation SeriesGenre = "reincarnation"
	SeriesGenreRevenge       SeriesGenre = "revenge"
	SeriesGenreRomance       SeriesGenre = "romance"
	SeriesGenreSchoolLife    SeriesGenre = "school-life"
	SeriesGenreSciFi         SeriesGenre = "sci-fi"
	SeriesGenreShoujo        SeriesGenre = "shoujo"
	SeriesGenreShounen       SeriesGenre = "shounen"
	SeriesGenreSystem        SeriesGenre = "system"
	SeriesGenreTower         SeriesGenre = "tower"
	SeriesGenreTragedy       SeriesGenre = "tragedy"
	SeriesGenreVillain       SeriesGenre = "villain"
	SeriesGenreViolence      SeriesGenre = "violence"
)

func IsSeriesGenre(s string) bool {
	values := []SeriesGenre{
		SeriesGenreAction,
		SeriesGenreAdventure,
		SeriesGenreComedy,
		SeriesGenreCrazyMc,
		SeriesGenreDarkFantasy,
		SeriesGenreDemon,
		SeriesGenreDrama,
		SeriesGenreDungeons,
		SeriesGenreFantasy,
		SeriesGenreGame,
		SeriesGenreGeniusMc,
		SeriesGenreIsekai,
		SeriesGenreKuchikuchi,
		SeriesGenreMagic,
		SeriesGenreMartialArts,
		SeriesGenreMurim,
		SeriesGenreMystery,
		SeriesGenreNecromancer,
		SeriesGenreOverpowered,
		SeriesGenrePsychological,
		SeriesGenreRegression,
		SeriesGenreReincarnation,
		SeriesGenreRevenge,
		SeriesGenreRomance,
		SeriesGenreSchoolLife,
		SeriesGenreSciFi,
		SeriesGenreShoujo,
		SeriesGenreShounen,
		SeriesGenreSystem,
		SeriesGenreTower,
		SeriesGenreTragedy,
		SeriesGenreVillain,
		SeriesGenreViolence,
	}

	return slices.Contains(values, SeriesGenre(s))
}

type SeriesStatus string

const (
	SeriesStatusOngoing   SeriesStatus = "ongoing"
	SeriesStatusCompleted SeriesStatus = "completed"
	SeriesStatusHiatus    SeriesStatus = "hiatus"
	SeriesStatusCancelled SeriesStatus = "cancelled"
	SeriesStatusDropped   SeriesStatus = "dropped"
	SeriesStatusNone      SeriesStatus = ""
)

func IsSeriesStatus(s string) bool {
	values := []SeriesStatus{
		SeriesStatusOngoing,
		SeriesStatusCompleted,
		SeriesStatusHiatus,
		SeriesStatusCancelled,
		SeriesStatusDropped,
		SeriesStatusNone,
	}

	return slices.Contains(values, SeriesStatus(s))
}

type SortType string

const (
	SortTypePopular      SortType = "popular"
	SortTypeLatest       SortType = "latest"
	SortTypeRating       SortType = "rating"
	SortTypeTitle        SortType = "title"
	SortTypeNewest       SortType = "newest"
	SortTypeLatestUpdate SortType = "latest"
	SortTypeNone         SortType = ""
)

func IsSeriesSortType(s string) bool {
	values := []SortType{
		SortTypePopular,
		SortTypeLatest,
		SortTypeRating,
		SortTypeTitle,
		SortTypeNewest,
		SortTypeLatestUpdate,
		SortTypeNone,
	}

	return slices.Contains(values, SortType(s))
}

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
	SortOrderNone SortOrder = ""
)

func IsSortOrder(s string) bool {
	values := []SortOrder{
		SortOrderAsc,
		SortOrderDesc,
		SortOrderNone,
	}

	return slices.Contains(values, SortOrder(s))
}

type SearchOpts struct {
	Search      string
	Sort        SortType
	SortOrder   SortOrder
	Status      SeriesStatus
	Type        SeriesType
	Artist      string
	Genres      []SeriesGenre
	Offset      int
	Limit       int
	MinChapters int
}

type SearchResult struct {
	Items []SearchResultItem
	Meta  SearchResultMeta
}

type SearchResultMeta struct {
	Total   int
	PerPage int
	HasMore bool
}

type SearchResultItem struct {
	LastChapterAt  time.Time
	UpdatedAt      time.Time
	CreatedAt      time.Time
	PublicURL      string
	SourceURL      string
	Cover          string
	Status         SeriesStatus
	Type           SeriesType
	Author         string
	Artist         string
	Description    string
	Slug           string
	Title          string
	AltTitles      []string
	Genres         []SeriesGenre
	LatestChapters []SearchResultItemChapter
	ChapterCount   int
	ID             int
	Rating         float64
	ReleaseYear    int
}

type SearchResultItemChapter struct {
	EarlyAccessUntil time.Time
	PublishedAt      time.Time
	Title            string
	ID               string
	Number           int
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
	Genres        []SeriesGenre
	ChapterCount  int
	Rating        float64
}

type Chapter struct {
	EarlyAccessUntil time.Time
	PublishedAt      time.Time
	ID               string
	Title            string
	Number           int
}

type GetImageURLsByChapterOpts struct {
	SeriesSlug string
	ChapterID  string
}
