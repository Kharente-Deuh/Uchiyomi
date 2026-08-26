// SPDX-License-Identifier: AGPL-3.0-or-later

import type { AsuraScansSearchItem } from '../types'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useAsuraScansSearchStore } from './asurascans-search.store'

function item(slug: string, internalId?: string): AsuraScansSearchItem {
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
    latestChapters: [],
    chapterCount: 1,
    rating: 0,
    releaseYear: 2020,
    lastChapterAt: new Date('2026-01-01'),
    updatedAt: new Date('2026-01-01'),
    createdAt: new Date('2026-01-01'),
    ...(internalId && { internalId }),
  }
}

describe('useAsuraScansSearchStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('starts with default sort and page', () => {
    const store = useAsuraScansSearchStore()

    expect(store.sort).toBe('popular')
    expect(store.page).toBe(1)
    expect(store.comics).toEqual([])
  })

  it('setComics replaces the current page', () => {
    const store = useAsuraScansSearchStore()
    store.setComics([item('solo-leveling')])
    expect(store.comics).toHaveLength(1)
  })

  it('setComicInternalId updates the matching slug', () => {
    const store = useAsuraScansSearchStore()
    store.setComics([item('solo-leveling')])
    store.setAccumulatedComics([item('solo-leveling')])

    store.setComicInternalId('solo-leveling', 'c1')

    expect(store.comics[0]?.internalId).toBe('c1')
    expect(store.accumulatedComics[0]?.internalId).toBe('c1')
  })

  it('setComicInternalId ignores an unknown slug', () => {
    const store = useAsuraScansSearchStore()
    store.setComics([item('solo-leveling')])

    store.setComicInternalId('one-piece', 'c1')

    expect(store.comics[0]?.internalId).toBeUndefined()
  })

  it('invalidate restores defaults', () => {
    const store = useAsuraScansSearchStore()
    store.setSearch('solo')
    store.setSort('title')
    store.setComics([item('solo-leveling')])
    store.setPage(3)

    store.invalidate()

    expect(store.search).toBeUndefined()
    expect(store.sort).toBe('popular')
    expect(store.page).toBe(1)
    expect(store.comics).toEqual([])
  })
})
