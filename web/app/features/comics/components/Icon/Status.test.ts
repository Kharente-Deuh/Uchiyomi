// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { ComicStatus } from '~/features/comics/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import Status from './Status.vue'

async function mount(status: ComicStatus, hasBackground = false): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Status, { status, withBackground: hasBackground })]),
  })
}

describe('asuraIconStatus', () => {
  it('renders an icon for an ongoing series', async () => {
    const wrapper = await mount('ongoing')
    expect(wrapper.find('.v-icon').exists()).toBe(true)
  })

  it('wraps the icon when withBackground is set', async () => {
    const wrapper = await mount('completed', true)
    expect(wrapper.find('.status-icon-box').exists()).toBe(true)
  })
})
