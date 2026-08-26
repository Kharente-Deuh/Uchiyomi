// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VImg } from 'vuetify/components'
import { ASURA_SOURCE_NAME } from '~/constants'
import Source from './Source.vue'

async function mount(size?: 'default' | 'small'): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Source, { source: ASURA_SOURCE_NAME, size })]),
  })
}

describe('comicsChipSource', () => {
  it('renders the source name by default', async () => {
    const wrapper = await mount()

    expect(wrapper.text()).toContain('Asura Scans')
    expect(wrapper.findComponent(VImg).exists()).toBe(true)
  })

  it('renders a compact image without the name', async () => {
    const wrapper = await mount('small')

    expect(wrapper.text()).not.toContain('Asura Scans')
    expect(wrapper.find('img').exists()).toBe(true)
  })
})
