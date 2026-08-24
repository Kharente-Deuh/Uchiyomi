// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { OidcProviderDetails } from '~/features/oidc/composables/oidc.api'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import Users from './Users.vue'

const { provider } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return { provider: ref<OidcProviderDetails>() }
})

vi.mock('~/features/oidc/composables/oidc-provider.composable', () => ({
  useOidcProvider: () => ({ provider }),
}))

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
  users: [
    { id: 'u1', username: 'alice', isAdmin: true, linkedAt: new Date('2026-01-10T00:00:00.000Z') },
    { id: 'u2', username: 'bob', isAdmin: false, linkedAt: new Date('2026-02-15T00:00:00.000Z') },
  ],
} satisfies OidcProviderDetails

async function mount(isLoading = false): Promise<VueWrapper> {
  return mountSuspended({ render: () => h(VApp, () => [h(Users, { loading: isLoading })]) })
}

beforeEach(() => {
  provider.value = { ...base, users: base.users.map(u => ({ ...u })) }
})

describe('oidcCardCategoryUsers', () => {
  it('lists the username of every linked user', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('bob')
  })

  it('labels an admin user with the admin role', async () => {
    const wrapper = await mount()
    const row = wrapper.findAll('.provider-users-grid').find(r => r.text().includes('alice'))!

    expect(row.text()).toContain('Admin')
  })

  it('labels a non-admin user with the regular user role', async () => {
    const wrapper = await mount()
    const row = wrapper.findAll('.provider-users-grid').find(r => r.text().includes('bob'))!

    expect(row.text()).toContain('User')
  })

  it('formats the linked date for the locale', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).toContain(base.users[0]!.linkedAt.toLocaleDateString('en'))
  })

  it('includes the user count in the title', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).toContain('Users (2)')
  })

  it('renders a no-users message when nothing is linked', async () => {
    provider.value = { ...base, users: [] }
    const wrapper = await mount()

    expect(wrapper.text()).toContain('No users are linked to this provider')
  })

  it('hides the no-users message while loading', async () => {
    provider.value = { ...base, users: [] }
    const wrapper = await mount(true)

    expect(wrapper.text()).not.toContain('No users are linked to this provider')
  })
})
