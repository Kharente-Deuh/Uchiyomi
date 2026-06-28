// SPDX-License-Identifier: AGPL-3.0-or-later

// File-local until the search UI (M4.1c) imports them; only SourceSearchResultDto
// is consumed today (route + presenter), so the sub-types stay unexported to keep
// `knip` clean. Re-export when the front consumes them.
export interface SourceSearchItemDto {
  id: string // Suwayomi manga id — used to add the series to the library
  title: string
  thumbnailUrl: string | null // BFF proxy path, or null when the series has no cover
  inLibrary: boolean
  sourceUrl: string | null // the original source's manga page URL, null when the source doesn't resolve it
}

export type SourceSearchQueryType = 'search' | 'popular' | 'latest'

export interface SourceSearchQueryDto {
  type: SourceSearchQueryType
  q?: string
  page: number
}

export interface SourceSearchResultDto {
  hasNextPage: boolean
  items: SourceSearchItemDto[]
}
