// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { KingOfShojoSearchItem } from '../types'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useToast } from '~/composables/toast.composable'
import { useKingOfShojoSearchStore } from '../stores/kingofshojo-search.store'
import { useKingOfShojoSearch } from './kingofshojo-search.composable'

const { search, create, deleteById, smAndDown } = vi.hoisted(() => ({
  search: vi.fn(),
  create: vi.fn(),
  deleteById: vi.fn(),
  smAndDown: { value: false },
}))

vi.mock('./kingofshojo.api', () => ({
  createKingOfShojoApi: () => ({ search }),
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

function item(slug: string, internalId?: string): KingOfShojoSearchItem {
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

describe('useKingOfShojoSearch library', () => {
  it('adds a comic and stores the returned id', async () => {
    create.mockResolvedValue({ success: true, data: { id: 'c1', slug: 'solo-leveling', source: 'kingofshojo', status: 'ongoing', chapterCount: 0 } })
    const kingofshojo = useKingOfShojoSearch({ doSearch: false })
    useKingOfShojoSearchStore().setComics([item('solo-leveling')])

    await kingofshojo.addComicInLibrary(item('solo-leveling'))

    expect(create).toHaveBeenCalledWith({ source: 'kingofshojo', slug: 'solo-leveling' })
    expect(useKingOfShojoSearchStore().comics[0]?.internalId).toBe('c1')
  })

  it('toasts a dedicated message on a 409', async () => {
    create.mockResolvedValue(apiError(409))

    await useKingOfShojoSearch({ doSearch: false }).addComicInLibrary(item('solo-leveling'))

    expect(useToast().messages.value).toEqual([{ text: 'comics.create.error.alreadyExists', color: 'error' }])
  })

  it('does not add a comic that already has an internal id', async () => {
    await useKingOfShojoSearch({ doSearch: false }).addComicInLibrary(item('solo-leveling', 'c1'))

    expect(create).not.toHaveBeenCalled()
  })

  it('removes a comic and clears its internal id', async () => {
    deleteById.mockResolvedValue({ success: true, data: undefined })
    const kingofshojo = useKingOfShojoSearch({ doSearch: false })
    useKingOfShojoSearchStore().setComics([item('solo-leveling', 'c1')])

    await kingofshojo.removeComicFromLibrary(item('solo-leveling', 'c1'))

    expect(deleteById).toHaveBeenCalledWith('c1')
    expect(useKingOfShojoSearchStore().comics[0]?.internalId).toBeUndefined()
  })

  it('does not delete a comic without an internal id', async () => {
    await useKingOfShojoSearch({ doSearch: false }).removeComicFromLibrary(item('solo-leveling'))

    expect(deleteById).not.toHaveBeenCalled()
  })

  it('toasts an unknown error when delete fails', async () => {
    deleteById.mockResolvedValue(apiError(500))

    await useKingOfShojoSearch({ doSearch: false }).removeComicFromLibrary(item('solo-leveling', 'c1'))

    expect(useToast().messages.value).toEqual([{ text: 'error.unknown', color: 'error' }])
  })
})
