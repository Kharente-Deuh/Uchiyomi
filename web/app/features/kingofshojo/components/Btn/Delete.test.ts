// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import Delete from './Delete.vue'

async function mount(mode: 'btn' | 'label'): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Delete, { mode })]),
  })
}

describe('asuraScansBtnDelete', () => {
  it('uses the in-library class in label mode', async () => {
    const wrapper = await mount('label')
    expect(wrapper.find('.remove-library-btn--label').exists()).toBe(true)
    expect(wrapper.text()).toContain('In library')
  })

  it('uses the remove class in btn mode', async () => {
    const wrapper = await mount('btn')
    expect(wrapper.find('.remove-library-btn--btn').exists()).toBe(true)
  })
})
