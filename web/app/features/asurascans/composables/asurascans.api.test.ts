// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ASURA_SOURCE_NAME } from '~/constants'
import { createAsuraScansApi } from './asurascans.api'

const mocks = vi.hoisted(() => {
  const call = vi.fn()
  const initApi = vi.fn(() => call)

  return { call, initApi }
})

vi.mock('~/utils/api', async importOriginal => ({
  ...(await importOriginal<typeof import('~/utils/api')>()),
  initApi: mocks.initApi,
}))

const rawItem = {
  slug: 'solo-leveling',
  title: 'Solo Leveling',
  cover: '/api/sources/cover/solo-leveling?source=asurascans',
  status: 'ongoing',
  type: 'manhwa',
  chapterCount: 12,
  lastChapterAt: '2026-01-01T00:00:00.000Z',
  updatedAt: '2026-01-01T00:00:00.000Z',
  createdAt: '2026-01-01T00:00:00.000Z',
  publicUrl: '',
  sourceUrl: '',
  author: '',
  artist: '',
  description: '',
  altTitles: [] as string[],
  genres: [] as string[],
  latestChapters: [] as [],
  rating: 0,
  releaseYear: 2020,
}

const item = {
  ...rawItem,
  lastChapterAt: new Date(rawItem.lastChapterAt),
  updatedAt: new Date(rawItem.updatedAt),
  createdAt: new Date(rawItem.createdAt),
  latestChapters: [],
}

beforeEach(() => {
  mocks.call.mockReset()
  mocks.initApi.mockClear()
})

describe('createAsuraScansApi', () => {
  it('uses ASURA_SOURCE_NAME as the API prefix', () => {
    createAsuraScansApi()
    expect(mocks.initApi).toHaveBeenCalledWith(`/sources/${ASURA_SOURCE_NAME}`)
  })
})

describe('createAsuraScansApi().search', () => {
  it('requests /search with page', async () => {
    mocks.call.mockResolvedValue({ items: [rawItem], hasNextPage: true })

    const res = await createAsuraScansApi().search({ page: 1, sort: 'popular' })

    expect(mocks.call).toHaveBeenCalledWith('/search', {
      method: 'GET',
      params: { sort: 'popular', page: 1 },
    })
    expect(res).toEqual({ success: true, data: { items: [item], hasNextPage: true } })
  })

  it('omits empty optional filters', async () => {
    mocks.call.mockResolvedValue({ items: [], hasNextPage: false })

    await createAsuraScansApi().search({ page: 1 })

    expect(mocks.call).toHaveBeenCalledWith('/search', {
      method: 'GET',
      params: { page: 1 },
    })
  })

  it('includes status, type and search when set', async () => {
    mocks.call.mockResolvedValue({ items: [], hasNextPage: false })

    await createAsuraScansApi().search({
      search: 'solo',
      status: 'ongoing',
      type: 'manhwa',
      page: 2,
    })

    expect(mocks.call).toHaveBeenCalledWith('/search', {
      method: 'GET',
      params: { search: 'solo', status: 'ongoing', type: 'manhwa', page: 2 },
    })
  })

  it('includes minChapters when greater than 0', async () => {
    mocks.call.mockResolvedValue({ items: [], hasNextPage: false })

    await createAsuraScansApi().search({ page: 1, minChapters: 5 })

    expect(mocks.call).toHaveBeenCalledWith('/search', {
      method: 'GET',
      params: { page: 1, minChapters: 5 },
    })
  })

  it('surfaces a failure with its status', async () => {
    mocks.call.mockRejectedValue({ statusCode: 502, data: {} })

    const res = await createAsuraScansApi().search({ page: 1 })

    expect(res.success === false && res.error.status).toBe(502)
  })
})
