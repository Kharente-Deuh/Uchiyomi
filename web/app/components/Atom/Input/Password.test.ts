// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import Password from './Password.vue'

function wrap(attrs: Record<string, unknown> = {}): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Password, attrs)]) }
}

describe('atomInputPassword', () => {
  it('masks the value by default', async () => {
    const wrapper = await mountSuspended(wrap())
    expect(wrapper.find('input').attributes('type')).toBe('password')
  })

  it('reveals the value once the toggle is clicked', async () => {
    const wrapper = await mountSuspended(wrap())

    await wrapper.find('.v-field__append-inner .v-icon').trigger('click')

    expect(wrapper.find('input').attributes('type')).toBe('text')
  })

  it('masks the value again on a second click', async () => {
    const wrapper = await mountSuspended(wrap())
    const toggle = wrapper.find('.v-field__append-inner .v-icon')

    await toggle.trigger('click')
    await toggle.trigger('click')

    expect(wrapper.find('input').attributes('type')).toBe('password')
  })

  it('forwards attributes to the underlying field', async () => {
    const wrapper = await mountSuspended(wrap({ label: 'Client secret' }))
    expect(wrapper.text()).toContain('Client secret')
  })
})
