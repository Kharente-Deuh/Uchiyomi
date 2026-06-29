// SPDX-License-Identifier: AGPL-3.0-or-later

// Flat ESM module, consumed via `import * as Catalogue`.

import type { MangaChapterSummaryModel, MangaDetailsModel, MangaSummaryModel } from './manga.domain'
import type { SourceModel } from './source.domain'

export type SourceBrowseType = 'search' | 'popular' | 'latest'

export type TriStateValue = 'IGNORE' | 'INCLUDE' | 'EXCLUDE'

export interface SortState {
  ascending: boolean
  index: number
}

export type SourceFilter
  = | { type: 'checkbox', position: number, name: string, default: boolean }
    | { type: 'tristate', position: number, name: string, default: TriStateValue }
    | { type: 'select', position: number, name: string, default: number, values: string[] }
    | { type: 'text', position: number, name: string, default: string }
    | { type: 'sort', position: number, name: string, default: SortState | null, values: string[] }
    | { type: 'group', position: number, name: string, filters: SourceFilter[] }
    | { type: 'header', position: number, name: string }
    | { type: 'separator', position: number, name: string }

export interface SourceFilterChange {
  position: number
  checkBoxState?: boolean
  triState?: TriStateValue
  selectState?: number
  textState?: string
  sortState?: SortState
  groupChange?: SourceFilterChange
}

export interface SearchMangaParams {
  sourceId: string
  query: string
  page: number
  type: SourceBrowseType
  filters?: SourceFilterChange[]
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
  getSourceFilters: (sourceId: string) => Promise<SourceFilter[]>
}
