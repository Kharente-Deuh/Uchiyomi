// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createChaptersApi } from './chapters.api'

const call = vi.fn()

vi.mock('~/utils/api', async importOriginal => ({
  ...(await importOriginal<typeof import('~/utils/api')>()),
  initApi: () => call,
}))

beforeEach(() => {
  call.mockReset()
})

describe('createChaptersApi().getByComicId', () => {
  it('lists chapters for the comic and deserializes dates', async () => {
    const publishedAt = '2026-01-01T00:00:00.000Z'
    const chapters = [{
      id: 'ch-1',
      comicId: 'c1',
      number: 1,
      publishedAt,
      earlyAccessUntil: '0001-01-01T00:00:00Z',
    }]
    call.mockResolvedValue(chapters)

    const res = await createChaptersApi().getByComicId('c1')

    expect(call).toHaveBeenCalledWith('/', { params: { comicId: 'c1' } })
    expect(res).toEqual({
      success: true,
      data: [{
        id: 'ch-1',
        comicId: 'c1',
        number: 1,
        publishedAt: new Date(publishedAt),
      }],
    })
  })

  it('keeps valid earlyAccessUntil as Date', async () => {
    const publishedAt = '2026-01-01T00:00:00.000Z'
    const earlyAccessUntil = '2026-02-01T00:00:00.000Z'
    const chapters = [{
      id: 'ch-1',
      comicId: 'c1',
      number: 1,
      publishedAt,
      earlyAccessUntil,
    }]
    call.mockResolvedValue(chapters)

    const res = await createChaptersApi().getByComicId('c1')

    expect(res).toEqual({
      success: true,
      data: [{
        id: 'ch-1',
        comicId: 'c1',
        number: 1,
        publishedAt: new Date(publishedAt),
        earlyAccessUntil: new Date(earlyAccessUntil),
      }],
    })
  })

  it('surfaces a 404 with its status', async () => {
    call.mockRejectedValue({ statusCode: 404, data: {} })

    const res = await createChaptersApi().getByComicId('missing')

    expect(res.success === false && res.error.status).toBe(404)
  })
})

describe('createChaptersApi().getByIds', () => {
  it('posts the chapter ids', async () => {
    const publishedAt = '2026-01-01T00:00:00.000Z'
    const chapters = [{ id: 'ch-1', comicId: 'c1', number: 1, publishedAt }]
    call.mockResolvedValue(chapters)

    const res = await createChaptersApi().getByIds(['ch-1', 'ch-2'])

    expect(call).toHaveBeenCalledWith('/list', { method: 'POST', body: { ids: ['ch-1', 'ch-2'] } })
    expect(res).toEqual({
      success: true,
      data: [{
        id: 'ch-1',
        comicId: 'c1',
        number: 1,
        publishedAt: new Date(publishedAt),
      }],
    })
  })

  it('surfaces a failure with its status', async () => {
    call.mockRejectedValue({ statusCode: 502, data: {} })

    const res = await createChaptersApi().getByIds(['ch-1'])

    expect(res.success === false && res.error.status).toBe(502)
  })
})

describe('createChaptersApi().getById', () => {
  it('returns the chapter with next and previous neighbors', async () => {
    const publishedAt = '2026-01-01T00:00:00.000Z'
    const next = { id: 'ch-3', title: 'Three', number: 3 }
    const previous = { id: 'ch-1', title: 'One', number: 1 }
    call.mockResolvedValue({
      id: 'ch-2',
      comicId: 'c1',
      number: 2,
      title: 'Two',
      publishedAt,
      pageUrls: ['/api/chapters/ch-2/pages/1'],
      next,
      previous,
    })

    const res = await createChaptersApi().getById('ch-2')

    expect(call).toHaveBeenCalledWith('/ch-2')
    expect(res).toEqual({
      success: true,
      data: {
        id: 'ch-2',
        comicId: 'c1',
        number: 2,
        title: 'Two',
        publishedAt: new Date(publishedAt),
        pageUrls: ['/api/chapters/ch-2/pages/1'],
        next,
        previous,
      },
    })
  })

  it('surfaces a 404 with its status', async () => {
    call.mockRejectedValue({ statusCode: 404, data: {} })

    const res = await createChaptersApi().getById('missing')

    expect(res.success === false && res.error.status).toBe(404)
  })
})

describe('createChaptersApi().saveProgress', () => {
  it('puts the page on the chapter progress route', async () => {
    const progress = { page: 4, updatedAt: '2026-08-22T00:00:00.000Z' }
    call.mockResolvedValue(progress)

    const res = await createChaptersApi().saveProgress({ id: 'ch-2', page: 4 })

    expect(call).toHaveBeenCalledWith('/ch-2/progress', { method: 'PUT', body: { page: 4 } })
    expect(res).toEqual({ success: true, data: progress })
  })

  it('surfaces a failure with its status', async () => {
    call.mockRejectedValue({ statusCode: 500, data: {} })

    const res = await createChaptersApi().saveProgress({ id: 'ch-2', page: 4 })

    expect(res.success === false && res.error.status).toBe(500)
  })
})

describe('createChaptersApi().retryDownload', () => {
  it('posts retry for the chapter id', async () => {
    call.mockResolvedValue(undefined)

    const res = await createChaptersApi().retryDownload('ch-1')

    expect(call).toHaveBeenCalledWith('/ch-1/retry', { method: 'POST' })
    expect(res).toEqual({ success: true, data: undefined })
  })

  it('surfaces a 409 with its status', async () => {
    call.mockRejectedValue({ statusCode: 409, data: {} })

    const res = await createChaptersApi().retryDownload('ch-1')

    expect(res.success === false && res.error.status).toBe(409)
  })
})

describe('createChaptersApi().deleteProgress', () => {
  it('sends DELETE to /:id/progress', async () => {
    call.mockResolvedValue(undefined)

    const res = await createChaptersApi().deleteProgress('ch-1')

    expect(call).toHaveBeenCalledWith('/ch-1/progress', { method: 'DELETE' })
    expect(res).toEqual({ success: true, data: undefined })
  })

  it('surfaces a failure with its status', async () => {
    call.mockRejectedValue({ statusCode: 500, data: {} })

    const res = await createChaptersApi().deleteProgress('ch-1')

    expect(res.success === false && res.error.status).toBe(500)
  })
})
