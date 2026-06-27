// SPDX-License-Identifier: AGPL-3.0-or-later
import type { BottomNavigationItemProps } from './Item.vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import BottomNavigation from './index.vue'
import Item from './Item.vue'

const items: BottomNavigationItemProps[] = [
  { icon: 'fa6-solid:house', to: '/', isActiveFn: () => false, baseRoute: '/somewhere-else' },
  { icon: 'fa6-solid:gear', to: '/settings', isActiveFn: () => false, baseRoute: '/somewhere-else' },
]

function wrap(): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(BottomNavigation, { items })]) }
}

describe('bottomNavigation', () => {
  it('renders one item per entry', async () => {
    const wrapper = await mountSuspended(wrap())
    expect(wrapper.findAllComponents(Item)).toHaveLength(2)
  })
})
