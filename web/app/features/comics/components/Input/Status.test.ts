// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { ComicStatus } from '~/features/comics/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VSelect } from 'vuetify/components'
import Status from './Status.vue'

async function mount(status?: ComicStatus, isDisabled = false): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Status, { modelValue: status, disabled: isDisabled })]),
  })
}

describe('comicsInputStatus', () => {
  it('offers each comic status', async () => {
    const wrapper = await mount()
    const items = wrapper.findComponent(VSelect).props('items') as { value: string, title: string }[]

    expect(items.map(i => i.value)).toEqual(['ongoing', 'completed', 'hiatus', 'cancelled'])
  })

  it('is clearable', async () => {
    const wrapper = await mount('ongoing')
    expect(wrapper.findComponent(VSelect).props('clearable')).toBe(true)
  })

  it('disables the select while loading', async () => {
    const wrapper = await mount(undefined, true)
    expect(wrapper.findComponent(VSelect).props('disabled')).toBe(true)
  })
})
