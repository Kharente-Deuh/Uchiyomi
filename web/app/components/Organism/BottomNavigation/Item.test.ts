// SPDX-License-Identifier: AGPL-3.0-or-later

import type { BottomNavigationItemProps } from './Item.vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VIcon } from 'vuetify/components'
import Item from './Item.vue'

// baseRoute differs from the default test router path ('/') so the link stays
// present in the default cases.
const base: BottomNavigationItemProps = {
  icon: 'fa6-solid:house',
  to: '/',
  isActiveFn: () => false,
  baseRoute: '/somewhere-else',
}

function wrap(props: BottomNavigationItemProps): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Item, props)]) }
}

describe('bottomNavigationItem', () => {
  it('renders the icon', async () => {
    const wrapper = await mountSuspended(wrap(base))
    expect(wrapper.findComponent(VIcon).props('icon')).toBe('fa6-solid:house')
  })

  it('links to the target and uses the secondary colour when inactive', async () => {
    const wrapper = await mountSuspended(wrap({ ...base, to: '/settings', isActiveFn: () => false }))
    expect(wrapper.find('a').attributes('href')).toBe('/settings')
    expect(wrapper.findComponent(VIcon).props('color')).toBe('secondary')
  })

  it('keeps the link and uses the primary colour when active (isActiveFn => true)', async () => {
    const wrapper = await mountSuspended(wrap({ ...base, to: '/settings', isActiveFn: () => true }))
    expect(wrapper.find('a').attributes('href')).toBe('/settings')
    expect(wrapper.findComponent(VIcon).props('color')).toBe('primary')
  })

  it('drops the link when baseRoute matches the current route', async () => {
    // The test router lands at '/login' (Nuxt redirects on app init).
    // Setting baseRoute to '/login' makes isBaseRoute true, so the link is dropped.
    const wrapper = await mountSuspended(wrap({ ...base, to: '/settings', baseRoute: '/login' }))
    expect(wrapper.find('a').exists()).toBe(false)
  })
})
