// SPDX-License-Identifier: AGPL-3.0-or-later
import type { NavigationDrawerListItemProps } from './Item.vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import Item from './Item.vue'

// baseRoute differs from the default test router path ('/') so the link stays
// present in the default cases.
const base: NavigationDrawerListItemProps = {
  icon: 'fa6-solid:gear',
  title: 'Settings',
  to: '/settings',
  isActiveFn: () => false,
  baseRoute: '/somewhere-else',
}

function wrap(
  props: NavigationDrawerListItemProps,
): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Item, props)]) }
}

describe('navigationDrawerListItem', () => {
  it('renders the title', async () => {
    const wrapper = await mountSuspended(wrap(base))
    expect(wrapper.text()).toContain('Settings')
  })

  it('links to the target and is not marked active when inactive', async () => {
    const wrapper = await mountSuspended(wrap({ ...base, isActiveFn: () => false }))
    expect(wrapper.find('a').attributes('href')).toBe('/settings')
    expect(wrapper.find('.navigation-drawer-list-item').classes())
      .not
      .toContain('navigation-drawer-list-item--active')
  })

  it('marks the item active and keeps the link when isActiveFn => true', async () => {
    const wrapper = await mountSuspended(wrap({ ...base, isActiveFn: () => true }))
    expect(wrapper.find('.navigation-drawer-list-item').classes())
      .toContain('navigation-drawer-list-item--active')
    expect(wrapper.find('a').attributes('href')).toBe('/settings')
  })

  it('drops the link when baseRoute matches the current route', async () => {
    // The test router lands at '/login' (Nuxt redirects on app init).
    // Setting baseRoute to '/login' makes isBaseRoute true, so the link is dropped.
    const wrapper = await mountSuspended(wrap({ ...base, baseRoute: '/login' }))
    expect(wrapper.find('a').exists()).toBe(false)
  })
})
