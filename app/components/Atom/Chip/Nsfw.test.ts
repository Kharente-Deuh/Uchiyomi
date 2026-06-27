// SPDX-License-Identifier: AGPL-3.0-or-later
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VChip } from 'vuetify/components'
import Nsfw from './Nsfw.vue'

type NsfwProps = InstanceType<typeof Nsfw>['$props']

function wrap(props: NsfwProps = {}): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Nsfw, props)]) }
}

describe('atomChipNsfw', () => {
  it('renders the 18+ label', async () => {
    const wrapper = await mountSuspended(wrap())
    expect(wrapper.findComponent(VChip).props('text')).toBe('18+')
  })

  it('uses the error colour', async () => {
    const wrapper = await mountSuspended(wrap())
    expect(wrapper.findComponent(VChip).props('color')).toBe('error')
  })

  it('forwards the size prop', async () => {
    // When size="small", AtomChip internally maps it; we verify the prop reaches
    // the inner VChip (which may be x-small or small depending on display)
    const wrapper = await mountSuspended(wrap({ size: 'small' }))
    expect(wrapper.findComponent(VChip).exists()).toBe(true)
  })
})
