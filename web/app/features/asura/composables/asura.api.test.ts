// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createAsuraApi } from './asura.api'

const call = vi.fn()

vi.mock('~/utils/api', async importOriginal => ({
  ...(await importOriginal<typeof import('~/utils/api')>()),
  initApi: () => call,
}))

const item = {
  slug: 'solo-leveling',
  title: 'Solo Leveling',
  cover: '/api/sources/cover/solo-leveling?source=asurascans',
  status: 'ongoing',
  type: 'manhwa',
  chapterCount: 12,
}

beforeEach(() => {
  call.mockReset()
})

describe('createAsuraApi().search', () => {
  it('requests /search with pagination', async () => {
    call.mockResolvedValue({ items: [item], total: 1 })

    const res = await createAsuraApi().search({ offset: 0, limit: 20, sort: 'popular' })

    expect(call).toHaveBeenCalledWith('/search', {
      method: 'GET',
      params: { sort: 'popular', offset: 0, limit: 20 },
    })
    expect(res).toEqual({ success: true, data: { items: [item], total: 1 } })
  })

  it('omits empty optional filters', async () => {
    call.mockResolvedValue({ items: [], total: 0 })

    await createAsuraApi().search({ offset: 0, limit: 20 })

    expect(call).toHaveBeenCalledWith('/search', {
      method: 'GET',
      params: { offset: 0, limit: 20 },
    })
  })

  it('includes status, type and search when set', async () => {
    call.mockResolvedValue({ items: [], total: 0 })

    await createAsuraApi().search({
      search: 'solo',
      status: 'ongoing',
      type: 'manhwa',
      offset: 20,
      limit: 20,
    })

    expect(call).toHaveBeenCalledWith('/search', {
      method: 'GET',
      params: { search: 'solo', status: 'ongoing', type: 'manhwa', offset: 20, limit: 20 },
    })
  })

  it('surfaces a failure with its status', async () => {
    call.mockRejectedValue({ statusCode: 502, data: {} })

    const res = await createAsuraApi().search({ offset: 0, limit: 20 })

    expect(res.success === false && res.error.status).toBe(502)
  })
})
