// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { ProviderSummary } from '~/features/auth/composables/auth.api'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { oidcStartUrl } from '~/features/auth/composables/auth.api'
import OidcProviderButtons from './OidcProviderButtons.vue'

const providers: ProviderSummary[] = [
  { id: 'okta', displayName: 'Okta' },
  { id: 'google', displayName: 'Google' },
]

describe('authOidcProviderButtons', () => {
  it('renders nothing when there are no providers', async () => {
    const wrapper = await mountSuspended(OidcProviderButtons, {
      props: { providers: [], redirect: '/library' },
    })

    expect(wrapper.findAll('a')).toHaveLength(0)
  })

  it('renders one real anchor per provider, linking to the oidc start url', async () => {
    const wrapper = await mountSuspended(OidcProviderButtons, {
      props: { providers, redirect: '/library' },
    })

    for (const provider of providers) {
      const anchor = wrapper.find(`[data-test="login-oidc-${provider.id}"]`)
      expect(anchor.exists()).toBe(true)
      expect(anchor.element.tagName).toBe('A')
      expect(anchor.attributes('href')).toBe(oidcStartUrl(provider.id, '/library'))
    }
  })

  it('renders the provider name in the button text', async () => {
    const wrapper = await mountSuspended(OidcProviderButtons, {
      props: { providers, redirect: '/library' },
    })

    expect(wrapper.text()).toContain('Sign in with Okta')
    expect(wrapper.text()).toContain('Sign in with Google')
  })
})
