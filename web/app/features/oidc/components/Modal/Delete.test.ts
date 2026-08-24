// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { OidcProvider } from '~/features/oidc/composables/oidc.api'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import Delete from './Delete.vue'

const { provider, deleteProvider, deleteLoading } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return { provider: ref<OidcProvider>(), deleteProvider: vi.fn(), deleteLoading: ref(false) }
})

vi.mock('~/features/oidc/composables/oidc-provider.composable', () => ({
  useOidcProvider: () => ({ provider, deleteProvider, deleteLoading }),
}))

const ConfirmationStub = defineComponent({
  name: 'OrganismModalConfirmation',
  props: {
    modelValue: { type: Boolean, default: false },
    text: { type: String, default: '' },
    loading: { type: Boolean, default: false },
    confirmText: { type: String, default: '' },
  },
  emits: ['confirm'],
  template: '<button data-test="confirm" @click="$emit(\'confirm\')" />',
})

const base = {
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
} satisfies OidcProvider

async function mount(show = ref(true)): Promise<VueWrapper> {
  return mountSuspended(
    {
      render: () => h(VApp, () => [h(Delete, {
        'modelValue': show.value,
        'onUpdate:modelValue': (isShown: boolean) => {
          show.value = isShown
        },
      })]),
    },
    { global: { stubs: { OrganismModalConfirmation: ConfirmationStub } } },
  )
}

beforeEach(() => {
  deleteProvider.mockReset()
  deleteLoading.value = false
  provider.value = { ...base }
})

describe('oidcModalDelete', () => {
  it('names the provider in the confirmation text', async () => {
    const wrapper = await mount()
    expect(wrapper.findComponent(ConfirmationStub).props('text')).toContain('PocketID')
  })

  it('renders a confirmation text even without a held provider', async () => {
    provider.value = undefined
    const wrapper = await mount()

    expect(wrapper.findComponent(ConfirmationStub).props('text')).not.toContain('undefined')
  })

  it('deletes the provider on confirmation', async () => {
    const wrapper = await mount()

    await wrapper.find('[data-test="confirm"]').trigger('click')

    expect(deleteProvider).toHaveBeenCalled()
  })

  it('forwards the delete loading state', async () => {
    deleteLoading.value = true
    const wrapper = await mount()

    expect(wrapper.findComponent(ConfirmationStub).props('loading')).toBe(true)
  })
})
