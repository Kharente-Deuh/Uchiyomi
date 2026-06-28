// SPDX-License-Identifier: AGPL-3.0-or-later

// Flat ESM module, consumed via `import * as Catalogue`.

import type { MangaChapterSummaryModel, MangaDetailsModel, MangaSummaryModel } from './manga.domain'
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

// The PORT — the GraphQL/Suwayomi adapter implements it.
export interface CatalogueRepository {
  listSources: () => Promise<SourceModel[]>
  searchSource: (p: SearchMangaParams) => Promise<SearchMangaResult>
  getMangaDetails: (p: GetMangaDetailsByIdParams) => Promise<MangaDetailsModel>
  getMangaChapterSummary: (mangaId: string) => Promise<MangaChapterSummaryModel | undefined>
}
