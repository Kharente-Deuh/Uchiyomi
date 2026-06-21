// SPDX-License-Identifier: AGPL-3.0-or-later
import type { SuwayomiClient } from '../../../../../utils/suwayomi/client'
import type * as Catalogue from '../../../catalogue.domain'
import type * as Manga from '../../../manga.domain'
import type * as Source from '../../../source.domain'
import { mangaDetailsToDomain, mangaSummaryToDomain, sourceToDomain } from './graphql-suwayomi-catalogue-repository.mapper'
import { GET_MANGA_DETAILS, LIST_SOURCES, SEARCH_SOURCE } from './graphql-suwayomi-catalogue.operations'

export class GraphqlSuwayomiCatalogueRepository implements Catalogue.Repository {
  constructor(private readonly client: SuwayomiClient) {}

  async listSources(): Promise<Source.Model[]> {
    const data = await this.client.execute(LIST_SOURCES)

    return data.sources.nodes.map(node => sourceToDomain(node))
  }

  async searchSource(p: Catalogue.SearchParams): Promise<Catalogue.SearchResult> {
    const data = await this.client.execute(SEARCH_SOURCE, { sourceId: p.sourceId, query: p.query, page: p.page })
    // fetchSourceManga is nullable in the SDL (returns null when the source is unavailable).
    if (!data.fetchSourceManga) {
      return { mangas: [], hasNextPage: false }
    }

    return {
      mangas: data.fetchSourceManga.mangas.map(m => mangaSummaryToDomain(m)),
      hasNextPage: data.fetchSourceManga.hasNextPage,
    }
  }

  async getMangaDetails(p: Catalogue.GetMangaDetailsParams): Promise<Manga.Details> {
    // manga(id: Int!) — convert domain string id to number.
    const data = await this.client.execute(GET_MANGA_DETAILS, { mangaId: Number(p.mangaId) })

    return mangaDetailsToDomain(data.manga)
  }
}
