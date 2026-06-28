// SPDX-License-Identifier: AGPL-3.0-or-later

// @vitest-environment node
import { describe, expect, it } from 'vitest'
import { toSourceSearchDto } from '../../server/domains/catalogue/infrastructure/transport/http/catalogue-http.presenter'

describe('toSourceSearchDto', () => {
  it('maps a cover to the proxy path and converts epoch ms to ISO', () => {
    const dto = toSourceSearchDto({
      hasNextPage: true,
      items: [{
        id: '42',
        title: 'A',
        thumbnailUrl: 'https://source/cover.jpg',
        inLibrary: false,
        chapterCount: 12,
        lastChapter: { name: 'Chapter 12', uploadDate: '1717200000000' },
      }],
    })

    expect(dto.hasNextPage).toBe(true)
    expect(dto.items[0]).toEqual({
      id: '42',
      title: 'A',
      thumbnailUrl: '/api/manga/42/thumbnail',
      inLibrary: false,
      chapterCount: 12,
      lastChapter: { name: 'Chapter 12', uploadedAt: '2024-06-01T00:00:00.000Z' },
    })
  })

  it('emits null thumbnailUrl when there is no cover, and null lastChapter', () => {
    const dto = toSourceSearchDto({
      hasNextPage: false,
      items: [{ id: '7', title: 'B', thumbnailUrl: undefined, inLibrary: true, chapterCount: null, lastChapter: null }],
    })

    expect(dto.items[0]!.thumbnailUrl).toBeNull()
    expect(dto.items[0]!.lastChapter).toBeNull()
    expect(dto.items[0]!.chapterCount).toBeNull()
  })
})
