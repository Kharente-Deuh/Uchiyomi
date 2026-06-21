// SPDX-License-Identifier: AGPL-3.0-or-later
import type { NavItem } from '~/composables/useNavigation'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import AppNavigation from './AppNavigation.vue'

const items: NavItem[] = [
  { key: 'home', to: '/', icon: 'fa6-solid:house', labelKey: 'nav.home' },
  { key: 'settings', to: '/settings', icon: 'fa6-solid:gear', labelKey: 'nav.settings' },
]

function wrapInVApp(variant: 'bottom' | 'rail'): { render: () => ReturnType<typeof h> } {
  return {
    render() {
      return h(VApp, () => [
        h(AppNavigation, {
          variant,
          items,
          active: '/',
        }),
      ])
    },
  }
}

describe('appNavigation', () => {
  it('renders one entry per item (bottom variant)', async () => {
    const wrapper = await mountSuspended(wrapInVApp('bottom'))
    expect(wrapper.find('[data-test="nav-home"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="nav-settings"]').exists()).toBe(true)
  })

  it('emits navigate with the item path on click (bottom variant)', async () => {
    const wrapper = await mountSuspended(wrapInVApp('bottom'))
    await wrapper.find('[data-test="nav-settings"]').trigger('click')
    const nav = wrapper.findComponent(AppNavigation)
    expect(nav.emitted('navigate')?.[0]).toEqual(['/settings'])
  })

  it('renders one entry per item (rail variant)', async () => {
    const wrapper = await mountSuspended(wrapInVApp('rail'))
    expect(wrapper.find('[data-test="nav-home"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="nav-settings"]').exists()).toBe(true)
  })
})
