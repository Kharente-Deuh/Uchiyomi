// SPDX-License-Identifier: AGPL-3.0-or-later

import type { LightComic } from '~/features/comics/types'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useLibraryStore } from './library-search.store'

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

describe('useLibraryStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('starts with default sort and offset', () => {
    const store = useLibraryStore()

    expect(store.sort).toBe('title')
    expect(store.offset).toBe(1)
    expect(store.comics).toEqual([])
    expect(store.search).toBeUndefined()
    expect(store.source).toBeUndefined()
  })

  it('setComics replaces the current page', () => {
    const store = useLibraryStore()
    store.setComics([comic()])
    expect(store.comics).toHaveLength(1)
  })

  it('setAccumulatedComics replaces the accumulated list', () => {
    const store = useLibraryStore()
    store.setAccumulatedComics([comic('c1'), comic('c2')])
    expect(store.accumulatedComics.map(c => c.id)).toEqual(['c1', 'c2'])
  })

  it('invalidate restores defaults', () => {
    const store = useLibraryStore()
    store.setSearch('solo')
    store.setSort('addedAt')
    store.setStatus('ongoing')
    store.setType('manhwa')
    store.setSource('asurascans')
    store.setComics([comic()])
    store.setAccumulatedComics([comic()])
    store.setLoading(true)
    store.setOffset(3)

    store.invalidate()

    expect(store.search).toBeUndefined()
    expect(store.sort).toBe('title')
    expect(store.status).toBeUndefined()
    expect(store.type).toBeUndefined()
    expect(store.source).toBeUndefined()
    expect(store.offset).toBe(1)
    expect(store.comics).toEqual([])
    expect(store.accumulatedComics).toEqual([])
    expect(store.loading).toBe(false)
  })
})
