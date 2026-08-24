// SPDX-License-Identifier: AGPL-3.0-or-later

import type { AsuraScansComicChapter } from '../types'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useAsuraScansChaptersStore } from './asurascans-chapters.store'

function chapter(overrides: Partial<AsuraScansComicChapter> = {}): AsuraScansComicChapter {
  return {
    id: 'ch-1',
    title: 'One',
    number: 1,
    publishedAt: new Date('2026-01-01'),
    ...overrides,
  }
}

describe('useAsuraScansChaptersStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('starts empty', () => {
    expect(useAsuraScansChaptersStore().chapters).toEqual([])
  })

  it('setChapters replaces the list', () => {
    const store = useAsuraScansChaptersStore()
    store.setChapters([chapter(), chapter({ id: 'ch-2', number: 2 })])

    expect(store.chapters.map(c => c.id)).toEqual(['ch-1', 'ch-2'])
  })

  it('invalidate clears the list', () => {
    const store = useAsuraScansChaptersStore()
    store.setChapters([chapter()])

    store.invalidate()

    expect(store.chapters).toEqual([])
  })
})
