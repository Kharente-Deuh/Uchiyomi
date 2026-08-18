// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { ComicType } from '../../types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import Type from './Type.vue'

async function mount(type: ComicType): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Type, { type })]),
  })
}

describe('comicsChipType', () => {
  it('renders the type label', async () => {
    const wrapper = await mount('manhwa')
    expect(wrapper.text()).toContain('Manhwa')
  })
})
