// SPDX-License-Identifier: AGPL-3.0-or-later
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import ConfirmationCard from './Card.vue'

function wrap(
  props: { text: string, loading?: boolean },
  slot = '',
): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(ConfirmationCard, props, { default: () => slot })]) }
}

describe('modalConfirmationCard', () => {
  it('renders the confirmation text and the default slot', async () => {
    const wrapper = await mountSuspended(wrap({ text: 'Are you sure?' }, 'EXTRA'))
    expect(wrapper.text()).toContain('Are you sure?')
    expect(wrapper.text()).toContain('EXTRA')
  })

  it('emits cancel from the cancel button', async () => {
    const wrapper = await mountSuspended(wrap({ text: 'x' }))
    await wrapper.findAll('button')[0]!.trigger('click')
    expect(wrapper.findComponent(ConfirmationCard).emitted('cancel')).toBeTruthy()
  })

  it('emits confirm from the continue button', async () => {
    const wrapper = await mountSuspended(wrap({ text: 'x' }))
    await wrapper.findAll('button')[1]!.trigger('click')
    expect(wrapper.findComponent(ConfirmationCard).emitted('confirm')).toBeTruthy()
  })
})
