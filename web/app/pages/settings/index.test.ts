// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import SettingsPage from './index.vue'

const { isAdmin } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return { isAdmin: ref(false) }
})

vi.mock('~/features/auth/composables/auth.composable', () => ({
  useAuth: () => ({ isAdmin }),
}))

async function mount(): Promise<VueWrapper> {
  return mountSuspended({ render: () => h(VApp, () => [h(SettingsPage)]) })
}

beforeEach(() => {
  isAdmin.value = false
})

describe('settings', () => {
  it('links a regular user to the reader module', async () => {
    const wrapper = await mount()

    expect(wrapper.findAll('a').map(a => a.attributes('href'))).toEqual(['/settings/reader'])
    expect(wrapper.text()).toContain('Reader')
  })

  it('hides the OIDC module from a regular user', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).not.toContain('OIDC')
    expect(wrapper.findAll('a').map(a => a.attributes('href'))).not.toContain('/settings/oidc')
  })

  it('links an admin to the reader and OIDC modules', async () => {
    isAdmin.value = true
    const wrapper = await mount()

    expect(wrapper.findAll('a').map(a => a.attributes('href'))).toEqual([
      '/settings/reader',
      '/settings/oidc',
    ])
  })

  it('describes the OIDC module', async () => {
    isAdmin.value = true
    const wrapper = await mount()

    expect(wrapper.text()).toContain('OIDC')
    expect(wrapper.text()).toContain('Manage OIDC providers')
  })

  it('follows the admin flag without a remount', async () => {
    const wrapper = await mount()

    isAdmin.value = true

    await vi.waitFor(() => expect(wrapper.findAll('a')).toHaveLength(2))
  })
})
