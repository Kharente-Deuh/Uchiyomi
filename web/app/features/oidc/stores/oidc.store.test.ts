// SPDX-License-Identifier: AGPL-3.0-or-later

import type { OidcProviderDetails } from '~/features/oidc/composables/oidc.api'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useOidcProviderStore } from './oidc.store'

const provider = {
  id: 'p1',
  displayName: 'PocketID',
  slug: 'pocket-id',
  issuerUrl: 'https://id.example.org',
  clientId: 'client',
  usernameClaim: 'preferred_username',
  scopes: ['openid'],
  autoProvision: false,
  createdAt: new Date('2026-01-02T03:04:05.000Z'),
  updatedAt: new Date('2026-02-03T04:05:06.000Z'),
  users: [],
} satisfies OidcProviderDetails

describe('useOidcProviderStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('starts with no provider', () => {
    expect(useOidcProviderStore().provider).toBeUndefined()
  })

  it('setProvider holds the provider', () => {
    const store = useOidcProviderStore()
    store.setProvider(provider)
    expect(store.provider).toEqual(provider)
  })

  it('setProvider replaces a previously held provider', () => {
    const store = useOidcProviderStore()
    store.setProvider(provider)
    store.setProvider({ ...provider, id: 'p2', displayName: 'Keycloak' })
    expect(store.provider?.displayName).toBe('Keycloak')
  })

  it('invalidate drops the provider', () => {
    const store = useOidcProviderStore()
    store.setProvider(provider)
    store.invalidate()
    expect(store.provider).toBeUndefined()
  })
})
