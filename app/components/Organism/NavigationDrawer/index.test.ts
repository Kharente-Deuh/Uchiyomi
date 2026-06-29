// SPDX-License-Identifier: AGPL-3.0-or-later

import type { NavigationDrawerListProps } from './List/index.vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h, nextTick } from 'vue'
import { VApp } from 'vuetify/components'
import NavigationDrawer from './index.vue'

const items: NavigationDrawerListProps[] = [
  {
    title: 'SYSTEM',
    items: [
      { icon: 'fa6-solid:gear', title: 'Settings', to: '/settings', isActiveFn: () => false, baseRoute: '/somewhere-else' },
    ],
  },
]

function wrap(): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(NavigationDrawer, { items })]) }
}

describe('navigationDrawer', () => {
  it('renders each section and its items', async () => {
    const wrapper = await mountSuspended(wrap())
    expect(wrapper.text()).toContain('SYSTEM')
    expect(wrapper.text()).toContain('Settings')
  })

  it('emits logout when the sign-out button is clicked', async () => {
    const wrapper = await mountSuspended(wrap())
    const logout = wrapper.find('[data-test="user-menu-logout"]')
    expect(logout.exists()).toBe(true)

    await logout.trigger('click')
    await nextTick()

    expect(wrapper.findComponent(NavigationDrawer).emitted('logout')).toBeTruthy()
  })
})
