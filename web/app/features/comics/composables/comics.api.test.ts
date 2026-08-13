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
