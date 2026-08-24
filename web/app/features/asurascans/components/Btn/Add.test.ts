// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import Add from './Add.vue'

async function mount(isLoading = false): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Add, { loading: isLoading })]),
  })
}

describe('asuraScansBtnAdd', () => {
  it('shows the book icon while idle', async () => {
    const wrapper = await mount()
    expect(wrapper.find('.v-icon').exists()).toBe(true)
  })

  it('shows a spinner while loading', async () => {
    const wrapper = await mount(true)
    expect(wrapper.find('.v-progress-circular').isVisible()).toBe(true)
  })
})
