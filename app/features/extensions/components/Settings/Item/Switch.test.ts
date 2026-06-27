// SPDX-License-Identifier: AGPL-3.0-or-later
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { h } from 'vue'
import { VApp, VSwitch } from 'vuetify/components'
import Switch from './Switch.vue'

function wrap(props: { modelValue?: boolean }): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Switch, props)]) }
}

describe('extensionsSettingsItemSwitch', () => {
  afterEach(() => vi.useRealTimers())

  it('forwards the model value to the switch', async () => {
    const wrapper = await mountSuspended(wrap({ modelValue: true }))
    expect(wrapper.findComponent(VSwitch).props('modelValue')).toBe(true)
  })

  it('emits the debounced value after 400ms', async () => {
    vi.useFakeTimers()
    const wrapper = await mountSuspended(wrap({ modelValue: false }))
    const cmp = wrapper.findComponent(Switch)
    ;(cmp.vm as any).$.setupState.internalModelValue = true
    await vi.advanceTimersByTimeAsync(400)
    expect(cmp.emitted('update:modelValue')!.at(-1)).toEqual([true])
  })
})
