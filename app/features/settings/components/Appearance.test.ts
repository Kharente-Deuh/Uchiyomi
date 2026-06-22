// SPDX-License-Identifier: AGPL-3.0-or-later
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import Appearance from './Appearance.vue'

function wrap(): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Appearance)]) }
}

describe('settingsAppearance', () => {
  it('renders the appearance card with a theme and a language control', async () => {
    const wrapper = await mountSuspended(wrap())
    expect(wrapper.text()).toContain('Appearance')
    expect(wrapper.findAll('.v-select')).toHaveLength(2)
  })
})
