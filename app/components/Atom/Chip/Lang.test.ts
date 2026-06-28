// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VChip } from 'vuetify/components'
import Lang from './Lang.vue'

type LangProps = InstanceType<typeof Lang>['$props']

function wrap(props: LangProps): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Lang, props)]) }
}

describe('atomChipLang', () => {
  it('renders the language code prefix (before the hyphen) for a specific language', async () => {
    const wrapper = await mountSuspended(wrap({ lang: 'en-US' }))
    const chip = wrapper.findComponent(VChip)
    expect(chip.props('text')).toBe('en')
  })

  it('renders the full lang value when there is no hyphen', async () => {
    const wrapper = await mountSuspended(wrap({ lang: 'fr' }))
    const chip = wrapper.findComponent(VChip)
    expect(chip.props('text')).toBe('fr')
  })

  it('uses the multiLanguages translation (not a language code) when lang is "all"', async () => {
    const wrapper = await mountSuspended(wrap({ lang: 'all' }))
    const chip = wrapper.findComponent(VChip)
    // i18n is active in the nuxt test env; assert the value differs from 'all'
    // and is not a bare language code
    const text = chip.props('text') as string
    expect(text).not.toBe('all')
    expect(text.length).toBeGreaterThan(1)
  })

  it('uses the secondary colour', async () => {
    const wrapper = await mountSuspended(wrap({ lang: 'en' }))
    expect(wrapper.findComponent(VChip).props('color')).toBe('secondary')
  })

  it('does not apply text-uppercase class when lang is "all"', async () => {
    const wrapper = await mountSuspended(wrap({ lang: 'all' }))
    // The wrapper span around AtomChip should not have text-uppercase
    expect(wrapper.html()).not.toContain('text-uppercase')
  })
})
