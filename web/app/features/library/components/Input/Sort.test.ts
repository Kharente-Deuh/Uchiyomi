// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { SearchComicSort } from '~/features/comics/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VSelect } from 'vuetify/components'
import Sort from './Sort.vue'

async function mount(sort: SearchComicSort = 'title', disabled = false): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Sort, { modelValue: sort, disabled })]),
  })
}

describe('libraryInputSort', () => {
  it('offers title and latest', async () => {
    const wrapper = await mount()
    const items = wrapper.findComponent(VSelect).props('items') as { value: string, title: string }[]

    expect(items.map(i => i.value)).toEqual(['title', 'latest'])
  })

  it('reflects the current sort', async () => {
    const wrapper = await mount('latest')
    expect(wrapper.findComponent(VSelect).props('modelValue')).toBe('latest')
  })

  it('disables the select while loading', async () => {
    const wrapper = await mount('title', true)
    expect(wrapper.findComponent(VSelect).props('disabled')).toBe(true)
  })
})
