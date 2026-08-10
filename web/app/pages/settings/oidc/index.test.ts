// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { LightOidcProvider } from '~/features/oidc/composables/oidc.api'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import OidcIndexPage from './index.vue'

const { providers, getAll } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return { providers: ref<LightOidcProvider[]>([]), getAll: vi.fn() }
})

vi.mock('~/features/oidc/composables/oidc.composable', () => ({
  useOidc: () => ({ providers, getAll, loading: ref(false), testLoading: ref(false), create: vi.fn(), test: vi.fn() }),
}))

const CreateModalStub = defineComponent({
  name: 'OidcModalCreate',
  props: { modelValue: { type: Boolean, default: false } },
  template: '<div data-test="create-modal" />',
})

const ProviderCardStub = defineComponent({
  name: 'OidcCardProvider',
  props: { id: { type: String, default: '' }, displayName: { type: String, default: '' } },
  template: '<div data-test="provider-card">{{ displayName }}</div>',
})

const provider = {
  id: 'p1',
  displayName: 'PocketID',
  createdAt: new Date('2026-01-02T03:04:05.000Z'),
  userCount: 3,
} satisfies LightOidcProvider

async function mount(): Promise<VueWrapper> {
  return mountSuspended(
    { render: () => h(VApp, () => [h(OidcIndexPage)]) },
    { global: { stubs: { OidcModalCreate: CreateModalStub, OidcCardProvider: ProviderCardStub } } },
  )
}

beforeEach(() => {
  getAll.mockReset()
  getAll.mockResolvedValue(undefined)
  providers.value = []
})

describe('settings/oidc', () => {
  it('loads the provider list on mount', async () => {
    await mount()
    await vi.waitFor(() => expect(getAll).toHaveBeenCalledTimes(1))
  })

  it('offers a second create button when no provider exists', async () => {
    const wrapper = await mount()

    await vi.waitFor(() => {
      expect(wrapper.find('[data-test="provider-card"]').exists()).toBe(false)
      expect(wrapper.findAll('button')).toHaveLength(2)
    })
  })

  it('renders a card per provider', async () => {
    providers.value = [provider, { ...provider, id: 'p2', displayName: 'Keycloak' }]
    const wrapper = await mount()

    await vi.waitFor(() => {
      expect(wrapper.findAll('[data-test="provider-card"]')).toHaveLength(2)
      expect(wrapper.text()).toContain('Keycloak')
    })
  })

  it('drops the empty state as soon as a provider is listed', async () => {
    providers.value = [provider]
    const wrapper = await mount()

    await vi.waitFor(() => expect(wrapper.findAll('button')).toHaveLength(1))
  })

  it('keeps the create modal closed until asked', async () => {
    const wrapper = await mount()

    await vi.waitFor(() => expect(wrapper.findComponent(CreateModalStub).props('modelValue')).toBe(false))
  })

  it('opens the create modal from the header button', async () => {
    const wrapper = await mount()

    await vi.waitFor(() => expect(getAll).toHaveBeenCalled())
    await wrapper.findAll('button')[0]!.trigger('click')

    await vi.waitFor(() => expect(wrapper.findComponent(CreateModalStub).props('modelValue')).toBe(true))
  })

  it('opens the create modal from the empty state button', async () => {
    const wrapper = await mount()

    await vi.waitFor(() => expect(wrapper.findAll('button')).toHaveLength(2))
    await wrapper.findAll('button').at(-1)!.trigger('click')

    await vi.waitFor(() => expect(wrapper.findComponent(CreateModalStub).props('modelValue')).toBe(true))
  })
})
