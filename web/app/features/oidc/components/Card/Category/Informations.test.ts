// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { OidcProvider } from '~/features/oidc/composables/oidc.api'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import Informations from './Informations.vue'

const { provider, update } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return { provider: ref<OidcProvider>(), update: vi.fn() }
})

vi.mock('~/features/oidc/composables/oidc-provider.composable', () => ({
  useOidcProvider: () => ({ provider, update, updateLoading: ref(false) }),
}))

const TestBtnStub = defineComponent({
  name: 'OidcBtnTest',
  props: { issuerUrl: { type: String, default: '' } },
  template: '<div data-test="test-btn-stub" />',
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

async function mount(isLoading = false): Promise<VueWrapper> {
  return mountSuspended(
    { render: () => h(VApp, () => [h(Informations, { loading: isLoading })]) },
    { global: { stubs: { OidcBtnTest: TestBtnStub } } },
  )
}

function inputByValue(wrapper: VueWrapper, value: string): ReturnType<VueWrapper['find']> {
  return wrapper.findAll('input').find(i => (i.element as HTMLInputElement).value === value)!
}

function saveButton(wrapper: VueWrapper): ReturnType<VueWrapper['find']> {
  return wrapper.findAll('button').at(-1)!
}

beforeEach(() => {
  update.mockReset()
  provider.value = { ...base, scopes: [...base.scopes] }
})

describe('oidcCardCategoryInformations', () => {
  it('fills the fields from the provider', async () => {
    const wrapper = await mount()
    const values = wrapper.findAll('input').map(i => (i.element as HTMLInputElement).value)

    expect(values).toContain('PocketID')
    expect(values).toContain('pocket-id')
    expect(values).toContain('client')
    expect(values).toContain('https://id.example.org')
  })

  it('reflects the auto-provision flag', async () => {
    provider.value = { ...base, autoProvision: true }
    const wrapper = await mount()

    expect(wrapper.find('input[type="checkbox"]').element).toHaveProperty('checked', true)
  })

  it('hands the issuer url over to the test button', async () => {
    const wrapper = await mount()
    expect(wrapper.findComponent(TestBtnStub).props('issuerUrl')).toBe('https://id.example.org')
  })

  it('keeps save disabled while nothing changed', async () => {
    const wrapper = await mount()
    expect(saveButton(wrapper).attributes('disabled')).toBeDefined()
  })

  it('enables save once a field changed', async () => {
    const wrapper = await mount()

    await inputByValue(wrapper, 'PocketID').setValue('Renamed')

    await vi.waitFor(() => expect(saveButton(wrapper).attributes('disabled')).toBeUndefined())
  })

  it('submits the provider merged with the edited values', async () => {
    const wrapper = await mount()

    await inputByValue(wrapper, 'PocketID').setValue('Renamed')
    await vi.waitFor(() => expect(saveButton(wrapper).attributes('disabled')).toBeUndefined())
    await saveButton(wrapper).trigger('click')

    await vi.waitFor(() => expect(update).toHaveBeenCalledWith(expect.objectContaining({
      id: 'p1',
      displayName: 'Renamed',
      issuerUrl: 'https://id.example.org',
      scopes: ['openid'],
    })))
  })

  it('keeps save disabled for a value that is not a url', async () => {
    const wrapper = await mount()

    await inputByValue(wrapper, 'https://id.example.org').setValue('not-a-url')

    await vi.waitFor(() => expect(saveButton(wrapper).attributes('disabled')).toBeDefined())
    expect(update).not.toHaveBeenCalled()
  })

  it('keeps save disabled while the display name is empty', async () => {
    const wrapper = await mount()

    await inputByValue(wrapper, 'PocketID').setValue('')

    await vi.waitFor(() => expect(saveButton(wrapper).attributes('disabled')).toBeDefined())
  })

  it('refills the fields when the provider lands after the mount', async () => {
    provider.value = undefined
    const wrapper = await mount()

    provider.value = { ...base, displayName: 'Keycloak' }

    await vi.waitFor(() => expect(
      wrapper.findAll('input').map(i => (i.element as HTMLInputElement).value),
    ).toContain('Keycloak'))
  })

  it('submits nothing while no provider is held', async () => {
    provider.value = undefined
    const wrapper = await mount()

    await saveButton(wrapper).trigger('click')

    expect(update).not.toHaveBeenCalled()
  })

  it('enables save once the slug changed', async () => {
    const wrapper = await mount()

    await inputByValue(wrapper, 'pocket-id').setValue('new-slug')

    await vi.waitFor(() => expect(saveButton(wrapper).attributes('disabled')).toBeUndefined())
  })

  it('submits the edited slug to update', async () => {
    const wrapper = await mount()

    await inputByValue(wrapper, 'pocket-id').setValue('new-slug')
    await vi.waitFor(() => expect(saveButton(wrapper).attributes('disabled')).toBeUndefined())
    await saveButton(wrapper).trigger('click')

    await vi.waitFor(() => expect(update).toHaveBeenCalledWith(expect.objectContaining({
      id: 'p1',
      slug: 'new-slug',
    })))
  })
})
