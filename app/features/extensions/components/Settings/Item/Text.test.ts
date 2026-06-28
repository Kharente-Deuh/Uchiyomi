// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { h } from 'vue'
import { VApp, VTextField } from 'vuetify/components'
import Text from './Text.vue'

function wrap(props: { modelValue?: string }): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Text, props)]) }
}

describe('extensionsSettingsItemText', () => {
  afterEach(() => vi.useRealTimers())

  it('forwards the model value to the text field', async () => {
    const wrapper = await mountSuspended(wrap({ modelValue: 'token' }))
    expect(wrapper.findComponent(VTextField).props('modelValue')).toBe('token')
  })

  it('emits the debounced value after 400ms', async () => {
    vi.useFakeTimers()
    const wrapper = await mountSuspended(wrap({ modelValue: 'a' }))
    const cmp = wrapper.findComponent(Text)
    ;(cmp.vm as any).$.setupState.internalModelValue = 'b'
    await vi.advanceTimersByTimeAsync(400)
    expect(cmp.emitted('update:modelValue')!.at(-1)).toEqual(['b'])
  })
})
