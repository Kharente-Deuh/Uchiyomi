// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ModalCardProps } from './Card.vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import ModalCard from './Card.vue'

function wrap(
  props: ModalCardProps = {},
  slots: Record<string, () => unknown> = {},
): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(ModalCard, props, slots)]) }
}

describe('modalCard', () => {
  it('renders the title and emits close from the close icon', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'My Modal' }))
    expect(wrapper.text()).toContain('My Modal')
    await wrapper.find('.modal-card-close-btn').trigger('click')
    expect(wrapper.findComponent(ModalCard).emitted('close')).toBeTruthy()
  })

  it('emits submit when the form is submitted', async () => {
    const wrapper = await mountSuspended(wrap())
    await wrapper.find('form').trigger('submit.prevent')
    expect(wrapper.findComponent(ModalCard).emitted('submit')).toBeTruthy()
  })

  it('emits cancel from the cancel button', async () => {
    const wrapper = await mountSuspended(wrap())
    await wrapper.findAll('button')[0]!.trigger('click')
    expect(wrapper.findComponent(ModalCard).emitted('cancel')).toBeTruthy()
  })

  it('renders the default and actions slots', async () => {
    const wrapper = await mountSuspended(wrap({}, {
      default: () => 'BODY_SLOT',
      actions: () => 'ACTIONS_SLOT',
    }))
    expect(wrapper.text()).toContain('BODY_SLOT')
    expect(wrapper.text()).toContain('ACTIONS_SLOT')
  })

  it('disables the submit button while the form is incomplete', async () => {
    const wrapper = await mountSuspended(wrap({ isFormComplete: false }))
    expect(wrapper.find('button[type="submit"]').attributes('disabled')).toBeDefined()
  })

  it('hides the action bar when noActions is set', async () => {
    const wrapper = await mountSuspended(wrap({ noActions: true }))
    expect(wrapper.find('button[type="submit"]').exists()).toBe(false)
  })
})
