// SPDX-License-Identifier: AGPL-3.0-or-later

// Flat ESM module, consumed via `import * as Catalogue`.

import type { MangaDetailsModel, MangaSummaryModel } from './manga.domain'
import type { SourceModel } from './source.domain'

export type SourceBrowseType = 'search' | 'popular' | 'latest'

export interface SearchMangaParams {
  sourceId: string
  query: string
  page: number
  type: SourceBrowseType
}

export interface GetMangaDetailsByIdParams {
  mangaId: string
}

export interface SearchMangaResult {
  mangas: MangaSummaryModel[]
  hasNextPage: boolean
}

export interface SourceMangaChapterSummary {
  // Total chapters the source reports. 0 when the source returns none.
  chapterCount: number
  // The most recently uploaded chapter, or null when there are none.
  lastChapter: { name: string, uploadDate: string } | null
}

export interface SearchedSourceMangaItem {
  id: string
  title: string
  thumbnailUrl?: string
  inLibrary: boolean
  // null when per-item enrichment failed (the page is never aborted for it).
  chapterCount: number | null
  lastChapter: { name: string, uploadDate: string } | null
}

export interface SearchSourceWithChaptersResult {
  items: SearchedSourceMangaItem[]
  hasNextPage: boolean
}

// The PORT — the GraphQL/Suwayomi adapter implements it.
export interface CatalogueRepository {
  listSources: () => Promise<SourceModel[]>
  searchSource: (p: SearchMangaParams) => Promise<SearchMangaResult>
  getMangaDetails: (p: GetMangaDetailsByIdParams) => Promise<MangaDetailsModel>
  getSourceMangaChapterSummary: (mangaId: string) => Promise<SourceMangaChapterSummary>
}
