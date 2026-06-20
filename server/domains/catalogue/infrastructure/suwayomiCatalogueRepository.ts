// SPDX-License-Identifier: AGPL-3.0-or-later
import type { SuwayomiClient } from '../../../utils/suwayomi/client'
import type { CatalogueRepository } from '../domain/repository'
import { mapMangaDetails, mapMangaSummary, mapSource } from './mappers'
import { GET_MANGA_DETAILS, LIST_SOURCES, SEARCH_SOURCE } from './operations'

export function createSuwayomiCatalogueRepository(client: SuwayomiClient): CatalogueRepository {
  return {
    async listSources() {
      const data = await client.execute(LIST_SOURCES)

      return data.sources.nodes.map(node => mapSource(node))
    },

    async searchSource(sourceId, query, page) {
      const data = await client.execute(SEARCH_SOURCE, { sourceId, query, page })
      // fetchSourceManga is nullable in the SDL (returns null when the source is unavailable).
      if (!data.fetchSourceManga) {
        return { mangas: [], hasNextPage: false }
      }

      return {
        mangas: data.fetchSourceManga.mangas.map(m => mapMangaSummary(m)),
        hasNextPage: data.fetchSourceManga.hasNextPage,
      }
    },

    async getMangaDetails(mangaId) {
      // manga(id: Int!) — convert domain string id to number.
      const data = await client.execute(GET_MANGA_DETAILS, { mangaId: Number(mangaId) })

      return mapMangaDetails(data.manga)
    },
  }
}
