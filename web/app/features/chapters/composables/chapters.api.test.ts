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
  it('lists chapters for the comic', async () => {
    const chapters = [{ id: 'ch-1', comicId: 'c1', number: 1 }]
    call.mockResolvedValue(chapters)

    const res = await createChaptersApi().getByComicId('c1')

    expect(call).toHaveBeenCalledWith('/', { params: { comicId: 'c1' } })
    expect(res).toEqual({ success: true, data: chapters })
  })

  it('surfaces a 404 with its status', async () => {
    call.mockRejectedValue({ statusCode: 404, data: {} })

    const res = await createChaptersApi().getByComicId('missing')

    expect(res.success === false && res.error.status).toBe(404)
  })
})

describe('createChaptersApi().getByIds', () => {
  it('posts the chapter ids', async () => {
    const chapters = [{ id: 'ch-1', comicId: 'c1', number: 1 }]
    call.mockResolvedValue(chapters)

    const res = await createChaptersApi().getByIds(['ch-1', 'ch-2'])

    expect(call).toHaveBeenCalledWith('/list', { method: 'POST', body: { ids: ['ch-1', 'ch-2'] } })
    expect(res).toEqual({ success: true, data: chapters })
  })

  it('surfaces a failure with its status', async () => {
    call.mockRejectedValue({ statusCode: 502, data: {} })

    const res = await createChaptersApi().getByIds(['ch-1'])

    expect(res.success === false && res.error.status).toBe(502)
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
