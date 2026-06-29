// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { h } from 'vue'
import { VApp, VSelect } from 'vuetify/components'
import List from './List.vue'

interface ListProps { modelValue?: string, entries: string[], entryValues: string[] }

function wrap(props: ListProps): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(List, props)]) }
}

const base: ListProps = { entries: ['English', 'French'], entryValues: ['en', 'fr'] }

describe('extensionsSettingsItemList', () => {
  afterEach(() => vi.useRealTimers())

  it('maps entries and entryValues into select items', async () => {
    const wrapper = await mountSuspended(wrap({ ...base, modelValue: 'en' }))
    expect(wrapper.findComponent(VSelect).props('items')).toEqual([
      { value: 'en', title: 'English' },
      { value: 'fr', title: 'French' },
    ])
  })

  it('emits the debounced selection after 400ms', async () => {
    vi.useFakeTimers()
    const wrapper = await mountSuspended(wrap({ ...base, modelValue: 'en' }))
    const cmp = wrapper.findComponent(List)
    ;(cmp.vm as any).$.setupState.internalModelValue = 'fr'
    await vi.advanceTimersByTimeAsync(400)
    expect(cmp.emitted('update:modelValue')!.at(-1)).toEqual(['fr'])
  })

  it('emits undefined when the selection is cleared', async () => {
    vi.useFakeTimers()
    const wrapper = await mountSuspended(wrap({ ...base, modelValue: 'en' }))
    const cmp = wrapper.findComponent(List)
    ;(cmp.vm as any).$.setupState.internalModelValue = null
    await vi.advanceTimersByTimeAsync(400)
    expect(cmp.emitted('update:modelValue')!.at(-1)).toEqual([undefined])
  })
})
