// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import Password from './Password.vue'

function wrap(attrs: Record<string, unknown> = {}): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Password, attrs)]) }
}

describe('atomInputPassword', () => {
  it('renders a masked password input by default', async () => {
    const wrapper = await mountSuspended(wrap())
    expect(wrapper.find('input').attributes('type')).toBe('password')
  })

  it('toggles between masked and visible when the append icon is clicked', async () => {
    const wrapper = await mountSuspended(wrap())
    await wrapper.find('.v-icon').trigger('click')
    expect(wrapper.find('input').attributes('type')).toBe('text')
    await wrapper.find('.v-icon').trigger('click')
    expect(wrapper.find('input').attributes('type')).toBe('password')
  })

  it('forwards attrs (e.g. label) to the inner text field', async () => {
    const wrapper = await mountSuspended(wrap({ label: 'Mot de passe' }))
    expect(wrapper.html()).toContain('Mot de passe')
  })
})
