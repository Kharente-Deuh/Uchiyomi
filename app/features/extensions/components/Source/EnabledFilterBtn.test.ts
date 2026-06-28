// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VBtn } from 'vuetify/components'
import EnabledFilterBtn from './EnabledFilterBtn.vue'

function wrap(modelValue?: boolean): { render: () => ReturnType<typeof h> } {
  return {
    render: () => h(VApp, () => [h(EnabledFilterBtn, { modelValue, 'onUpdate:modelValue': () => {} })]),
  }
}

describe('extensionsSourceEnabledFilterBtn', () => {
  it('uses secondary color when modelValue is undefined', async () => {
    const wrapper = await mountSuspended(wrap())
    expect(wrapper.findComponent(VBtn).props('color')).toBe('secondary')
  })

  it('uses primary color when modelValue is true', async () => {
    const wrapper = await mountSuspended(wrap(true))
    expect(wrapper.findComponent(VBtn).props('color')).toBe('primary')
  })

  it('uses error color when modelValue is false', async () => {
    const wrapper = await mountSuspended(wrap(false))
    expect(wrapper.findComponent(VBtn).props('color')).toBe('error')
  })

  it('cycles undefined → true on click', async () => {
    const emitted: Array<boolean | undefined> = []
    const wrapper = await mountSuspended({
      render: () => h(VApp, () => [
        h(EnabledFilterBtn, {
          'modelValue': undefined,
          'onUpdate:modelValue': (v: boolean | undefined) => emitted.push(v),
        }),
      ]),
    })
    await wrapper.findComponent(VBtn).trigger('click')
    expect(emitted[0]).toBe(true)
  })

  it('cycles true → false on click', async () => {
    const emitted: Array<boolean | undefined> = []
    const wrapper = await mountSuspended({
      render: () => h(VApp, () => [
        h(EnabledFilterBtn, {
          'modelValue': true,
          'onUpdate:modelValue': (v: boolean | undefined) => emitted.push(v),
        }),
      ]),
    })
    await wrapper.findComponent(VBtn).trigger('click')
    expect(emitted[0]).toBe(false)
  })

  it('cycles false → undefined on click', async () => {
    const emitted: Array<boolean | undefined> = []
    const wrapper = await mountSuspended({
      render: () => h(VApp, () => [
        h(EnabledFilterBtn, {
          'modelValue': false,
          'onUpdate:modelValue': (v: boolean | undefined) => emitted.push(v),
        }),
      ]),
    })
    await wrapper.findComponent(VBtn).trigger('click')
    expect(emitted[0]).toBeUndefined()
  })
})
