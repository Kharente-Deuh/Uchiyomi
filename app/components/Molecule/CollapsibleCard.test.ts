// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VBtn } from 'vuetify/components'
import CollapsibleCard from './CollapsibleCard.vue'

type CardProps = InstanceType<typeof CollapsibleCard>['$props']

function wrap(
  props: CardProps,
  slot = 'SLOT_CONTENT',
): { render: () => ReturnType<typeof h> } {
  return {
    render: () => h(VApp, () => [h(CollapsibleCard, props, { default: () => slot })]),
  }
}

describe('moleculeCollapsibleCard', () => {
  it('renders the title', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'Filters' }))
    expect(wrapper.text()).toContain('Filters')
  })

  it('hides the slot content initially (collapsed)', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'Filters' }))
    // v-show keeps the element in the DOM but sets display:none
    // The collapsible div is the last child div inside the card wrapper
    const allDivs = wrapper.findAll('div')
    const collapsible = allDivs.find(d => d.attributes('style')?.includes('display'))
    expect(collapsible).toBeDefined()
    expect(collapsible!.attributes('style')).toContain('display: none')
  })

  it('reveals the slot content after clicking the toggle button', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'Filters' }))
    await wrapper.findComponent(VBtn).trigger('click')
    // After click, no div should have display:none
    const hiddenDiv = wrapper.findAll('div').find(d => d.attributes('style')?.includes('display: none'))
    expect(hiddenDiv).toBeUndefined()
  })

  it('collapses again on a second click', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'Filters' }))
    const btn = wrapper.findComponent(VBtn)
    await btn.trigger('click')
    await btn.trigger('click')
    const collapsible = wrapper.findAll('div').find(d => d.attributes('style')?.includes('display'))
    expect(collapsible!.attributes('style')).toContain('display: none')
  })

  it('renders slot content inside the collapsible area', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'Filters' }))
    expect(wrapper.html()).toContain('SLOT_CONTENT')
  })
})
