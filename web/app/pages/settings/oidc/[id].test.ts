// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { OidcProvider } from '~/features/oidc/composables/oidc.api'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import ProviderPage from './[id].vue'

const { provider, fetchProvider, invalidate, onBeforeRouteLeave } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return {
    provider: ref<OidcProvider>(),
    fetchProvider: vi.fn(),
    invalidate: vi.fn(),
    onBeforeRouteLeave: vi.fn(),
  }
})

vi.mock('~/features/oidc/composables/oidc-provider.composable', () => ({
  useOidcProvider: () => ({ provider, fetchProvider, invalidate, fetchLoading: ref(false) }),
}))

vi.mock('vue-router', async importOriginal => ({
  ...(await importOriginal<typeof import('vue-router')>()),
  onBeforeRouteLeave,
}))

function cardStub(name: string): ReturnType<typeof defineComponent> {
  return defineComponent({
    name,
    props: { loading: { type: Boolean, default: false } },
    template: `<div data-test="${name}" />`,
  })
}

const InformationsStub = cardStub('OidcCardCategoryInformations')
const ClaimsStub = cardStub('OidcCardCategoryClaims')
const UsersStub = cardStub('OidcCardCategoryUsers')

const DeleteModalStub = defineComponent({
  name: 'OidcModalDelete',
  props: { modelValue: { type: Boolean, default: false } },
  template: '<div data-test="delete-modal" />',
})

const base = {
  id: 'p1',
  displayName: 'PocketID',
  issuerUrl: 'https://id.example.org',
  clientId: 'client',
  usernameClaim: 'preferred_username',
  scopes: ['openid'],
  autoProvision: false,
  createdAt: new Date('2026-01-02T03:04:05.000Z'),
  updatedAt: new Date('2026-02-03T04:05:06.000Z'),
} satisfies OidcProvider

async function mount(): Promise<VueWrapper> {
  return mountSuspended(
    { render: () => h(VApp, () => [h(ProviderPage)]) },
    {
      global: {
        stubs: {
          OidcCardCategoryInformations: InformationsStub,
          OidcCardCategoryClaims: ClaimsStub,
          OidcCardCategoryUsers: UsersStub,
          OidcModalDelete: DeleteModalStub,
        },
      },
    },
  )
}

function deleteButton(wrapper: VueWrapper): ReturnType<VueWrapper['find']> {
  return wrapper.findAll('button').at(-1)!
}

beforeEach(() => {
  fetchProvider.mockReset()
  invalidate.mockReset()
  onBeforeRouteLeave.mockReset()
  provider.value = undefined
})

describe('settings/oidc/[id]', () => {
  it('fetches the provider on mount', async () => {
    await mount()
    await vi.waitFor(() => expect(fetchProvider).toHaveBeenCalledTimes(1))
  })

  it('drops the held provider when the route is left', async () => {
    await mount()
    expect(onBeforeRouteLeave).toHaveBeenCalledWith(invalidate)
  })

  it('renders every category card', async () => {
    const wrapper = await mount()

    expect(wrapper.find('[data-test="OidcCardCategoryInformations"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="OidcCardCategoryClaims"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="OidcCardCategoryUsers"]').exists()).toBe(true)
  })

  it('titles the page with the provider name', async () => {
    provider.value = { ...base }
    const wrapper = await mount()

    expect(wrapper.text()).toContain('PocketID')
  })

  it('falls back to a loading title until the provider is in', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).not.toContain('PocketID')
  })

  it('keeps delete out of reach until a provider is held', async () => {
    const wrapper = await mount()

    expect(deleteButton(wrapper).attributes('disabled')).toBeDefined()
    expect(wrapper.find('[data-test="delete-modal"]').exists()).toBe(false)
  })

  it('opens the delete modal from the delete button', async () => {
    provider.value = { ...base }
    const wrapper = await mount()

    expect(wrapper.findComponent(DeleteModalStub).props('modelValue')).toBe(false)
    await deleteButton(wrapper).trigger('click')

    await vi.waitFor(() => expect(wrapper.findComponent(DeleteModalStub).props('modelValue')).toBe(true))
  })
})
