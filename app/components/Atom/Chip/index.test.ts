// SPDX-License-Identifier: AGPL-3.0-or-later
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h, ref } from 'vue'
import { VApp, VChip } from 'vuetify/components'
import Chip from './index.vue'

type ChipProps = InstanceType<typeof Chip>['$props']

// Default mock: desktop (mobile = false)
function useDisplayMock(): { mobile: ReturnType<typeof ref<boolean>> } {
  return { mobile: ref(false) }
}

mockNuxtImport('useDisplay', () => useDisplayMock)

function wrap(props: ChipProps): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Chip, props)]) }
}

describe('atomChip', () => {
  it('renders the provided text', async () => {
    const wrapper = await mountSuspended(wrap({ text: 'Hello' }))
    expect(wrapper.findComponent(VChip).props('text')).toBe('Hello')
  })

  it('forwards the colour prop', async () => {
    const wrapper = await mountSuspended(wrap({ text: 'Test', color: 'primary' }))
    expect(wrapper.findComponent(VChip).props('color')).toBe('primary')
  })

  it('uses tonal variant by default', async () => {
    const wrapper = await mountSuspended(wrap({ text: 'Test' }))
    expect(wrapper.findComponent(VChip).props('variant')).toBe('tonal')
  })

  it('applies text-label-large class for default size on desktop', async () => {
    const wrapper = await mountSuspended(wrap({ text: 'Test', size: 'default' }))
    const chip = wrapper.findComponent(VChip)
    expect(chip.classes()).toContain('text-label-large')
  })

  it('applies text-label-small class for small size on desktop', async () => {
    const wrapper = await mountSuspended(wrap({ text: 'Test', size: 'small' }))
    const chip = wrapper.findComponent(VChip)
    expect(chip.classes()).toContain('text-label-small')
  })
})
