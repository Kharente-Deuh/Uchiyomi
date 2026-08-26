// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { SourceSearchItem } from '../types'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useToast } from '~/composables/toast.composable'
import { useSourceSearchStore } from '../stores/sources-search.store'
import { useSourceSearch } from './sources-search.composable'

const { search, getInfosBySlug, create, deleteById, smAndDown } = vi.hoisted(() => ({
  search: vi.fn(),
  getInfosBySlug: vi.fn(),
  create: vi.fn(),
  deleteById: vi.fn(),
  smAndDown: { value: false },
}))

vi.mock('./sources.api', () => ({
  createSourceApi: () => ({ search, getInfosBySlug }),
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
  getInfosBySlug.mockReset()
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

describe('useSourceSearch series infos enrichment', () => {
  it('does not call getInfosBySlug for asurascans search', async () => {
    search.mockResolvedValue({ success: true, data: { items: [item('solo')], hasNextPage: false } })
    useSourceSearch('asurascans', { doSearch: true })
    await vi.waitFor(() => expect(search).toHaveBeenCalled())
    expect(getInfosBySlug).not.toHaveBeenCalled()
  })

  it('does not enrich when doSearch is false', async () => {
    useSourceSearch('kingofshojo', { doSearch: false })
    expect(search).not.toHaveBeenCalled()
    expect(getInfosBySlug).not.toHaveBeenCalled()
  })

  it('patches status type and chapterCount from getInfosBySlug', async () => {
    search.mockResolvedValue({ success: true, data: { items: [item('solo')], hasNextPage: false } })
    getInfosBySlug.mockResolvedValue({
      success: true,
      data: { slug: 'solo', status: 'completed', type: 'manhwa', chapterCount: 120 },
    })

    useSourceSearch('kingofshojo', { doSearch: true })
    await vi.waitFor(() => expect(getInfosBySlug).toHaveBeenCalledWith('solo'))
    await vi.waitFor(() => {
      expect(useSourceSearchStore('kingofshojo').comics[0]).toMatchObject({
        status: 'completed',
        type: 'manhwa',
        chapterCount: 120,
      })
    })
  })

  it('does not overwrite internalId when patching infos', async () => {
    search.mockResolvedValue({ success: true, data: { items: [{ ...item('solo'), internalId: 'lib-1' }], hasNextPage: false } })
    getInfosBySlug.mockResolvedValue({
      success: true,
      data: { slug: 'solo', status: 'hiatus', type: 'manga', chapterCount: 3, internalId: 'from-series' },
    })

    useSourceSearch('kingofshojo', { doSearch: true })
    await vi.waitFor(() => {
      expect(useSourceSearchStore('kingofshojo').comics[0]?.internalId).toBe('lib-1')
      expect(useSourceSearchStore('kingofshojo').comics[0]?.status).toBe('hiatus')
    })
  })

  it('leaves a failed slug unchanged, logs, and does not toast', async () => {
    search.mockResolvedValue({
      success: true,
      data: { items: [item('ok'), item('missing')], hasNextPage: false },
    })
    getInfosBySlug.mockImplementation(async (slug: string) => {
      if (slug === 'missing') {
        return { success: false, error: { status: 404, message: 'missing' } }
      }

      return { success: true, data: { slug, status: 'completed', type: 'manhwa', chapterCount: 10 } }
    })

    useSourceSearch('kingofshojo', { doSearch: true })
    await vi.waitFor(() => expect(getInfosBySlug).toHaveBeenCalledTimes(2))
    await vi.waitFor(() => {
      expect(useSourceSearchStore('kingofshojo').comics.find(c => c.slug === 'ok')).toMatchObject({
        status: 'completed',
        chapterCount: 10,
      })
    })
    expect(useSourceSearchStore('kingofshojo').comics.find(c => c.slug === 'missing')?.chapterCount).toBe(1)
    expect(console.error).toHaveBeenCalled()
    expect(useToast().messages.value).toEqual([])
  })

  it('ignores stale enrichment after a new search', async () => {
    const { promise: firstInfos, resolve: resolveFirst } = Promise.withResolvers<unknown>()

    search
      .mockResolvedValueOnce({ success: true, data: { items: [item('old')], hasNextPage: false } })
      .mockResolvedValueOnce({ success: true, data: { items: [item('new')], hasNextPage: false } })

    getInfosBySlug.mockImplementation(async (slug: string) => {
      if (slug === 'old') {
        await firstInfos

        return { success: true, data: { slug: 'old', status: 'completed', type: 'manga', chapterCount: 99 } }
      }

      return { success: true, data: { slug: 'new', status: 'ongoing', type: 'manhwa', chapterCount: 2 } }
    })

    const sourceSearch = useSourceSearch('kingofshojo', { doSearch: true })
    await vi.waitFor(() => expect(getInfosBySlug).toHaveBeenCalledWith('old'))

    sourceSearch.sort.value = 'latest'
    await vi.waitFor(() => expect(getInfosBySlug).toHaveBeenCalledWith('new'))
    await vi.waitFor(() => {
      expect(useSourceSearchStore('kingofshojo').comics.map(c => c.slug)).toEqual(['new'])
    })

    resolveFirst(undefined)
    await Promise.resolve()
    expect(useSourceSearchStore('kingofshojo').comics[0]?.slug).toBe('new')
    expect(useSourceSearchStore('kingofshojo').comics[0]?.chapterCount).not.toBe(99)
  })

  it('exposes infosLoading while enrichment is in flight', async () => {
    const { promise: infosPromise, resolve: resolveInfos } = Promise.withResolvers<unknown>()

    getInfosBySlug.mockImplementation(() => infosPromise)
    search.mockResolvedValue({ success: true, data: { items: [item('solo')], hasNextPage: false } })

    const sourceSearch = useSourceSearch('kingofshojo', { doSearch: true })
    await vi.waitFor(() => expect(sourceSearch.infosLoading.value.solo).toBe(true))

    resolveInfos({ success: true, data: { slug: 'solo', status: 'ongoing', type: 'manga', chapterCount: 4 } })
    await vi.waitFor(() => expect(sourceSearch.infosLoading.value.solo).toBeUndefined())
  })
})
