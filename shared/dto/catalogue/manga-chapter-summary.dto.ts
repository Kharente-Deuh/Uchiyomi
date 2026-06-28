// SPDX-License-Identifier: AGPL-3.0-or-later

// File-local until the front consumes it; only MangaChapterSummaryDto is used today.
interface MangaChapterSummaryItemDto {
  name: string
  uploadedAt: string
}

export interface MangaChapterSummaryDto {
  chapterCount: number
  lastChapter: MangaChapterSummaryItemDto | null
}
