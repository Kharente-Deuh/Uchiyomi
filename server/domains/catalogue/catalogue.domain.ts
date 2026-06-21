// SPDX-License-Identifier: AGPL-3.0-or-later
// Flat ESM module, consumed via `import * as Catalogue`.
import type * as Manga from './manga.domain'
import type * as Source from './source.domain'

export interface SearchParams {
  sourceId: string
  query: string
  page: number
}

export interface GetMangaDetailsParams {
  mangaId: string
}

export interface SearchResult {
  mangas: Manga.Summary[]
  hasNextPage: boolean
}

// The PORT — the GraphQL/Suwayomi adapter implements it.
export interface Repository {
  listSources: () => Promise<Source.Model[]>
  searchSource: (p: SearchParams) => Promise<SearchResult>
  getMangaDetails: (p: GetMangaDetailsParams) => Promise<Manga.Details>
}
