// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { ComicSource } from '~/features/comics/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VSelect } from 'vuetify/components'
import { ASURA_SOURCE_NAME, KING_OF_SHOJO_SOURCE_NAME } from '~/constants'
import Source from './Source.vue'

async function mount(source?: ComicSource, isDisabled = false): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Source, { modelValue: source, disabled: isDisabled })]),
  })
}

describe('libraryInputSource', () => {
  it('offers asurascans', async () => {
    const wrapper = await mount()
    const items = wrapper.findComponent(VSelect).props('items') as { value: string, title: string }[]

    expect(items.map(i => i.value)).toEqual([ASURA_SOURCE_NAME, KING_OF_SHOJO_SOURCE_NAME])
  })

  it('is clearable', async () => {
    const wrapper = await mount(ASURA_SOURCE_NAME)
    expect(wrapper.findComponent(VSelect).props('clearable')).toBe(true)
  })

  it('disables the select while loading', async () => {
    const wrapper = await mount(undefined, true)
    expect(wrapper.findComponent(VSelect).props('disabled')).toBe(true)
  })
})
