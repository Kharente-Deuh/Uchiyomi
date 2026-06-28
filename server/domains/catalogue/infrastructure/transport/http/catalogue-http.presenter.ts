// SPDX-License-Identifier: AGPL-3.0-or-later

import type { MangaChapterSummaryDto } from '#shared/dto/catalogue/manga-chapter-summary.dto'
import type { SourceFilterDto } from '#shared/dto/catalogue/source-filters.dto'
import type { SourceSearchResultDto } from '#shared/dto/catalogue/source-search.dto'
import type { SearchMangaResult, SourceFilter } from '../../../catalogue.domain'
import type { MangaChapterSummaryModel } from '../../../manga.domain'

export function toSourceSearchDto(result: SearchMangaResult): SourceSearchResultDto {
  return {
    hasNextPage: result.hasNextPage,
    items: result.mangas.map(item => ({
      id: item.id,
      title: item.title,
      thumbnailUrl: item.thumbnailUrl ? `/api/manga/${item.id}/thumbnail` : null,
      inLibrary: item.inLibrary,
      sourceUrl: item.realUrl ?? null,
    })),
  }
}

function toFilterDto(filter: SourceFilter): SourceFilterDto {
  if (filter.type === 'group') {
    return { type: 'group', position: filter.position, name: filter.name, filters: filter.filters.map(child => toFilterDto(child)) }
  }

  // Leaf variants are structurally identical between domain and DTO.
  return filter
}

export function toSourceFiltersDto(filters: SourceFilter[]): SourceFilterDto[] {
  return filters.map(filter => toFilterDto(filter))
}

export function toMangaChapterSummaryDto(summary: MangaChapterSummaryModel): MangaChapterSummaryDto {
  return {
    chapterCount: summary.chapterCount,
    lastChapter: summary.lastChapter
      ? {
          name: summary.lastChapter.name,
          // The domain carries Suwayomi's raw epoch-ms string; the API contract is ISO 8601.
          uploadedAt: new Date(Number(summary.lastChapter.uploadedAt)).toISOString(),
        }
      : null,
  }
}
