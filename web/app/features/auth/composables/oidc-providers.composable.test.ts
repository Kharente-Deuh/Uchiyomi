// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useOIDCProviders } from './oidc-providers.composable'

const getProviders = vi.fn()

vi.mock('./auth.api', () => ({ createAuthApi: () => ({ getProviders }) }))

const providers = [
  { id: 'id-google', slug: 'google', displayName: 'Google' },
  { id: 'id-okta', slug: 'okta', displayName: 'Okta' },
]

beforeEach(() => {
  getProviders.mockReset()
})

describe('useOIDCProviders().fetchProviders', () => {
  it('populates providers on success', async () => {
    getProviders.mockResolvedValue({ success: true, data: providers })
    const oidc = useOIDCProviders()

    await oidc.fetchProviders()

    expect(oidc.providers.value).toEqual(providers)
  })

  it('empties providers and logs on failure', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    getProviders.mockResolvedValue({ success: false, error: { status: 500 } })
    const oidc = useOIDCProviders()
    oidc.providers.value = providers

    await oidc.fetchProviders()

    expect(oidc.providers.value).toEqual([])
    expect(console.error).toHaveBeenCalled()
  })

  it('toggles loading around a successful call', async () => {
    let loadingDuringCall: boolean | undefined
    const oidc = useOIDCProviders()
    getProviders.mockImplementation(async () => {
      loadingDuringCall = oidc.loading.value

      return { success: true, data: providers }
    })

    await oidc.fetchProviders()

    expect(loadingDuringCall).toBe(true)
    expect(oidc.loading.value).toBe(false)
  })

  it('toggles loading around a failed call', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
    let loadingDuringCall: boolean | undefined
    const oidc = useOIDCProviders()
    getProviders.mockImplementation(async () => {
      loadingDuringCall = oidc.loading.value

      return { success: false, error: { status: 500 } }
    })

    await oidc.fetchProviders()

    expect(loadingDuringCall).toBe(true)
    expect(oidc.loading.value).toBe(false)
  })
})
