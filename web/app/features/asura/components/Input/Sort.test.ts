// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { AsuraSort } from '../../types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VSelect } from 'vuetify/components'
import Sort from './Sort.vue'

async function mount(sort: AsuraSort = 'popular', isDisabled = false): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Sort, { modelValue: sort, disabled: isDisabled })]),
  })
}

describe('asuraInputSort', () => {
  it('offers popular, latest, title and rating', async () => {
    const wrapper = await mount()
    const items = wrapper.findComponent(VSelect).props('items') as { value: string, title: string }[]

    expect(items.map(i => i.value)).toEqual(['popular', 'latest', 'title', 'rating'])
  })

  it('reflects the current sort', async () => {
    const wrapper = await mount('title')
    expect(wrapper.findComponent(VSelect).props('modelValue')).toBe('title')
  })

  it('disables the select while loading', async () => {
    const wrapper = await mount('popular', true)
    expect(wrapper.findComponent(VSelect).props('disabled')).toBe(true)
  })
})
