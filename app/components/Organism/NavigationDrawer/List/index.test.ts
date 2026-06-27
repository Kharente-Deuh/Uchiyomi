// SPDX-License-Identifier: AGPL-3.0-or-later
import type { NavigationDrawerListProps } from './index.vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import List from './index.vue'

const props: NavigationDrawerListProps = {
  title: 'SYSTEM',
  items: [
    { icon: 'fa6-solid:gear', title: 'Settings', to: '/settings', isActiveFn: () => false, baseRoute: '/somewhere-else' },
    { icon: 'fa6-solid:house', title: 'Home', to: '/', isActiveFn: () => false, baseRoute: '/somewhere-else' },
  ],
}

function wrap(p: NavigationDrawerListProps): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(List, p)]) }
}

describe('navigationDrawerList', () => {
  it('renders the section title', async () => {
    const wrapper = await mountSuspended(wrap(props))
    expect(wrapper.text()).toContain('SYSTEM')
  })

  it('renders one entry per item', async () => {
    const wrapper = await mountSuspended(wrap(props))
    expect(wrapper.findAll('.navigation-drawer-list-item')).toHaveLength(2)
    expect(wrapper.text()).toContain('Settings')
    expect(wrapper.text()).toContain('Home')
  })

  it('renders no entries for an empty list', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'READING', items: [] }))
    expect(wrapper.findAll('.navigation-drawer-list-item')).toHaveLength(0)
    expect(wrapper.text()).toContain('READING')
  })
})
