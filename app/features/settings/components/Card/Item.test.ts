// SPDX-License-Identifier: AGPL-3.0-or-later

import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h, ref } from 'vue'
import Item from './Item.vue'

// The card is responsive: the subtitle lives in the desktop branch only. Force
// the desktop layout so these tests cover it deterministically (the test env
// otherwise resolves `mobile` to true).
function useDisplayMock(): { mobile: ReturnType<typeof ref<boolean>> } {
  return { mobile: ref(false) }
}

mockNuxtImport('useDisplay', () => useDisplayMock)

type ItemProps = InstanceType<typeof Item>['$props']

function wrap(
  props: ItemProps,
  slot = 'CONTROL_SLOT',
): { render: () => ReturnType<typeof h> } {
  return { render: () => h(Item, props, { default: () => slot }) }
}

describe('settingsCardItem', () => {
  it('renders the title and the control slot', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'Theme' }))
    expect(wrapper.text()).toContain('Theme')
    expect(wrapper.text()).toContain('CONTROL_SLOT')
  })

  it('renders the subtitle when provided', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'Theme', subtitle: 'Light or dark' }))
    expect(wrapper.text()).toContain('Light or dark')
  })

  it('omits the subtitle node when not provided', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'Theme' }))
    expect(wrapper.findAll('span')).toHaveLength(1)
  })
})
