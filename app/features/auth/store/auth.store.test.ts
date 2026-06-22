// SPDX-License-Identifier: AGPL-3.0-or-later
import type { UserDto } from '#shared/dto/identity'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useAuthStore } from '~/features/auth/store/auth.store'

const adminUser: UserDto = {
  id: 'u1',
  email: 'admin@uchiyomi.test',
  displayName: 'Admin',
  role: 'ADMIN',
  status: 'ACTIVE',
  canManageExtensions: true,
  canDownload: true,
  allowNsfw: false,
}

describe('useAuthStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('starts unauthenticated', () => {
    const store = useAuthStore()
    expect(store.user).toBeUndefined()
    expect(store.isAuthenticated).toBe(false)
    expect(store.isAdmin).toBe(false)
    expect(store.capabilities).toEqual({ canManageExtensions: false, canDownload: false, allowNsfw: false })
  })

  it('reflects the user after setUser, and resets on clear', () => {
    const store = useAuthStore()
    store.setUser(adminUser)
    expect(store.isAuthenticated).toBe(true)
    expect(store.isAdmin).toBe(true)
    expect(store.capabilities).toEqual({ canManageExtensions: true, canDownload: true, allowNsfw: false })
    store.clear()
    expect(store.user).toBeUndefined()
    expect(store.isAuthenticated).toBe(false)
  })

  it('setSetupStatus records needsAdmin and the password policy', () => {
    const store = useAuthStore()
    store.setSetupStatus({ required: true, minPasswordLength: 10 })
    expect(store.needsAdmin).toBe(true)
    expect(store.minPasswordLength).toBe(10)
  })

  it('markAdminCreated clears needsAdmin', () => {
    const store = useAuthStore()
    store.setSetupStatus({ required: true, minPasswordLength: 10 })
    store.markAdminCreated()
    expect(store.needsAdmin).toBe(false)
  })
})
