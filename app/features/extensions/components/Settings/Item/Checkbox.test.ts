// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { h } from 'vue'
import { VApp, VCheckbox } from 'vuetify/components'
import Checkbox from './Checkbox.vue'

function wrap(props: { modelValue?: boolean }): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Checkbox, props)]) }
}

describe('extensionsSettingsItemCheckbox', () => {
  afterEach(() => vi.useRealTimers())

  it('forwards the model value to the checkbox', async () => {
    const wrapper = await mountSuspended(wrap({ modelValue: true }))
    expect(wrapper.findComponent(VCheckbox).props('modelValue')).toBe(true)
  })

  it('defaults the checkbox to false when no value is given', async () => {
    const wrapper = await mountSuspended(wrap({}))
    expect(wrapper.findComponent(VCheckbox).props('modelValue')).toBe(false)
  })

  it('emits the debounced value after 400ms', async () => {
    vi.useFakeTimers()
    const wrapper = await mountSuspended(wrap({ modelValue: false }))
    const cmp = wrapper.findComponent(Checkbox)
    ;(cmp.vm as any).$.setupState.internalModelValue = true
    await vi.advanceTimersByTimeAsync(400)
    expect(cmp.emitted('update:modelValue')!.at(-1)).toEqual([true])
  })
})
