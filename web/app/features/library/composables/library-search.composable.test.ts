// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { LightComic } from '~/features/comics/types'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useToast } from '~/composables/toast.composable'
import { useLibraryStore } from '../stores/library-search.store'
import { useLibrarySearch } from './library-search.composable'

const { search, smAndDown } = vi.hoisted(() => ({
  search: vi.fn(),
  smAndDown: { value: false },
}))

vi.mock('~/features/comics/composables/comics.api', () => ({
  createComicsApi: () => ({ search }),
}))

function i18nStub(): { t: (key: string) => string } {
  return { t: (key: string) => key }
}

function displayStub(): { smAndDown: { value: boolean } } {
  return { smAndDown }
}

function debounceStub(fn: () => unknown): () => unknown {
  return fn
}

mockNuxtImport('useI18n', () => i18nStub)
mockNuxtImport('useDisplay', () => displayStub)
mockNuxtImport('useDebounceFn', debounceStub)

function comic(id = 'c1'): LightComic {
  return {
    id,
    title: 'Solo Leveling',
    slug: 'solo-leveling',
    source: 'asurascans',
    status: 'ongoing',
    chapterCount: 12,
    cover: '/cover',
  }
}

function apiError(status: number): { success: false, error: { status: number, message: string } } {
  return { success: false, error: { status, message: 'boom' } }
}

beforeEach(() => {
  setActivePinia(createPinia())
  useToast().messages.value.length = 0
  search.mockReset()
  smAndDown.value = false
})

describe('useLibrarySearch', () => {
  it('does not search when doSearch is false', async () => {
    useLibrarySearch({ doSearch: false })
    await Promise.resolve()

    expect(search).not.toHaveBeenCalled()
  })

  it('searches with default title sort and first-page offset', async () => {
    search.mockResolvedValue({ success: true, data: { items: [comic()], total: 1 } })

    const library = useLibrarySearch({ doSearch: true })
    await vi.waitFor(() => expect(search).toHaveBeenCalled())

    expect(search).toHaveBeenCalledWith({
      sort: 'title',
      order: 'asc',
      offset: 0,
      limit: 20,
    })
    expect(library.comics.value).toEqual([comic()])
    expect(library.maxPage.value).toBe(1)
  })

  it('uses descending order for latest sort', async () => {
    search.mockResolvedValue({ success: true, data: { items: [], total: 0 } })
    const library = useLibrarySearch({ doSearch: true })
    await vi.waitFor(() => expect(search).toHaveBeenCalled())
    search.mockClear()

    library.sort.value = 'addedAt'
    await vi.waitFor(() => expect(search).toHaveBeenCalled())

    expect(search).toHaveBeenCalledWith({
      sort: 'addedAt',
      order: 'desc',
      offset: 0,
      limit: 20,
    })
  })

  it('forwards search, status, type and source filters', async () => {
    search.mockResolvedValue({ success: true, data: { items: [], total: 0 } })
    const library = useLibrarySearch({ doSearch: true })
    await vi.waitFor(() => expect(search).toHaveBeenCalled())
    search.mockClear()

    library.search.value = 'solo'
    library.status.value = 'ongoing'
    library.type.value = 'manhwa'
    library.source.value = 'asurascans'
    await vi.waitFor(() => expect(search).toHaveBeenCalled())

    expect(search).toHaveBeenLastCalledWith({
      search: 'solo',
      status: 'ongoing',
      type: 'manhwa',
      source: 'asurascans',
      sort: 'title',
      order: 'asc',
      offset: 0,
      limit: 20,
    })
  })

  it('converts page 2 to a 20-item offset', async () => {
    search.mockResolvedValue({ success: true, data: { items: [], total: 40 } })
    const library = useLibrarySearch({ doSearch: true })
    await vi.waitFor(() => expect(search).toHaveBeenCalled())
    search.mockClear()

    library.page.value = 2
    await vi.waitFor(() => expect(search).toHaveBeenCalled())

    expect(search).toHaveBeenCalledWith({
      sort: 'title',
      order: 'asc',
      offset: 20,
      limit: 20,
    })
    expect(library.maxPage.value).toBe(2)
  })

  it('accumulates pages on small screens', async () => {
    smAndDown.value = true
    search
      .mockResolvedValueOnce({ success: true, data: { items: [comic('c1')], total: 40 } })
      .mockResolvedValueOnce({ success: true, data: { items: [comic('c2')], total: 40 } })

    const library = useLibrarySearch({ doSearch: true })
    await vi.waitFor(() => expect(library.comics.value).toEqual([comic('c1')]))

    library.page.value = 2
    await vi.waitFor(() => expect(library.comics.value.map(c => c.id)).toEqual(['c1', 'c2']))
  })

  it('toasts and clears results when search fails', async () => {
    search.mockResolvedValue(apiError(500))

    const library = useLibrarySearch({ doSearch: true })
    await vi.waitFor(() => expect(useToast().messages.value.length).toBe(1))

    expect(useToast().messages.value).toEqual([{ text: 'sources.asurascans.search.error', color: 'error' }])
    expect(library.comics.value).toEqual([])
    expect(library.maxPage.value).toBe(0)
  })

  it('resetFilters restores the store', async () => {
    search.mockResolvedValue({ success: true, data: { items: [], total: 0 } })
    const library = useLibrarySearch({ doSearch: false })
    library.search.value = 'solo'
    library.sort.value = 'addedAt'

    library.resetFilters()

    expect(useLibraryStore().search).toBeUndefined()
    expect(useLibraryStore().sort).toBe('title')
  })
})
