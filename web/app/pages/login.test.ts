// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { ProviderSummary } from '~/features/auth/composables/auth.api'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp } from 'vuetify/components'
import LoginPage from './login.vue'

const { login, fetchProviders, providers, query, navigateTo } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return {
    login: vi.fn(),
    fetchProviders: vi.fn(),
    providers: ref<ProviderSummary[]>([]),
    query: { redirect: undefined as string | undefined, error: undefined as string | undefined },
    navigateTo: vi.fn(),
  }
})

vi.mock('~/features/auth/composables/auth.composable', () => ({
  useAuth: () => ({ login }),
}))

vi.mock('~/features/auth/composables/oidc-providers.composable', () => ({
  useOIDCProviders: () => ({ providers, fetchProviders }),
}))

function routeStub(): { query: typeof query } {
  return { query }
}

mockNuxtImport('useRoute', () => routeStub)
mockNuxtImport('navigateTo', () => navigateTo)

const OidcStub = defineComponent({
  name: 'AuthOidcProviderButtons',
  props: { providers: { type: Array, required: true }, redirect: { type: String, required: true } },
  template: '<div data-test="oidc-buttons" />',
})

async function mount(): Promise<VueWrapper> {
  return mountSuspended(
    { render: () => h(VApp, () => [h(LoginPage)]) },
    { global: { stubs: { AuthOidcProviderButtons: OidcStub } } },
  )
}

beforeEach(() => {
  login.mockReset()
  fetchProviders.mockReset()
  navigateTo.mockReset()
  providers.value = []
  query.redirect = undefined
  query.error = undefined
  login.mockResolvedValue('ok')
})

describe('loginPage', () => {
  it('renders the form and fetches providers', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).toContain('Welcome back')
    expect(wrapper.find('[data-test="login-username"]').exists()).toBe(true)
    expect(fetchProviders).toHaveBeenCalled()
  })

  it('shows an oidc error from the query string', async () => {
    query.error = 'oidcDenied'
    const wrapper = await mount()

    expect(wrapper.find('[data-test="auth-form-error"]').text()).toContain('Sign-in was cancelled.')
  })

  it('signs in and redirects', async () => {
    const wrapper = await mount()

    await wrapper.find('[data-test="login-username"] input').setValue('alice')
    await wrapper.find('[data-test="login-password"] input').setValue('secret')
    await wrapper.find('form').trigger('submit')

    await vi.waitFor(() => expect(login).toHaveBeenCalledWith({ username: 'alice', password: 'secret' }))
    expect(navigateTo).toHaveBeenCalledWith('/feed')
  })

  it('shows invalid credentials', async () => {
    login.mockResolvedValue('invalid-credentials')
    const wrapper = await mount()

    await wrapper.find('[data-test="login-username"] input').setValue('alice')
    await wrapper.find('[data-test="login-password"] input').setValue('wrong')
    await wrapper.find('form').trigger('submit')

    await vi.waitFor(() => expect(wrapper.find('[data-test="auth-form-error"]').text()).toContain('Invalid username/password'))
    expect(navigateTo).not.toHaveBeenCalled()
  })

  it('renders oidc buttons when providers are available', async () => {
    providers.value = [{ id: 'id-okta', slug: 'okta', displayName: 'Okta' }]
    const wrapper = await mount()

    expect(wrapper.find('[data-test="oidc-buttons"]').exists()).toBe(true)
  })
})
