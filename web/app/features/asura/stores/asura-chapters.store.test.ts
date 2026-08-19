// SPDX-License-Identifier: AGPL-3.0-or-later

import type { AsuraComicChapter } from '../types'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useAsuraChaptersStore } from './asura-chapters.store'

function chapter(overrides: Partial<AsuraComicChapter> = {}): AsuraComicChapter {
  return {
    id: 'ch-1',
    title: 'One',
    number: 1,
    publishedAt: new Date('2026-01-01'),
    ...overrides,
  }
}

describe('useAsuraChaptersStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('starts empty', () => {
    expect(useAsuraChaptersStore().chapters).toEqual([])
  })

  it('setChapters replaces the list', () => {
    const store = useAsuraChaptersStore()
    store.setChapters([chapter(), chapter({ id: 'ch-2', number: 2 })])

    expect(store.chapters.map(c => c.id)).toEqual(['ch-1', 'ch-2'])
  })

  it('invalidate clears the list', () => {
    const store = useAsuraChaptersStore()
    store.setChapters([chapter()])

    store.invalidate()

    expect(store.chapters).toEqual([])
  })
})
