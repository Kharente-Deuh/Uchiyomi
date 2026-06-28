// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VChip } from 'vuetify/components'
import Mobile from './Mobile.vue'

type MobileProps = InstanceType<typeof Mobile>['$props']

function wrap(
  props: MobileProps,
  slot = 'BODY_SLOT',
): { render: () => ReturnType<typeof h> } {
  return {
    render: () => h(VApp, () => [h(Mobile, props, { default: () => slot })]),
  }
}

describe('settingsCardMobile', () => {
  it('renders the title', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'Privacy' }))
    expect(wrapper.text()).toContain('Privacy')
  })

  it('renders the default slot content', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'Privacy' }))
    expect(wrapper.text()).toContain('BODY_SLOT')
  })

  it('shows the Admin chip when adminOnly is set', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'Users', adminOnly: true }))
    const chip = wrapper.findComponent(VChip)
    expect(chip.exists()).toBe(true)
    // i18n is active in the nuxt test env; the chip text is a non-empty translated string
    expect((chip.props('text') as string).length).toBeGreaterThan(0)
  })

  it('does not show the Admin chip when adminOnly is absent', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'Privacy' }))
    expect(wrapper.findComponent(VChip).exists()).toBe(false)
  })
})
