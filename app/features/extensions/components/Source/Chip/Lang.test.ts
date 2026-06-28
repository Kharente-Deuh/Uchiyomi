// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import Lang from './Lang.vue'

function wrap(props: { lang: string, enabled?: boolean }): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Lang, props)]) }
}

describe('extensionsSourceChipLang', () => {
  it('displays the first segment of a BCP-47 lang code (CSS text-uppercase handles visual casing)', async () => {
    const wrapper = await mountSuspended(wrap({ lang: 'en-US' }))
    // The component uses CSS text-uppercase; the DOM text is lowercase
    expect(wrapper.text()).toBe('en')
  })

  it('displays * for the all language', async () => {
    const wrapper = await mountSuspended(wrap({ lang: 'all' }))
    expect(wrapper.text()).toBe('*')
  })

  it('applies the active class when enabled is true', async () => {
    const wrapper = await mountSuspended(wrap({ lang: 'fr', enabled: true }))
    expect(wrapper.find('.source-lang-code').classes()).toContain('source-lang-code--active')
  })

  it('does not apply the active class when enabled is false', async () => {
    const wrapper = await mountSuspended(wrap({ lang: 'fr', enabled: false }))
    expect(wrapper.find('.source-lang-code').classes()).not.toContain('source-lang-code--active')
  })

  it('uses text-title-large for the all language', async () => {
    const wrapper = await mountSuspended(wrap({ lang: 'all' }))
    expect(wrapper.find('.text-title-large').exists()).toBe(true)
  })

  it('uses text-label-medium for a regular language code', async () => {
    const wrapper = await mountSuspended(wrap({ lang: 'ja' }))
    expect(wrapper.find('.text-label-medium').exists()).toBe(true)
  })
})
