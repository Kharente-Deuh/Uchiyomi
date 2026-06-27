import type { UserDto } from '#shared/dto/identity/user.dto'
// SPDX-License-Identifier: AGPL-3.0-or-later
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

// navigateTo must be mocked before the middleware module is imported.
const { navigateToMock } = vi.hoisted(() => ({
  navigateToMock: vi.fn(),
}))

mockNuxtImport('navigateTo', () => navigateToMock)

const { default: authMiddleware } = await import('~/middleware/auth.global')
const { useAuthStore } = await import('~/features/auth/store/auth.store')

const adminUser: UserDto = {
  id: 'u1',
  accountName: 'admin',
  displayName: 'Admin',
  role: 'ADMIN',
  status: 'ACTIVE',
  canManageExtensions: true,
  canDownload: true,
  allowNsfw: false,
  showNsfw: false,
}

function makeRoute(path: string, fullPath?: string): Parameters<typeof authMiddleware>[0] {
  return { path, fullPath: fullPath ?? path, params: {}, name: undefined, query: {}, hash: '', matched: [], meta: {}, redirectedFrom: undefined } as unknown as Parameters<typeof authMiddleware>[0]
}

describe('auth.global middleware', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('redirects unauthenticated user to /login when accessing a protected route', async () => {
    // authStore defaults to not authenticated
    await authMiddleware(makeRoute('/settings'), makeRoute('/'))
    expect(navigateToMock).toHaveBeenCalledWith('/login?redirect=%2Fsettings')
  })

  it('allows an authenticated user through a protected route', async () => {
    const authStore = useAuthStore()
    authStore.setUser(adminUser)
    await authMiddleware(makeRoute('/settings'), makeRoute('/'))
    expect(navigateToMock).not.toHaveBeenCalled()
  })

  it('redirects authenticated user away from /login to /', async () => {
    const authStore = useAuthStore()
    authStore.setUser(adminUser)
    await authMiddleware(makeRoute('/login'), makeRoute('/'))
    expect(navigateToMock).toHaveBeenCalledWith('/')
  })

  it('allows unauthenticated user to access /login', async () => {
    await authMiddleware(makeRoute('/login'), makeRoute('/'))
    expect(navigateToMock).not.toHaveBeenCalled()
  })

  it('redirects every route to /setup during first run (needsAdmin = true)', async () => {
    const authStore = useAuthStore()
    authStore.setSetupStatus({ required: true, minPasswordLength: 8 })
    await authMiddleware(makeRoute('/'), makeRoute('/'))
    expect(navigateToMock).toHaveBeenCalledWith('/setup')
  })

  it('allows /setup itself during first run', async () => {
    const authStore = useAuthStore()
    authStore.setSetupStatus({ required: true, minPasswordLength: 8 })
    await authMiddleware(makeRoute('/setup'), makeRoute('/'))
    expect(navigateToMock).not.toHaveBeenCalled()
  })

  it('does not call navigateTo when the redirect target equals the current fullPath', async () => {
    // resolveAuthGuard returns { redirect: '/login?redirect=...' } for a guest on
    // a protected route, which will never equal to.fullPath, but if a route
    // somehow had redirect === fullPath, navigateTo should be skipped.
    await authMiddleware(makeRoute('/login', '/login'), makeRoute('/'))
    // Guest accessing /login → no redirect (already handled by resolveAuthGuard)
    expect(navigateToMock).not.toHaveBeenCalled()
  })
})
