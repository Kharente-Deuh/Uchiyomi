// SPDX-License-Identifier: AGPL-3.0-or-later

import type { AsuraSearchItem } from '../composables/asura.api'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useAsuraSearchStore } from './asura-search.store'

function item(slug: string, internalId?: string): AsuraSearchItem {
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

describe('useAsuraSearchStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('starts with default sort and offset', () => {
    const store = useAsuraSearchStore()

    expect(store.sort).toBe('popular')
    expect(store.sortOrder).toBe('desc')
    expect(store.offset).toBe(1)
    expect(store.comics).toEqual([])
  })

  it('setComics replaces the current page', () => {
    const store = useAsuraSearchStore()
    store.setComics([item('solo-leveling')])
    expect(store.comics).toHaveLength(1)
  })

  it('setComicInternalId updates the matching slug', () => {
    const store = useAsuraSearchStore()
    store.setComics([item('solo-leveling')])
    store.setAccumulatedComics([item('solo-leveling')])

    store.setComicInternalId('solo-leveling', 'c1')

    expect(store.comics[0]?.internalId).toBe('c1')
    expect(store.accumulatedComics[0]?.internalId).toBe('c1')
  })

  it('setComicInternalId ignores an unknown slug', () => {
    const store = useAsuraSearchStore()
    store.setComics([item('solo-leveling')])

    store.setComicInternalId('one-piece', 'c1')

    expect(store.comics[0]?.internalId).toBeUndefined()
  })

  it('invalidate restores defaults', () => {
    const store = useAsuraSearchStore()
    store.setSearch('solo')
    store.setSort('title')
    store.setComics([item('solo-leveling')])
    store.setOffset(3)

    store.invalidate()

    expect(store.search).toBeUndefined()
    expect(store.sort).toBe('popular')
    expect(store.offset).toBe(1)
    expect(store.comics).toEqual([])
  })
})
