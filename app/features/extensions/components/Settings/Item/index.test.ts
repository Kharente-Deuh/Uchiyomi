// SPDX-License-Identifier: AGPL-3.0-or-later

import type { PreferenceDto } from '#shared/dto/extensions'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import Checkbox from './Checkbox.vue'
import Item from './index.vue'
import List from './List.vue'
import MultiSelect from './MultiSelect.vue'
import Switch from './Switch.vue'
import Text from './Text.vue'

function wrap(item: PreferenceDto): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Item, { item })]) }
}

const items = {
  checkbox: { position: 0, type: 'checkbox', visible: true, booleanValue: true, booleanDefault: false },
  switch: { position: 0, type: 'switch', visible: true, booleanValue: true, booleanDefault: false },
  editText: { position: 0, type: 'editText', visible: true, textValue: 'x' },
  list: { position: 0, type: 'list', visible: true, entries: ['A'], entryValues: ['a'], textValue: 'a' },
  multiSelect: { position: 0, type: 'multiSelect', visible: true, entries: ['A'], entryValues: ['a'], multiValue: [] },
} satisfies Record<string, PreferenceDto>

describe('extensionsSettingsItem', () => {
  it('renders the matching control for each preference type', async () => {
    const checkbox = await mountSuspended(wrap(items.checkbox))
    expect(checkbox.findComponent(Checkbox).exists()).toBe(true)
    const switchEl = await mountSuspended(wrap(items.switch))
    expect(switchEl.findComponent(Switch).exists()).toBe(true)
    const text = await mountSuspended(wrap(items.editText))
    expect(text.findComponent(Text).exists()).toBe(true)
    const list = await mountSuspended(wrap(items.list))
    expect(list.findComponent(List).exists()).toBe(true)
    const multiSelect = await mountSuspended(wrap(items.multiSelect))
    expect(multiSelect.findComponent(MultiSelect).exists()).toBe(true)
  })

  it('forwards the control update event', async () => {
    const wrapper = await mountSuspended(wrap(items.checkbox))
    wrapper.findComponent(Checkbox).vm.$emit('update:model-value', false)
    expect(wrapper.findComponent(Item).emitted('update:model-value')!.at(-1)).toEqual([false])
  })
})
