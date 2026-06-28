// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VChip } from 'vuetify/components'
import Filter from './Filter.vue'

type FilterProps = InstanceType<typeof Filter>['$props']

function wrap(props: FilterProps): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Filter, props)]) }
}

describe('atomFilter', () => {
  it('uses the secondary colour and neutral border when the value is undefined', async () => {
    const wrapper = await mountSuspended(wrap({ label: 'NSFW', modelValue: undefined }))
    const chip = wrapper.findComponent(VChip)
    expect(chip.props('color')).toBe('secondary')
    expect(chip.classes()).toContain('border-thin-secondary')
  })

  it('uses the primary colour and border when the value is true', async () => {
    const wrapper = await mountSuspended(wrap({ label: 'NSFW', modelValue: true }))
    const chip = wrapper.findComponent(VChip)
    expect(chip.props('color')).toBe('primary')
    expect(chip.classes()).toContain('border-thin-primary')
  })

  it('uses the error colour and border when the value is false', async () => {
    const wrapper = await mountSuspended(wrap({ label: 'NSFW', modelValue: false }))
    const chip = wrapper.findComponent(VChip)
    expect(chip.props('color')).toBe('error')
    expect(chip.classes()).toContain('border-thin-error')
  })

  it('cycles undefined → true on click', async () => {
    const wrapper = await mountSuspended(wrap({ label: 'NSFW', modelValue: undefined }))
    await wrapper.findComponent(VChip).trigger('click')
    expect(wrapper.findComponent(Filter).emitted('update:modelValue')![0]).toEqual([true])
  })

  it('cycles true → false on click', async () => {
    const wrapper = await mountSuspended(wrap({ label: 'NSFW', modelValue: true }))
    await wrapper.findComponent(VChip).trigger('click')
    expect(wrapper.findComponent(Filter).emitted('update:modelValue')![0]).toEqual([false])
  })

  it('cycles false → undefined on click', async () => {
    const wrapper = await mountSuspended(wrap({ label: 'NSFW', modelValue: false }))
    await wrapper.findComponent(VChip).trigger('click')
    expect(wrapper.findComponent(Filter).emitted('update:modelValue')![0]).toEqual([undefined])
  })

  it('renders the label and the prepend icon', async () => {
    const wrapper = await mountSuspended(wrap({ label: 'NSFW', modelValue: undefined, icon: 'fa6-solid:ban' }))
    const chip = wrapper.findComponent(VChip)
    expect(chip.props('text')).toBe('NSFW')
    expect(chip.props('prependIcon')).toBe('fa6-solid:ban')
  })

  it('forwards the disabled prop', async () => {
    const wrapper = await mountSuspended(wrap({ label: 'NSFW', modelValue: undefined, disabled: true }))
    expect(wrapper.findComponent(VChip).props('disabled')).toBe(true)
  })
})
