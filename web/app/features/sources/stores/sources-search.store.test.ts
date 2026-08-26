// SPDX-License-Identifier: AGPL-3.0-or-later

import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useSourceSearchStore } from './sources-search.store'

describe('useSourceSearchStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('isolates state between different sources', () => {
    const asuraStore = useSourceSearchStore('asurascans')
    const kingStore = useSourceSearchStore('kingofshojo')

    asuraStore.setSearch('solo')
    expect(asuraStore.search).toBe('solo')
    expect(kingStore.search).toBeUndefined()
  })

  it('updates internalId on comics list', () => {
    const store = useSourceSearchStore('asurascans')
    store.setComics([
      { slug: 'manga-1', title: 'Manga 1' } as any,
    ])

    store.setComicInternalId('manga-1', 'internal-id-123')
    expect(store.comics).toHaveLength(1)
    expect(store.comics[0]!.internalId).toBe('internal-id-123')
  })

  it('invalidates state properly', () => {
    const store = useSourceSearchStore('asurascans')
    store.setSearch('query')
    store.setPage(3)
    store.invalidate()

    expect(store.search).toBeUndefined()
    expect(store.page).toBe(1)
  })

  it('patches status type and chapterCount on both comic lists', () => {
    const store = useSourceSearchStore('kingofshojo')
    const comic = { slug: 'manga-1', title: 'Manga 1', status: 'ongoing', type: 'manga', chapterCount: 0, internalId: 'keep-me' } as any
    store.setComics([comic])
    store.setAccumulatedComics([comic])

    store.patchComic('manga-1', { status: 'completed', type: 'manhwa', chapterCount: 42 })

    expect(store.comics[0]).toMatchObject({
      slug: 'manga-1',
      status: 'completed',
      type: 'manhwa',
      chapterCount: 42,
      internalId: 'keep-me',
    })
    expect(store.accumulatedComics[0]).toMatchObject({
      slug: 'manga-1',
      status: 'completed',
      type: 'manhwa',
      chapterCount: 42,
      internalId: 'keep-me',
    })
  })

  it('does not add a comic when patching an unknown slug', () => {
    const store = useSourceSearchStore('kingofshojo')
    store.setComics([])
    store.patchComic('missing', { chapterCount: 9 })
    expect(store.comics).toEqual([])
  })

  it('tracks infos loading per slug and clears it on invalidate', () => {
    const store = useSourceSearchStore('kingofshojo')
    store.setInfosLoading('manga-1', true)
    expect(store.infosLoading['manga-1']).toBe(true)

    store.setInfosLoading('manga-1', false)
    expect(store.infosLoading['manga-1']).toBeUndefined()

    store.setInfosLoading('manga-1', true)
    store.invalidate()
    expect(store.infosLoading).toEqual({})
  })
})
