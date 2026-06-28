// SPDX-License-Identifier: AGPL-3.0-or-later

import type { UserDto } from '#shared/dto/identity'
import { afterEach, describe, expect, it, vi } from 'vitest'

const apiFetch = vi.fn()
vi.mock('~/utils/api/api-fetch', () => ({ apiFetch }))

const { createAuthApi } = await import('~/features/auth/api/auth.api')
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
  showNsfw: false,
}

describe('authApi', () => {
  afterEach(() => apiFetch.mockReset())

  it('getSetupStatus resolves the full status', async () => {
    apiFetch.mockResolvedValue({ required: true, minPasswordLength: 10 })
    const res = await createAuthApi().getSetupStatus()
    expect(res).toEqual({ success: true, data: { required: true, minPasswordLength: 10 } })
  })

  it('setup posts the body and unwraps res.user', async () => {
    apiFetch.mockResolvedValue({ user })
    const body = { accountName: 'alice', displayName: 'A', password: 'secret1234' }
    const res = await createAuthApi().setup(body)
    expect(res).toEqual({ success: true, data: user })
    expect(apiFetch).toHaveBeenCalledWith('/api/auth/setup', { method: 'POST', body })
  })

  it('login posts to /api/auth/login', async () => {
    apiFetch.mockResolvedValue({ ok: true })
    const res = await createAuthApi().login({ accountName: 'alice', password: 'secret1234' })
    expect(res).toEqual({ success: true, data: undefined })
    expect(apiFetch).toHaveBeenCalledWith('/api/auth/login', { method: 'POST', body: { accountName: 'alice', password: 'secret1234' } })
  })

  it('me unwraps res.user', async () => {
    apiFetch.mockResolvedValue({ user })
    const res = await createAuthApi().me()
    expect(res).toEqual({ success: true, data: user })
  })

  it('updateMe PATCHes /api/auth/me and unwraps res.user', async () => {
    apiFetch.mockResolvedValue({ user })
    const res = await createAuthApi().updateMe({ displayName: 'New Name' })
    expect(res).toEqual({ success: true, data: user })
    expect(apiFetch).toHaveBeenCalledWith('/api/auth/me', { method: 'PATCH', body: { displayName: 'New Name' } })
  })

  it('changePassword posts to /api/auth/me/password', async () => {
    apiFetch.mockResolvedValue({ ok: true })
    const body = { currentPassword: 'old1234567', newPassword: 'new1234567', logoutOtherDevices: true }
    const res = await createAuthApi().changePassword(body)
    expect(res).toEqual({ success: true, data: undefined })
    expect(apiFetch).toHaveBeenCalledWith('/api/auth/me/password', { method: 'POST', body })
  })

  it('returns a normalised ApiError instead of throwing', async () => {
    apiFetch.mockRejectedValue(Object.assign(new Error('Unauthenticated'), {
      statusCode: 401,
      data: { statusMessage: 'Unauthenticated' },
    }))
    const res = await createAuthApi().me()
    expect(res.success).toBe(false)
    if (!res.success) {
      expect(res.error).toBeInstanceOf(ApiError)
      expect(res.error.status).toBe(401)
      expect(res.error.message).toBe('Unauthenticated')
    }
  })
})
