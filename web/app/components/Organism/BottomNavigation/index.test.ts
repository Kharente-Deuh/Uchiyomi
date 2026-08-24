// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

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

  it('does not stretch to 100% height (that swallows the PWA home indicator)', async () => {
    const wrapper = await mountSuspended(wrap())
    expect(wrapper.find('.nvagation-bottom').classes()).not.toContain('h-100')
  })

  it('does not bake the safe-area into both height and padding (that doubles the PWA inset)', async () => {
    const source = await import('./index.vue?raw').then(m => m.default as string)
    expect(source).toMatch(/padding-bottom:\s*env\(safe-area-inset-bottom/)
    expect(source).not.toMatch(/height:\s*var\(--bottom-navigation-height\)/)
  })
})
