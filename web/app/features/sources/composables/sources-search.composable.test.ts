// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { SourceSearchItem } from '../types'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useToast } from '~/composables/toast.composable'
import { useSourceSearchStore } from '../stores/sources-search.store'
import { useSourceSearch } from './sources-search.composable'

const { search, create, deleteById, smAndDown } = vi.hoisted(() => ({
  search: vi.fn(),
  create: vi.fn(),
  deleteById: vi.fn(),
  smAndDown: { value: false },
}))

vi.mock('./sources.api', () => ({
  createSourceApi: () => ({ search }),
}))

vi.mock('~/features/comics/composables/comics.api', () => ({
  createComicsApi: () => ({ create, deleteById }),
}))

function i18nStub(): { t: (key: string) => string } {
  return { t: (key: string) => key }
}

function displayStub(): { smAndDown: { value: boolean } } {
  return { smAndDown }
}

mockNuxtImport('useI18n', () => i18nStub)
mockNuxtImport('useDisplay', () => displayStub)

function item(slug: string, internalId?: string): SourceSearchItem {
  return {
    slug,
    title: slug,
    cover: `/cover/${slug}`,
    publicUrl: '',
    sourceUrl: '',
    status: 'ongoing',
    type: 'manhwa',
    author: '',
    artist: '',
    description: '',
    altTitles: [],
    genres: [],
    chapterCount: 1,
    rating: 0,
    releaseYear: 2020,
    lastChapterAt: new Date('2026-01-01'),
    updatedAt: new Date('2026-01-01'),
    createdAt: new Date('2026-01-01'),
    ...(internalId && { internalId }),
  }
}

function apiError(status: number): { success: false, error: { status: number, message: string } } {
  return { success: false, error: { status, message: 'boom' } }
}

beforeEach(() => {
  setActivePinia(createPinia())
  useToast().messages.value.length = 0
  vi.spyOn(console, 'error').mockImplementation(() => {})
  search.mockReset()
  create.mockReset()
  deleteById.mockReset()
  smAndDown.value = false
})

describe('useSourceSearch', () => {
  it('initializes search composable with store bindings', () => {
    const composable = useSourceSearch('asurascans', { doSearch: false })
    expect(composable.search.value).toBeUndefined()
    expect(composable.page.value).toBe(1)
    expect(composable.isLoading.value).toBe(false)
  })
})

describe('useSourceSearch library', () => {
  it('adds a comic and stores the returned id', async () => {
    create.mockResolvedValue({ success: true, data: { id: 'c1', slug: 'solo-leveling', source: 'asurascans', status: 'ongoing', chapterCount: 0 } })
    const sourceSearch = useSourceSearch('asurascans', { doSearch: false })
    useSourceSearchStore('asurascans').setComics([item('solo-leveling')])

    await sourceSearch.addComicInLibrary(item('solo-leveling'))

    expect(create).toHaveBeenCalledWith({ source: 'asurascans', slug: 'solo-leveling' })
    expect(useSourceSearchStore('asurascans').comics[0]?.internalId).toBe('c1')
  })

  it('toasts a dedicated message on a 409', async () => {
    create.mockResolvedValue(apiError(409))

    await useSourceSearch('asurascans', { doSearch: false }).addComicInLibrary(item('solo-leveling'))

    expect(useToast().messages.value).toEqual([{ text: 'comics.create.error.alreadyExists', color: 'error' }])
  })

  it('does not add a comic that already has an internal id', async () => {
    await useSourceSearch('asurascans', { doSearch: false }).addComicInLibrary(item('solo-leveling', 'c1'))

    expect(create).not.toHaveBeenCalled()
  })

  it('removes a comic and clears its internal id', async () => {
    deleteById.mockResolvedValue({ success: true, data: undefined })
    const sourceSearch = useSourceSearch('asurascans', { doSearch: false })
    useSourceSearchStore('asurascans').setComics([item('solo-leveling', 'c1')])

    await sourceSearch.removeComicFromLibrary(item('solo-leveling', 'c1'))

    expect(deleteById).toHaveBeenCalledWith('c1')
    expect(useSourceSearchStore('asurascans').comics[0]?.internalId).toBeUndefined()
  })

  it('does not delete a comic without an internal id', async () => {
    await useSourceSearch('asurascans', { doSearch: false }).removeComicFromLibrary(item('solo-leveling'))

    expect(deleteById).not.toHaveBeenCalled()
  })

  it('toasts an unknown error when delete fails', async () => {
    deleteById.mockResolvedValue(apiError(500))

    await useSourceSearch('asurascans', { doSearch: false }).removeComicFromLibrary(item('solo-leveling', 'c1'))

    expect(useToast().messages.value).toEqual([{ text: 'error.unknown', color: 'error' }])
  })
})

describe('useSourceSearch pagination', () => {
  it('searches with page 1 and stores hasNextPage', async () => {
    search.mockResolvedValue({ success: true, data: { items: [item('solo')], hasNextPage: true } })
    const sourceSearch = useSourceSearch('asurascans', { doSearch: true })
    await vi.waitFor(() => {
      expect(search).toHaveBeenCalledWith(expect.objectContaining({ page: 1 }))
    })
    expect(sourceSearch.hasNextPage.value).toBe(true)
    expect(sourceSearch.series.value).toHaveLength(1)
  })

  it('resets page to 1 when filters change', async () => {
    search.mockResolvedValue({ success: true, data: { items: [], hasNextPage: false } })
    const sourceSearch = useSourceSearch('asurascans', { doSearch: true })
    await vi.waitFor(() => expect(search).toHaveBeenCalled())
    sourceSearch.page.value = 3
    sourceSearch.sort.value = 'latest'
    await vi.waitFor(() => {
      expect(sourceSearch.page.value).toBe(1)
    })
  })
})
