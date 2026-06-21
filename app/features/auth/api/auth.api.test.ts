// SPDX-License-Identifier: AGPL-3.0-or-later
import type { UserDto } from '#shared/dto/identity'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { createAuthApi } from '~/features/auth/api/auth.api'
import { ApiError } from '~/utils/api'

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
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('isSetupStatusRequired resolves a success Result unwrapping res.required', async () => {
    vi.stubGlobal('$fetch', vi.fn().mockResolvedValue({ required: true }))
    const res = await createAuthApi().isSetupStatusRequired()
    expect(res).toEqual({ success: true, data: true })
  })

  it('setup posts the body and resolves a success Result with res.user', async () => {
    const fetchMock = vi.fn().mockResolvedValue({ user })
    vi.stubGlobal('$fetch', fetchMock)
    const body = { email: 'a@b.c', displayName: 'A', password: 'secret1234' }
    const res = await createAuthApi().setup(body)
    expect(res).toEqual({ success: true, data: user })
    expect(fetchMock).toHaveBeenCalledWith('/api/auth/setup', { method: 'POST', body })
  })

  it('login posts to /api/auth/login and resolves a success Result', async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true })
    vi.stubGlobal('$fetch', fetchMock)
    const res = await createAuthApi().login({ email: 'a@b.c', password: 'secret1234' })
    expect(res).toEqual({ success: true, data: undefined })
    expect(fetchMock).toHaveBeenCalledWith('/api/auth/login', { method: 'POST', body: { email: 'a@b.c', password: 'secret1234' } })
  })

  it('me resolves a success Result unwrapping res.user', async () => {
    vi.stubGlobal('$fetch', vi.fn().mockResolvedValue({ user }))
    const res = await createAuthApi().me()
    expect(res).toEqual({ success: true, data: user })
  })

  it('returns a failure Result with a normalised ApiError instead of throwing', async () => {
    const fetchErr = Object.assign(new Error('Unauthenticated'), {
      statusCode: 401,
      data: { statusMessage: 'Unauthenticated' },
    })
    vi.stubGlobal('$fetch', vi.fn().mockRejectedValue(fetchErr))
    const res = await createAuthApi().me()
    expect(res.success).toBe(false)
    if (!res.success) {
      expect(res.error).toBeInstanceOf(ApiError)
      expect(res.error.status).toBe(401)
      expect(res.error.message).toBe('Unauthenticated')
    }
  })
})
