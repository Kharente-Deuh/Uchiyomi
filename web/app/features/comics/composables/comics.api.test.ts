// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createComicsApi } from './comics.api'

const call = vi.fn()

vi.mock('~/utils/api', async importOriginal => ({
  ...(await importOriginal<typeof import('~/utils/api')>()),
  initApi: () => call,
}))

const comic = { id: 'c1', slug: 'solo-leveling', source: 'asurascans', status: 'ongoing', chapterCount: 12 }

beforeEach(() => {
  call.mockReset()
})

describe('createComicsApi().create', () => {
  it('posts the source and slug', async () => {
    call.mockResolvedValue(comic)

    const res = await createComicsApi().create({ source: 'asurascans', slug: 'solo-leveling' })

    expect(call).toHaveBeenCalledWith('/', { method: 'POST', body: { source: 'asurascans', slug: 'solo-leveling' } })
    expect(res).toEqual({ success: true, data: comic })
  })

  it('surfaces a 409 with its status', async () => {
    call.mockRejectedValue({ statusCode: 409, data: {} })

    const res = await createComicsApi().create({ source: 'asurascans', slug: 'solo-leveling' })

    expect(res.success === false && res.error.status).toBe(409)
  })
})

describe('createComicsApi().deleteById', () => {
  it('deletes the comic by id', async () => {
    call.mockResolvedValue(undefined)

    const res = await createComicsApi().deleteById('c1')

    expect(call).toHaveBeenCalledWith('/c1', { method: 'DELETE' })
    expect(res).toEqual({ success: true, data: undefined })
  })

  it('surfaces a 404 with its status', async () => {
    call.mockRejectedValue({ statusCode: 404, data: {} })

    const res = await createComicsApi().deleteById('missing')

    expect(res.success === false && res.error.status).toBe(404)
  })
})

describe('createComicsApi().search', () => {
  it('requests / with pagination', async () => {
    call.mockResolvedValue({ items: [comic], total: 1 })

    const res = await createComicsApi().search({
      sort: 'title',
      order: 'asc',
      offset: 0,
      limit: 20,
    })

    expect(call).toHaveBeenCalledWith('/', {
      method: 'GET',
      params: { sort: 'title', order: 'asc', offset: 0, limit: 20 },
    })
    expect(res).toEqual({ success: true, data: { items: [comic], total: 1 } })
  })

  it('omits empty optional filters', async () => {
    call.mockResolvedValue({ items: [], total: 0 })

    await createComicsApi().search({
      sort: 'title',
      order: 'asc',
      offset: 0,
      limit: 20,
    })

    expect(call).toHaveBeenCalledWith('/', {
      method: 'GET',
      params: { sort: 'title', order: 'asc', offset: 0, limit: 20 },
    })
  })

  it('includes search, source, status and type when set', async () => {
    call.mockResolvedValue({ items: [], total: 0 })

    await createComicsApi().search({
      search: 'solo',
      source: 'asurascans',
      status: 'ongoing',
      type: 'manhwa',
      sort: 'latest',
      order: 'desc',
      offset: 20,
      limit: 20,
    })

    expect(call).toHaveBeenCalledWith('/', {
      method: 'GET',
      params: {
        search: 'solo',
        source: 'asurascans',
        status: 'ongoing',
        type: 'manhwa',
        sort: 'latest',
        order: 'desc',
        offset: 20,
        limit: 20,
      },
    })
  })

  it('surfaces a failure with its status', async () => {
    call.mockRejectedValue({ statusCode: 502, data: {} })

    const res = await createComicsApi().search({
      sort: 'title',
      order: 'asc',
      offset: 0,
      limit: 20,
    })

    expect(res.success === false && res.error.status).toBe(502)
  })
})
