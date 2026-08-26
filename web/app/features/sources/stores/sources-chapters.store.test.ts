// SPDX-License-Identifier: AGPL-3.0-or-later

import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useSourceChaptersStore } from './sources-chapters.store'

describe('useSourceChaptersStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('isolates chapters between sources', () => {
    const store1 = useSourceChaptersStore('asurascans')
    const store2 = useSourceChaptersStore('kingofshojo')

    store1.setChapters([{ id: '1', title: 'Ch 1', number: 1 } as any])
    expect(store1.chapters.length).toBe(1)
    expect(store2.chapters.length).toBe(0)
  })

  it('clears chapters on invalidate', () => {
    const store = useSourceChaptersStore('asurascans')
    store.setChapters([{ id: '1', title: 'Ch 1', number: 1 } as any])
    store.invalidate()
    expect(store.chapters.length).toBe(0)
  })
})
