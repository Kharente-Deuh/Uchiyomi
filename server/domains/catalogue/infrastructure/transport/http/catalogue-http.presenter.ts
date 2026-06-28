// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceSearchResultDto } from '#shared/dto/catalogue/source-search.dto'
import type { SearchSourceWithChaptersResult } from '../../../catalogue.domain'

export function toSourceSearchDto(result: SearchSourceWithChaptersResult): SourceSearchResultDto {
  return {
    hasNextPage: result.hasNextPage,
    items: result.items.map(item => ({
      id: item.id,
      title: item.title,
      // Suwayomi stays internal (ADR-0001): hand the client a BFF proxy path,
      // keyed by manga id, rather than the raw source cover URL.
      thumbnailUrl: item.thumbnailUrl ? `/api/manga/${item.id}/thumbnail` : null,
      inLibrary: item.inLibrary,
      chapterCount: item.chapterCount,
      lastChapter: item.lastChapter
        ? { name: item.lastChapter.name, uploadedAt: new Date(Number(item.lastChapter.uploadDate)).toISOString() }
        : null,
    })),
  }
}
