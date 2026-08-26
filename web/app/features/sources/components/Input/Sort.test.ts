// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { SourceSort } from '../../types'
import type { ComicSource } from '~/features/comics/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VSelect } from 'vuetify/components'
import { ASURA_SOURCE_NAME, KING_OF_SHOJO_SOURCE_NAME } from '~/constants'
import Sort from './Sort.vue'

async function mount(
  source: ComicSource = ASURA_SOURCE_NAME,
  sort: SourceSort = 'popular',
  isDisabled = false,
): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Sort, { source, modelValue: sort, disabled: isDisabled })]),
  })
}

describe('sourcesInputSort', () => {
  it('offers asurascans sorts including rating', async () => {
    const wrapper = await mount()
    const items = wrapper.findComponent(VSelect).props('items') as { value: string, title: string }[]

    expect(items.map(i => i.value)).toEqual(['popular', 'latest', 'rating', 'title', 'newest'])
  })

  it('omits rating for kingofshojo', async () => {
    const wrapper = await mount(KING_OF_SHOJO_SOURCE_NAME)
    const items = wrapper.findComponent(VSelect).props('items') as { value: string, title: string }[]

    expect(items.map(i => i.value)).toEqual(['popular', 'latest', 'title', 'newest'])
  })

  it('reflects the current sort', async () => {
    const wrapper = await mount(ASURA_SOURCE_NAME, 'latest')
    expect(wrapper.findComponent(VSelect).props('modelValue')).toBe('latest')
  })

  it('disables the select while loading', async () => {
    const wrapper = await mount(ASURA_SOURCE_NAME, 'popular', true)
    expect(wrapper.findComponent(VSelect).props('disabled')).toBe(true)
  })
})
