// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '~/utils/api'
import { createAuthApi, oidcStartUrl } from './auth.api'

const call = vi.fn()

vi.mock('~/utils/api', async importOriginal => ({
  ...(await importOriginal<typeof import('~/utils/api')>()),
  initApi: () => call,
}))

describe('createAuthApi().loginWithPwd', () => {
  beforeEach(() => {
    call.mockReset()
  })

  const user = { id: 'e2d1', username: 'alice', isAdmin: true }

  it('returns the user carried by a successful login', async () => {
    call.mockResolvedValue(user)

    await expect(createAuthApi().loginWithPwd({ username: 'a', password: 'b' })).resolves.toEqual({
      status: 'ok',
      user,
    })
  })

  it('returns invalid-credentials on a 401', async () => {
    call.mockRejectedValue({ statusCode: 401, data: {} })

    await expect(createAuthApi().loginWithPwd({ username: 'a', password: 'b' })).resolves.toEqual({
      status: 'invalid-credentials',
    })
  })

  it('returns unknown-error on a 500', async () => {
    call.mockRejectedValue({ statusCode: 500, data: {} })

    await expect(createAuthApi().loginWithPwd({ username: 'a', password: 'b' })).resolves.toEqual({
      status: 'unknown-error',
    })
  })
})

describe('createAuthApi().getProviders', () => {
  beforeEach(() => {
    call.mockReset()
  })

  const providers = [
    { id: 'id-google', slug: 'google', displayName: 'Google' },
    { id: 'id-okta', slug: 'okta', displayName: 'Okta' },
  ]

  it('returns the provider list on success', async () => {
    call.mockResolvedValue(providers)

    await expect(createAuthApi().getProviders()).resolves.toEqual({
      success: true,
      data: providers,
    })
  })

  it('returns an ApiError on failure', async () => {
    call.mockRejectedValue({ statusCode: 500, data: {} })

    const result = await createAuthApi().getProviders()

    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error).toBeInstanceOf(ApiError)
      expect(result.error.status).toBe(500)
    }
  })
})

describe('createAuthApi().logout', () => {
  beforeEach(() => {
    call.mockReset()
  })

  it('returns the parsed response on success', async () => {
    const data = { endSessionUrl: 'https://id.example.org/logout' }
    call.mockResolvedValue(data)

    await expect(createAuthApi().logout()).resolves.toEqual({
      success: true,
      data,
    })
    expect(call).toHaveBeenCalledWith('/logout', { method: 'POST' })
  })

  it('returns an empty response when no endSessionUrl is provided', async () => {
    call.mockResolvedValue({})

    await expect(createAuthApi().logout()).resolves.toEqual({
      success: true,
      data: {},
    })
  })

  it('returns an ApiError on failure', async () => {
    call.mockRejectedValue({ statusCode: 500, data: {} })

    const result = await createAuthApi().logout()

    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error).toBeInstanceOf(ApiError)
      expect(result.error.status).toBe(500)
    }
  })
})

describe('oidcStartUrl', () => {
  it('builds the start URL with an encoded redirect', () => {
    expect(oidcStartUrl('google', '/library')).toBe('/api/auth/oidc/google/start?redirect=%2Flibrary')
  })

  it('encodes special characters in the redirect', () => {
    expect(oidcStartUrl('okta', '/library?tag=sci-fi&sort=asc')).toBe(
      '/api/auth/oidc/okta/start?redirect=%2Flibrary%3Ftag%3Dsci-fi%26sort%3Dasc',
    )
  })

  it('encodes unicode characters in the redirect', () => {
    expect(oidcStartUrl('okta', '/library/漫画')).toBe(
      '/api/auth/oidc/okta/start?redirect=%2Flibrary%2F%E6%BC%AB%E7%94%BB',
    )
  })
})
