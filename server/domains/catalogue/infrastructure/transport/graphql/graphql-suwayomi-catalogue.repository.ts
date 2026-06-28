// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SuwayomiClient } from '../../../../../utils/suwayomi/client'
import type { CatalogueRepository, GetMangaDetailsByIdParams, SearchMangaParams, SearchMangaResult, SourceMangaChapterSummary } from '../../../catalogue.domain'
import type { MangaDetailsModel } from '../../../manga.domain'
import type { SourceModel } from '../../../source.domain'
import { chapterSummaryFromFetched, mangaDetailsToDomain, mangaSummaryToDomain, sourceToDomain } from './graphql-suwayomi-catalogue-repository.mapper'
import { FETCH_CHAPTERS, GET_MANGA_DETAILS, LIST_SOURCES, SEARCH_SOURCE } from './graphql-suwayomi-catalogue.operations'

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

  async getMangaDetails(p: GetMangaDetailsByIdParams): Promise<MangaDetailsModel> {
    // manga(id: Int!) — convert domain string id to number.
    const data = await this.client.execute(GET_MANGA_DETAILS, { mangaId: Number(p.mangaId) })

    return mangaDetailsToDomain(data.manga)
  }

  async getSourceMangaChapterSummary(mangaId: string): Promise<SourceMangaChapterSummary> {
    // manga(id) is Int! — convert the domain string id to a number.
    const data = await this.client.execute(FETCH_CHAPTERS, { mangaId: Number(mangaId) })
    // fetchChapters is nullable in the SDL (null when the source is unavailable).
    const chapters = data.fetchChapters?.chapters ?? []

    return chapterSummaryFromFetched(chapters)
  }
}
