// SPDX-License-Identifier: AGPL-3.0-or-later

import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useAuthStore } from '~/features/auth/stores/auth.store'
import { useAuth } from './auth.composable'

const loginWithPwd = vi.fn()
const logout = vi.fn()
const getCurrentUser = vi.fn()

vi.mock('./auth.api', () => ({ createAuthApi: () => ({ loginWithPwd, logout }) }))
vi.mock('~/features/users/composables/users.api', () => ({ createUsersApi: () => ({ getCurrentUser }) }))

const user = { id: 'e2d1', username: 'alice', isAdmin: false }
const admin = { id: 'a1', username: 'root', isAdmin: true }

function apiError(status: number): { success: false, error: { status: number } } {
  return { success: false, error: { status } }
}

beforeEach(() => {
  setActivePinia(createPinia())
  loginWithPwd.mockReset()
  logout.mockReset()
  getCurrentUser.mockReset()
})

describe('useAuth().fetchMe', () => {
  it('returns ok and stores the user', async () => {
    getCurrentUser.mockResolvedValue({ success: true, data: user })
    const auth = useAuth()

    await expect(auth.fetchMe()).resolves.toBe('ok')
    expect(auth.user.value).toEqual(user)
  })

  it('returns unauthenticated on a 401', async () => {
    getCurrentUser.mockResolvedValue(apiError(401))

    await expect(useAuth().fetchMe()).resolves.toBe('unauthenticated')
  })

  it('returns unreachable when the API cannot answer', async () => {
    getCurrentUser.mockResolvedValue(apiError(0))

    await expect(useAuth().fetchMe()).resolves.toBe('unreachable')
  })

  it('does not collapse a 500 into unauthenticated', async () => {
    getCurrentUser.mockResolvedValue(apiError(500))

    await expect(useAuth().fetchMe()).resolves.toBe('unreachable')
  })

  it('drops a previously stored user on failure', async () => {
    useAuthStore().setUser(user)
    getCurrentUser.mockResolvedValue(apiError(401))

    const auth = useAuth()
    await auth.fetchMe()

    expect(auth.user.value).toBeUndefined()
  })

  it('refuses to cache: every call hits the API', async () => {
    getCurrentUser.mockResolvedValue({ success: true, data: user })
    const auth = useAuth()

    await auth.fetchMe()
    await auth.fetchMe()

    expect(getCurrentUser).toHaveBeenCalledTimes(2)
  })
})

describe('useAuth().login', () => {
  it('stores the user and returns ok', async () => {
    loginWithPwd.mockResolvedValue({ status: 'ok', user })
    const auth = useAuth()

    await expect(auth.login({ username: 'alice', password: 'pwd' })).resolves.toBe('ok')
    expect(auth.user.value).toEqual(user)
  })

  it('leaves the store empty on invalid credentials', async () => {
    loginWithPwd.mockResolvedValue({ status: 'invalid-credentials' })
    const auth = useAuth()

    await expect(auth.login({ username: 'alice', password: 'nope' })).resolves.toBe('invalid-credentials')
    expect(auth.user.value).toBeUndefined()
  })

  it('forwards an unknown error status', async () => {
    loginWithPwd.mockResolvedValue({ status: 'unknown-error' })

    await expect(useAuth().login({ username: 'alice', password: 'pwd' })).resolves.toBe('unknown-error')
  })
})

describe('useAuth().logout', () => {
  it('clears the user', async () => {
    logout.mockResolvedValue({ success: true, data: undefined })
    const auth = useAuth()
    useAuthStore().setUser(user)

    await auth.logout()

    expect(auth.user.value).toBeUndefined()
  })

  it('clears the user even when the API call fails', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    logout.mockResolvedValue(apiError(500))
    const auth = useAuth()
    useAuthStore().setUser(user)

    await auth.logout()

    expect(auth.user.value).toBeUndefined()
  })

  it('leaves loading false once settled', async () => {
    logout.mockResolvedValue({ success: true, data: undefined })
    const auth = useAuth()

    await auth.logout()

    expect(auth.loading.value).toBe(false)
  })
})

describe('useAuth() flags', () => {
  it('isAuthenticated is false without a user', () => {
    expect(useAuth().isAuthenticated.value).toBe(false)
  })

  it('isAuthenticated follows the store', () => {
    const auth = useAuth()
    useAuthStore().setUser(user)
    expect(auth.isAuthenticated.value).toBe(true)
  })

  it('isAdmin stays false for a regular user', () => {
    const auth = useAuth()
    useAuthStore().setUser(user)
    expect(auth.isAdmin.value).toBe(false)
  })

  it('isAdmin follows the admin flag', () => {
    const auth = useAuth()
    useAuthStore().setUser(admin)
    expect(auth.isAdmin.value).toBe(true)
  })
})
