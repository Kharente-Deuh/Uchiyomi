// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SuwayomiClient } from '../../../../../utils/suwayomi/client'
import type { CatalogueRepository, GetMangaDetailsByIdParams, SearchMangaParams, SearchMangaResult, SourceFilter } from '../../../catalogue.domain'
import type { MangaChapterSummaryModel, MangaDetailsModel } from '../../../manga.domain'
import type { SourceModel } from '../../../source.domain'
import type { FilterNode } from './graphql-suwayomi-catalogue-repository.mapper'
import { chapterSummaryFromFetched, filterChangeToInput, mangaDetailsToDomain, mangaSummaryToDomain, sourceFiltersToDomain, sourceToDomain } from './graphql-suwayomi-catalogue-repository.mapper'
import { FETCH_CHAPTERS, GET_MANGA_DETAILS, GET_SOURCE_FILTERS, LIST_SOURCES, MANGA_EXISTS, SEARCH_SOURCE } from './graphql-suwayomi-catalogue.operations'

// Domain browse type → Suwayomi FetchSourceMangaType enum value.
const SUWAYOMI_BROWSE_TYPE = { search: 'SEARCH', popular: 'POPULAR', latest: 'LATEST' } as const

export class GraphqlSuwayomiCatalogueRepository implements CatalogueRepository {
  constructor(private readonly client: SuwayomiClient) {}

  async listSources(): Promise<SourceModel[]> {
    const data = await this.client.execute(LIST_SOURCES)

    return data.sources.nodes.map(node => sourceToDomain(node))
  }

  async searchSource(p: SearchMangaParams): Promise<SearchMangaResult> {
    const data = await this.client.execute(SEARCH_SOURCE, {
      sourceId: p.sourceId,
      type: SUWAYOMI_BROWSE_TYPE[p.type],
      // Only SEARCH uses the query; popular/latest browse the source and ignore it.
      query: p.type === 'search' ? p.query : undefined,
      page: p.page,
      // Filters apply to SEARCH only; popular/latest are non-filterable browse modes.
      filters: p.type === 'search' && p.filters?.length ? p.filters.map(change => filterChangeToInput(change)) : undefined,
    })
    // fetchSourceManga is nullable in the SDL (returns null when the source is unavailable).
    if (!data.fetchSourceManga) {
      return { mangas: [], hasNextPage: false }
    }

    return {
      mangas: data.fetchSourceManga.mangas.map(m => mangaSummaryToDomain(m)),
      hasNextPage: data.fetchSourceManga.hasNextPage,
    }
  }

  async getSourceFilters(sourceId: string): Promise<SourceFilter[]> {
    const data = await this.client.execute(GET_SOURCE_FILTERS, { sourceId })

    // The codegen union for GroupFilter.filters is wider than FilterLeafNode[] because
    // the SDL theoretically allows nesting, but the query only selects leaf fields.
    // Cast to the structural mapper type that matches our actual query selection.
    return sourceFiltersToDomain(data.source.filters as FilterNode[])
  }

  async getMangaDetails(p: GetMangaDetailsByIdParams): Promise<MangaDetailsModel> {
    // manga(id: Int!) — convert domain string id to number.
    const data = await this.client.execute(GET_MANGA_DETAILS, { mangaId: Number(p.mangaId) })

    return mangaDetailsToDomain(data.manga)
  }

  async getMangaChapterSummary(mangaId: string): Promise<MangaChapterSummaryModel | undefined> {
    // An unknown id makes Suwayomi raise on the scrape and would surface as a 5xx,
    // indistinguishable from a source being down — so probe existence first
    // (cheap DB read) and let the caller answer 404 on undefined.
    if (!(await this.mangaExists(mangaId))) {
      return
    }

    // manga(id) is Int! — convert the domain string id to a number.
    try {
      const data = await this.client.execute(FETCH_CHAPTERS, { mangaId: Number(mangaId) })

      // The manga exists, so a thrown/empty fetch here means the source is
      // unavailable: a genuine error bubbles up (→ 5xx), null degrades to 0 chapters.
      return chapterSummaryFromFetched(data.fetchChapters?.chapters ?? [])
    } catch (error) {
      if ((error as Error).message.startsWith('No chapters found')) {
        return
      }

      throw error
    }
  }

  // Adapter-internal probe (not on the port): only getMangaChapterSummary needs it.
  private async mangaExists(mangaId: string): Promise<boolean> {
    // mangas(filter) is Int! on id — convert the domain string id to a number.
    const data = await this.client.execute(MANGA_EXISTS, { id: Number(mangaId) })

    return data.mangas.totalCount > 0
  }
}
