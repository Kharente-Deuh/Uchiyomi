// SPDX-License-Identifier: AGPL-3.0-or-later
import type { MangaDetails, SearchResult, Source } from './types'

export interface CatalogueRepository {
  listSources: () => Promise<Source[]>
  searchSource: (sourceId: string, query: string, page: number) => Promise<SearchResult>
  getMangaDetails: (mangaId: string) => Promise<MangaDetails>
}
