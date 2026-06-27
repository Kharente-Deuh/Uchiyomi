// SPDX-License-Identifier: AGPL-3.0-or-later
// File-local until the search UI (M4.1c) imports them; only SourceSearchResultDto
// is consumed today (route + presenter), so the sub-types stay unexported to keep
// `knip` clean. Re-export when the front consumes them.
interface SourceSearchLastChapterDto {
  name: string
  uploadedAt: string // ISO 8601
}

interface SourceSearchItemDto {
  id: string // Suwayomi manga id — used to add the series to the library
  title: string
  thumbnailUrl: string | null // BFF proxy path, or null when the series has no cover
  inLibrary: boolean
  chapterCount: number | null // source total; null when enrichment failed
  lastChapter: SourceSearchLastChapterDto | null
}

export interface SourceSearchResultDto {
  hasNextPage: boolean
  items: SourceSearchItemDto[]
}
