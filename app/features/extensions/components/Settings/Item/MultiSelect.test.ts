// SPDX-License-Identifier: AGPL-3.0-or-later
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { h } from 'vue'
import { VApp, VChip } from 'vuetify/components'
import MultiSelect from './MultiSelect.vue'

interface MultiSelectProps { modelValue?: string[], entries: string[], entryValues: string[] }

function wrap(props: MultiSelectProps): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(MultiSelect, props)]) }
}

const base: MultiSelectProps = { entries: ['English', 'French'], entryValues: ['en', 'fr'] }

describe('extensionsSettingsItemMultiSelect', () => {
  afterEach(() => vi.useRealTimers())

  it('renders a chip per entry', async () => {
    const wrapper = await mountSuspended(wrap({ ...base }))
    const chips = wrapper.findAllComponents(VChip)
    expect(chips).toHaveLength(2)
    expect(chips.map(c => c.props('text'))).toEqual(['English', 'French'])
  })

  it('marks the pre-selected chips with the primary colour', async () => {
    const wrapper = await mountSuspended(wrap({ ...base, modelValue: ['fr'] }))
    const chips = wrapper.findAllComponents(VChip)
    expect(chips[0]!.props('color')).toBe('secondary')
    expect(chips[1]!.props('color')).toBe('primary')
  })

  it('toggles a value on chip click and emits the debounced array', async () => {
    vi.useFakeTimers()
    const wrapper = await mountSuspended(wrap({ ...base }))
    const cmp = wrapper.findComponent(MultiSelect)
    await wrapper.findAllComponents(VChip)[0]!.trigger('click')
    await vi.advanceTimersByTimeAsync(400)
    expect(cmp.emitted('update:modelValue')!.at(-1)).toEqual([['en']])
  })
})
