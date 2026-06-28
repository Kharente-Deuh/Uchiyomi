// SPDX-License-Identifier: AGPL-3.0-or-later

import type { MangaChapterSummaryDto } from '#shared/dto/catalogue/manga-chapter-summary.dto'
import type { SourceSearchResultDto } from '#shared/dto/catalogue/source-search.dto'
import type { SearchMangaResult } from '../../../catalogue.domain'
import type { MangaChapterSummaryModel } from '../../../manga.domain'

export function toSourceSearchDto(result: SearchMangaResult): SourceSearchResultDto {
  return {
    hasNextPage: result.hasNextPage,
    items: result.mangas.map(item => ({
      id: item.id,
      title: item.title,
      thumbnailUrl: item.thumbnailUrl ? `/api/manga/${item.id}/thumbnail` : null,
      inLibrary: item.inLibrary,
    })),
  }
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
