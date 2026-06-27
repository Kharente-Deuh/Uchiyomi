// SPDX-License-Identifier: AGPL-3.0-or-later
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VChip } from 'vuetify/components'
import Desktop from './Desktop.vue'

type DesktopProps = InstanceType<typeof Desktop>['$props']

function wrap(
  props: DesktopProps,
  slot = 'BODY_SLOT',
): { render: () => ReturnType<typeof h> } {
  return {
    render: () => h(VApp, () => [h(Desktop, props, { default: () => slot })]),
  }
}

describe('settingsCardDesktop', () => {
  it('renders the title', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'Appearance' }))
    expect(wrapper.text()).toContain('Appearance')
  })

  it('renders the subtitle when provided', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'Appearance', subtitle: 'Personalize the look' }))
    expect(wrapper.text()).toContain('Personalize the look')
  })

  it('renders the default slot content', async () => {
    const wrapper = await mountSuspended(wrap({ title: 'Appearance' }))
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
    const wrapper = await mountSuspended(wrap({ title: 'Appearance' }))
    expect(wrapper.findComponent(VChip).exists()).toBe(false)
  })
})
