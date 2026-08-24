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

describe('createComicsApi().getById', () => {
  it('gets the comic by id', async () => {
    const full = { ...comic, title: 'Solo Leveling', author: 'Chugong', artist: 'Jang', type: 'manhwa', description: '', cover: '/cover', genres: [], altTitles: [] }
    call.mockResolvedValue(full)

    const res = await createComicsApi().getById('c1')

    expect(call).toHaveBeenCalledWith('/c1', { method: 'GET' })
    expect(res).toEqual({ success: true, data: full })
  })

  it('surfaces a 404 with its status', async () => {
    call.mockRejectedValue({ statusCode: 404, data: {} })

    const res = await createComicsApi().getById('missing')

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
      sort: 'addedAt',
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
        sort: 'addedAt',
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

describe('createComicsApi().refreshById', () => {
  it('posts to /:id/refresh and returns the refreshed comic', async () => {
    call.mockResolvedValue(comic)

    const res = await createComicsApi().refreshById('c1')

    expect(call).toHaveBeenCalledWith('/c1/refresh', { method: 'POST' })
    expect(res).toEqual({ success: true, data: comic })
  })

  it('surfaces a failure with its status', async () => {
    call.mockRejectedValue({ statusCode: 500, data: {} })

    const res = await createComicsApi().refreshById('c1')

    expect(res.success === false && res.error.status).toBe(500)
  })
})

describe('createComicsApi().getProgress', () => {
  it('gets the comic progress by id', async () => {
    const progressData = { continue: { chapterId: 'ch-1', page: 5 } }
    call.mockResolvedValue(progressData)

    const res = await createComicsApi().getProgress('c1')

    expect(call).toHaveBeenCalledWith('/c1/progress', { method: 'GET' })
    expect(res).toEqual({ success: true, data: { continue: progressData.continue } })
  })

  it('surfaces a failure with its status', async () => {
    call.mockRejectedValue({ statusCode: 404, data: {} })

    const res = await createComicsApi().getProgress('c1')

    expect(res.success === false && res.error.status).toBe(404)
  })
})

describe('createComicsApi().setChaptersProgress', () => {
  it('posts chapter progress updates to /:comicId/progress', async () => {
    call.mockResolvedValue(undefined)

    const res = await createComicsApi().setChaptersProgress({
      comicId: 'c1',
      chapterIds: ['ch-1', 'ch-2'],
      read: true,
    })

    expect(call).toHaveBeenCalledWith('/c1/progress', {
      method: 'POST',
      body: { chapterIds: ['ch-1', 'ch-2'], read: true },
    })
    expect(res).toEqual({ success: true, data: undefined })
  })

  it('surfaces a failure with its status', async () => {
    call.mockRejectedValue({ statusCode: 400, data: {} })

    const res = await createComicsApi().setChaptersProgress({
      comicId: 'c1',
      chapterIds: ['ch-1'],
      read: false,
    })

    expect(res.success === false && res.error.status).toBe(400)
  })
})

describe('createComicsApi().retryChaptersDownload', () => {
  it('posts retry request to /:comicId/retry', async () => {
    call.mockResolvedValue(undefined)

    const res = await createComicsApi().retryChaptersDownload('c1', ['ch-1', 'ch-2'])

    expect(call).toHaveBeenCalledWith('/c1/retry', {
      method: 'POST',
      body: { chapterIds: ['ch-1', 'ch-2'] },
    })
    expect(res).toEqual({ success: true, data: undefined })
  })

  it('surfaces a failure with its status', async () => {
    call.mockRejectedValue({ statusCode: 500, data: {} })

    const res = await createComicsApi().retryChaptersDownload('c1', ['ch-1'])

    expect(res.success === false && res.error.status).toBe(500)
  })
})
