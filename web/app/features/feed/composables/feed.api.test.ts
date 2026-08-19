// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createFeedApi } from './feed.api'

const call = vi.fn()

vi.mock('~/utils/api', async importOriginal => ({
  ...(await importOriginal<typeof import('~/utils/api')>()),
  initApi: () => call,
}))

const chapterId = 'ae74a0c7-bd26-47d3-8dc7-42d8287767fd'
const comicId = '210b8853-1c1b-4511-8950-c406ba51fb9d'
const publishedAt = '2026-03-19T01:28:20.000Z'

beforeEach(() => {
  call.mockReset()
})

describe('createFeedApi().getFeed', () => {
  it('omits Go zero earlyAccessUntil so relative time uses publishedAt', async () => {
    call.mockResolvedValue({
      total: 1,
      items: [{
        id: comicId,
        title: 'The Greatest Estate Developer',
        slug: 'the-greatest-estate-developer',
        cover: `/api/comics/${comicId}/cover`,
        source: 'asurascans',
        status: 'completed',
        type: 'manhwa',
        latestChapters: [{
          id: chapterId,
          title: 'Special Chapter [END]',
          number: 223,
          download: 100,
          publishedAt,
          earlyAccessUntil: '0001-01-01T00:00:00Z',
        }],
      }],
    })

    const res = await createFeedApi().getFeed({ offset: 20, limit: 20 })

    expect(call).toHaveBeenCalledWith('/', { params: { offset: 20, limit: 20 } })
    expect(res.success).toBe(true)
    if (!res.success) {
      return
    }

    expect(res.data.items[0]?.latestChapters[0]).toEqual({
      id: chapterId,
      title: 'Special Chapter [END]',
      number: 223,
      download: 100,
      publishedAt: new Date(publishedAt),
    })
  })

  it('keeps a real earlyAccessUntil', async () => {
    const until = '2026-03-26T01:28:20.000Z'
    call.mockResolvedValue({
      total: 1,
      items: [{
        id: comicId,
        title: 'Title',
        slug: 'slug',
        cover: `/api/comics/${comicId}/cover`,
        source: 'asurascans',
        status: 'ongoing',
        type: 'manhwa',
        latestChapters: [{
          id: chapterId,
          number: 1,
          download: 100,
          publishedAt,
          earlyAccessUntil: until,
        }],
      }],
    })

    const res = await createFeedApi().getFeed({ offset: 0, limit: 20 })

    expect(res.success).toBe(true)
    if (!res.success) {
      return
    }

    expect(res.data.items[0]?.latestChapters[0]?.earlyAccessUntil).toEqual(new Date(until))
  })
})
