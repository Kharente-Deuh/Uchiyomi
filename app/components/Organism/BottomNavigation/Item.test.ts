// SPDX-License-Identifier: AGPL-3.0-or-later
import type { BottomNavigationItemProps } from './Item.vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VIcon } from 'vuetify/components'
import Item from './Item.vue'

const base: BottomNavigationItemProps = {
  icon: 'fa6-solid:house',
  to: '/',
  isActiveFn: () => false,
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

  it('drops the link and uses the primary colour when active', async () => {
    const wrapper = await mountSuspended(wrap({ ...base, to: '/settings', isActiveFn: () => true }))
    expect(wrapper.find('a').exists()).toBe(false)
    expect(wrapper.findComponent(VIcon).props('color')).toBe('primary')
  })
})
