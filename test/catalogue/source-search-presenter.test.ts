// SPDX-License-Identifier: AGPL-3.0-or-later

// @vitest-environment node
import { describe, expect, it } from 'vitest'
import { toMangaChapterSummaryDto, toSourceSearchDto } from '../../server/domains/catalogue/infrastructure/transport/http/catalogue-http.presenter'
import { MangaChapterSummaryModel } from '../../server/domains/catalogue/manga.domain'

describe('toSourceSearchDto', () => {
  it('maps a cover to the BFF proxy path', () => {
    const dto = toSourceSearchDto({
      hasNextPage: true,
      mangas: [{ id: '42', title: 'A', thumbnailUrl: 'https://source/cover.jpg', inLibrary: false }],
    })

    expect(dto.hasNextPage).toBe(true)
    expect(dto.items[0]).toEqual({
      id: '42',
      title: 'A',
      thumbnailUrl: '/api/manga/42/thumbnail',
      inLibrary: false,
    })
  })

  it('emits null thumbnailUrl when there is no cover', () => {
    const dto = toSourceSearchDto({
      hasNextPage: false,
      mangas: [{ id: '7', title: 'B', thumbnailUrl: undefined, inLibrary: true }],
    })

    expect(dto.items[0]!.thumbnailUrl).toBeNull()
  })
})

describe('toMangaChapterSummaryDto', () => {
  it('converts the last chapter epoch ms to ISO 8601', () => {
    const dto = toMangaChapterSummaryDto(new MangaChapterSummaryModel({
      chapterCount: 12,
      lastChapter: { name: 'Chapter 12', uploadedAt: '1717200000000' },
    }))

    expect(dto).toEqual({
      chapterCount: 12,
      lastChapter: { name: 'Chapter 12', uploadedAt: '2024-06-01T00:00:00.000Z' },
    })
  })

  it('passes through a null last chapter', () => {
    const dto = toMangaChapterSummaryDto(new MangaChapterSummaryModel({ chapterCount: 0, lastChapter: null }))

    expect(dto).toEqual({ chapterCount: 0, lastChapter: null })
  })
})
