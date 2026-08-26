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
    expect(store.comics[0].internalId).toBe('internal-id-123')
  })

  it('invalidates state properly', () => {
    const store = useSourceSearchStore('asurascans')
    store.setSearch('query')
    store.setPage(3)
    store.invalidate()

    expect(store.search).toBeUndefined()
    expect(store.page).toBe(1)
  })
})
