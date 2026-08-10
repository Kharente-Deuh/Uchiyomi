// SPDX-License-Identifier: AGPL-3.0-or-later

import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useAuthStore } from './auth.store'

const user = { id: 'e2d1', username: 'alice', isAdmin: true }

describe('useAuthStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('starts with no user', () => {
    expect(useAuthStore().user).toBeUndefined()
  })

  it('setUser holds the user', () => {
    const store = useAuthStore()
    store.setUser(user)
    expect(store.user).toEqual(user)
  })

  it('setUser replaces a previously held user', () => {
    const store = useAuthStore()
    store.setUser(user)
    store.setUser({ id: 'f3a2', username: 'bob', isAdmin: false })
    expect(store.user).toEqual({ id: 'f3a2', username: 'bob', isAdmin: false })
  })

  it('invalidate drops the user', () => {
    const store = useAuthStore()
    store.setUser(user)
    store.invalidate()
    expect(store.user).toBeUndefined()
  })

  it('invalidate is a no-op when no user is held', () => {
    const store = useAuthStore()
    store.invalidate()
    expect(store.user).toBeUndefined()
  })
})
