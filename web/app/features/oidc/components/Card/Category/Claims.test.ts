// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { OidcProvider } from '~/features/oidc/composables/oidc.api'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import Claims from './Claims.vue'

const { provider, update } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return { provider: ref<OidcProvider>(), update: vi.fn() }
})

vi.mock('~/features/oidc/composables/oidc-provider.composable', () => ({
  useOidcProvider: () => ({ provider, update, updateLoading: ref(false) }),
}))

const base = {
  id: 'p1',
  displayName: 'PocketID',
  issuerUrl: 'https://id.example.org',
  clientId: 'client',
  usernameClaim: 'preferred_username',
  scopes: ['openid', 'email'],
  roleClaim: 'groups',
  adminValues: ['admins'],
  allowedValues: ['users'],
  autoProvision: false,
  createdAt: new Date('2026-01-02T03:04:05.000Z'),
  updatedAt: new Date('2026-02-03T04:05:06.000Z'),
} satisfies OidcProvider

async function mount(): Promise<VueWrapper> {
  return mountSuspended({ render: () => h(VApp, () => [h(Claims, { loading: false })]) })
}

function inputs(wrapper: VueWrapper): HTMLInputElement[] {
  return wrapper.findAll('input').map(i => i.element as HTMLInputElement)
}

function saveButton(wrapper: VueWrapper): ReturnType<VueWrapper['find']> {
  return wrapper.findAll('button').at(-1)!
}

beforeEach(() => {
  update.mockReset()
  provider.value = { ...base, scopes: [...base.scopes], adminValues: [...base.adminValues], allowedValues: [...base.allowedValues] }
})

describe('oidcCardCategoryClaims', () => {
  it('mounts with a reactive provider held in the store', async () => {
    await expect(mount()).resolves.toBeTruthy()
  })

  it('fills the username claim from the provider', async () => {
    const wrapper = await mount()
    expect(inputs(wrapper).some(i => i.value === 'preferred_username')).toBe(true)
  })

  it('fills the role claim from the provider', async () => {
    const wrapper = await mount()
    expect(inputs(wrapper).some(i => i.value === 'groups')).toBe(true)
  })

  it('renders the scopes as chips', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('openid')
    expect(wrapper.text()).toContain('email')
  })

  it('renders the role value lists as chips', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('admins')
    expect(wrapper.text()).toContain('users')
  })

  it('keeps save disabled while nothing changed', async () => {
    const wrapper = await mount()
    expect(saveButton(wrapper).attributes('disabled')).toBeDefined()
  })

  it('enables save once a field changed', async () => {
    const wrapper = await mount()
    const usernameClaim = wrapper.findAll('input').find(i => (i.element as HTMLInputElement).value === 'preferred_username')!

    await usernameClaim.setValue('email')

    await vi.waitFor(() => expect(saveButton(wrapper).attributes('disabled')).toBeUndefined())
  })

  it('submits the provider merged with the edited values', async () => {
    const wrapper = await mount()
    const usernameClaim = wrapper.findAll('input').find(i => (i.element as HTMLInputElement).value === 'preferred_username')!

    await usernameClaim.setValue('email')
    await vi.waitFor(() => expect(saveButton(wrapper).attributes('disabled')).toBeUndefined())
    await saveButton(wrapper).trigger('click')

    await vi.waitFor(() => expect(update).toHaveBeenCalledWith(expect.objectContaining({
      id: 'p1',
      usernameClaim: 'email',
      scopes: ['openid', 'email'],
    })))
  })

  it('clears both value lists when the role claim is emptied', async () => {
    const wrapper = await mount()
    const roleClaim = wrapper.findAll('input').find(i => (i.element as HTMLInputElement).value === 'groups')!

    await roleClaim.setValue('')

    await vi.waitFor(() => {
      expect(wrapper.text()).not.toContain('admins')
      expect(wrapper.text()).not.toContain('users')
    })
  })

  it('does not submit while the username claim is empty', async () => {
    const wrapper = await mount()
    const usernameClaim = wrapper.findAll('input').find(i => (i.element as HTMLInputElement).value === 'preferred_username')!

    await usernameClaim.setValue('')
    await vi.waitFor(() => expect(saveButton(wrapper).attributes('disabled')).toBeDefined())

    expect(update).not.toHaveBeenCalled()
  })
})
