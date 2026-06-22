// SPDX-License-Identifier: AGPL-3.0-or-later
import type { UserDto } from '#shared/dto/identity'
import { afterEach, describe, expect, it, vi } from 'vitest'

const apiFetch = vi.fn()
vi.mock('~/utils/api/api-fetch', () => ({ apiFetch }))

const { createAuthApi } = await import('~/features/auth/api/auth.api')
const { ApiError } = await import('~/utils/api')

const user: UserDto = {
  id: 'u1',
  email: 'admin@uchiyomi.test',
  displayName: 'Admin',
  role: 'ADMIN',
  status: 'ACTIVE',
  canManageExtensions: true,
  canDownload: true,
  allowNsfw: false,
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
    const body = { email: 'a@b.c', displayName: 'A', password: 'secret1234' }
    const res = await createAuthApi().setup(body)
    expect(res).toEqual({ success: true, data: user })
    expect(apiFetch).toHaveBeenCalledWith('/api/auth/setup', { method: 'POST', body })
  })

  it('login posts to /api/auth/login', async () => {
    apiFetch.mockResolvedValue({ ok: true })
    const res = await createAuthApi().login({ email: 'a@b.c', password: 'secret1234' })
    expect(res).toEqual({ success: true, data: undefined })
    expect(apiFetch).toHaveBeenCalledWith('/api/auth/login', { method: 'POST', body: { email: 'a@b.c', password: 'secret1234' } })
  })

  it('me unwraps res.user', async () => {
    apiFetch.mockResolvedValue({ user })
    const res = await createAuthApi().me()
    expect(res).toEqual({ success: true, data: user })
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
