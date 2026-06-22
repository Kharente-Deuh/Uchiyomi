// SPDX-License-Identifier: AGPL-3.0-or-later
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import Card from './index.vue'

type CardProps = InstanceType<typeof Card>['$props']

function wrap(
  props: CardProps,
  slot = 'CARD_BODY',
): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Card, props, { default: () => slot })]) }
}

describe('settingsCard', () => {
  it('renders the title, subtitle and body slot', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'Appearance', subtitle: 'Personalize' }))
    expect(wrapper.text()).toContain('Appearance')
    expect(wrapper.text()).toContain('Personalize')
    expect(wrapper.text()).toContain('CARD_BODY')
  })

  it('shows the Admin chip only when adminOnly is set', async () => {
    const plain = await mountSuspended(wrap({ title: 'Appearance' }))
    expect(plain.find('.v-chip').exists()).toBe(false)

    const admin = await mountSuspended(wrap({ title: 'Users', adminOnly: true }))
    const chip = admin.find('.v-chip')
    expect(chip.exists()).toBe(true)
    expect(chip.text()).toContain('Admin')
  })
})
