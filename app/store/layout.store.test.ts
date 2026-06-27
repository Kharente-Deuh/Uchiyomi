// SPDX-License-Identifier: AGPL-3.0-or-later
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useLayoutStore } from '~/store/layout.store'

describe('useLayoutStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('starts with paginationEnabled false', () => {
    const store = useLayoutStore()
    expect(store.paginationEnabled).toBe(false)
  })

  it('starts with navigationDrawerCompact false', () => {
    const store = useLayoutStore()
    expect(store.navigationDrawerCompact).toBe(false)
  })

  it('setPaginationEnabled(true) enables pagination', () => {
    const store = useLayoutStore()
    store.setPaginationEnabled(true)
    expect(store.paginationEnabled).toBe(true)
  })

  it('setPaginationEnabled(false) disables pagination after being enabled', () => {
    const store = useLayoutStore()
    store.setPaginationEnabled(true)
    store.setPaginationEnabled(false)
    expect(store.paginationEnabled).toBe(false)
  })

  it('setNavigationDrawerCompact(true) enables compact mode', () => {
    const store = useLayoutStore()
    store.setNavigationDrawerCompact(true)
    expect(store.navigationDrawerCompact).toBe(true)
  })

  it('setNavigationDrawerCompact(false) disables compact mode after being enabled', () => {
    const store = useLayoutStore()
    store.setNavigationDrawerCompact(true)
    store.setNavigationDrawerCompact(false)
    expect(store.navigationDrawerCompact).toBe(false)
  })

  it('paginationEnabled and navigationDrawerCompact are independent', () => {
    const store = useLayoutStore()
    store.setPaginationEnabled(true)
    expect(store.navigationDrawerCompact).toBe(false)
    store.setNavigationDrawerCompact(true)
    expect(store.paginationEnabled).toBe(true)
  })
})
