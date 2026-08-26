// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import * as apiModule from '~/utils/api'
import { createSourceApi } from './sources.api'

vi.mock('~/utils/api', async (importOriginal) => {
  const actual = await importOriginal<typeof apiModule>()

  return {
    ...actual,
    initApi: vi.fn(),
  }
})

describe('createSourceApi', () => {
  const mockApi = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(apiModule.initApi).mockReturnValue(mockApi as never)
  })

  it('calls correct endpoint for search', async () => {
    mockApi.mockResolvedValueOnce({
      items: [{ slug: 'comic-1', updatedAt: '2026-01-01', createdAt: '2026-01-01' }],
      hasNextPage: false,
    })

    const sourceApi = createSourceApi('asurascans')
    const res = await sourceApi.search({ page: 1, search: 'test' })

    expect(apiModule.initApi).toHaveBeenCalledWith('/sources/asurascans')
    expect(mockApi).toHaveBeenCalledWith('/search', expect.objectContaining({
      method: 'GET',
      params: expect.objectContaining({ search: 'test', page: 1 }),
    }))
    expect(res.success).toBe(true)
    if (res.success) {
      expect(res.data.items).toHaveLength(1)
      expect(res.data.items[0]!.updatedAt).toBeInstanceOf(Date)
    }
  })

  it('calls correct endpoint for series infos', async () => {
    mockApi.mockResolvedValueOnce({
      slug: 'test-comic',
      title: 'Test',
      updatedAt: '2026-01-01',
      createdAt: '2026-01-01',
    })

    const sourceApi = createSourceApi('kingofshojo')
    const res = await sourceApi.getInfosBySlug('test-comic')

    expect(apiModule.initApi).toHaveBeenCalledWith('/sources/kingofshojo')
    expect(mockApi).toHaveBeenCalledWith('/series/test-comic', { method: 'GET' })
    expect(res.success).toBe(true)
  })

  it('calls correct endpoint for chapters', async () => {
    mockApi.mockResolvedValueOnce([
      { id: '1', title: 'Chapter 1', number: 1, publishedAt: '2026-01-01' },
    ])

    const sourceApi = createSourceApi('asurascans')
    const res = await sourceApi.getSeriesChapters('test-comic')

    expect(mockApi).toHaveBeenCalledWith('/series/test-comic/chapters', { method: 'GET' })
    expect(res.success).toBe(true)
    if (res.success) {
      expect(res.data).toHaveLength(1)
      expect(res.data[0]!.publishedAt).toBeInstanceOf(Date)
    }
  })
})
