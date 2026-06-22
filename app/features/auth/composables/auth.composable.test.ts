// SPDX-License-Identifier: AGPL-3.0-or-later
import type { UserDto } from '#shared/dto/identity'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const { mockApi } = vi.hoisted(() => ({
  mockApi: {
    getSetupStatus: vi.fn(),
    setup: vi.fn(),
    login: vi.fn(),
    logout: vi.fn(),
    me: vi.fn(),
  },
}))

vi.mock('~/features/auth/api', () => ({ createAuthApi: () => mockApi }))

const { clearMock } = vi.hoisted(() => ({ clearMock: vi.fn() }))

function useUserSessionMock(): { loggedIn: Ref<boolean>, clear: () => void, fetch: () => void } {
  return { loggedIn: ref(false), clear: clearMock, fetch: vi.fn() }
}

mockNuxtImport('useUserSession', () => useUserSessionMock)

const { useAuth } = await import('~/features/auth/composables/auth.composable')
const { useAuthStore } = await import('~/features/auth/store/auth.store')
const { ApiError } = await import('~/utils/api')

const user: UserDto = {
  id: 'u1',
  accountName: 'admin',
  displayName: 'Admin',
  role: 'ADMIN',
  status: 'ACTIVE',
  canManageExtensions: true,
  canDownload: true,
  allowNsfw: false,
}

describe('useAuth', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('login hydrates the store via fetchMe and returns the success Result', async () => {
    mockApi.login.mockResolvedValue({ success: true, data: undefined })
    mockApi.me.mockResolvedValue({ success: true, data: user })
    const auth = useAuth()
    const res = await auth.login({ accountName: 'alice', password: 'secret1234' })
    expect(res).toEqual({ success: true, data: undefined })
    expect(useAuthStore().user).toEqual(user)
    expect(auth.isAuthenticated.value).toBe(true)
    expect(auth.loading.value).toBe(false)
  })

  it('login failure returns the Result, skips /me and leaves the store empty', async () => {
    const error = new ApiError('Invalid credentials', 401)
    mockApi.login.mockResolvedValue({ success: false, error })
    const auth = useAuth()
    const res = await auth.login({ accountName: 'alice', password: 'wrong' })
    expect(res).toEqual({ success: false, error })
    expect(mockApi.me).not.toHaveBeenCalled()
    expect(useAuthStore().user).toBeUndefined()
    expect(auth.loading.value).toBe(false)
  })

  it('setup hydrates the store from the setup Result', async () => {
    mockApi.setup.mockResolvedValue({ success: true, data: user })
    const auth = useAuth()
    const res = await auth.setup({ accountName: 'alice', displayName: 'A', password: 'secret1234' })
    expect(res).toEqual({ success: true, data: user })
    expect(useAuthStore().user).toEqual(user)
  })

  it('logout clears the store and the session cookie on success', async () => {
    mockApi.logout.mockResolvedValue({ success: true, data: undefined })
    const store = useAuthStore()
    store.setUser(user)
    const auth = useAuth()
    const res = await auth.logout()
    expect(res.success).toBe(true)
    expect(store.user).toBeUndefined()
    expect(clearMock).toHaveBeenCalledOnce()
  })

  it('fetchMe stores the user on success', async () => {
    mockApi.me.mockResolvedValue({ success: true, data: user })
    const auth = useAuth()
    await auth.fetchMe()
    expect(useAuthStore().user).toEqual(user)
  })

  it('fetchMe clears the store + session on a 401 and resolves void', async () => {
    mockApi.me.mockResolvedValue({ success: false, error: new ApiError('Unauthenticated', 401) })
    const store = useAuthStore()
    store.setUser(user)
    const auth = useAuth()
    await expect(auth.fetchMe()).resolves.toBeUndefined()
    expect(store.user).toBeUndefined()
    expect(clearMock).toHaveBeenCalledOnce()
  })

  it('fetchMe leaves the store untouched on a non-401 failure', async () => {
    mockApi.me.mockResolvedValue({ success: false, error: new ApiError('Server error', 500) })
    const store = useAuthStore()
    store.setUser(user)
    const auth = useAuth()
    await auth.fetchMe()
    expect(store.user).toEqual(user)
    expect(clearMock).not.toHaveBeenCalled()
  })

  it('getSetupStatus records the status in the store', async () => {
    mockApi.getSetupStatus.mockResolvedValue({ success: true, data: { required: true, minPasswordLength: 10 } })
    const { getSetupStatus, needsAdmin, minPasswordLength } = useAuth()
    const res = await getSetupStatus()
    expect(res.success).toBe(true)
    expect(needsAdmin.value).toBe(true)
    expect(minPasswordLength.value).toBe(10)
  })

  it('setup marks the admin created on success', async () => {
    mockApi.setup.mockResolvedValue({ success: true, data: user })
    const store = useAuthStore()
    store.setSetupStatus({ required: true, minPasswordLength: 8 })
    const { setup, needsAdmin } = useAuth()
    await setup({ accountName: 'alice', displayName: 'A', password: 'longenough1' })
    expect(needsAdmin.value).toBe(false)
  })
})
