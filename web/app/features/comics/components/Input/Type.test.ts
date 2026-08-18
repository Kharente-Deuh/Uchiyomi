// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { ComicType } from '~/features/comics/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VSelect } from 'vuetify/components'
import Type from './Type.vue'

async function mount(type?: ComicType, isDisabled = false): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Type, { modelValue: type, disabled: isDisabled })]),
  })
}

describe('comicsInputType', () => {
  it('offers each comic type', async () => {
    const wrapper = await mount()
    const items = wrapper.findComponent(VSelect).props('items') as { value: string, title: string }[]

    expect(items.map(i => i.value)).toEqual(['manhwa', 'mangatoon', 'manga', 'manhua'])
  })

  it('is clearable', async () => {
    const wrapper = await mount('manhwa')
    expect(wrapper.findComponent(VSelect).props('clearable')).toBe(true)
  })

  it('disables the select while loading', async () => {
    const wrapper = await mount(undefined, true)
    expect(wrapper.findComponent(VSelect).props('disabled')).toBe(true)
  })
})
