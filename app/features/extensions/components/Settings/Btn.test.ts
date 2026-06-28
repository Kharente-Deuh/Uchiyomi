// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VChip } from 'vuetify/components'
import Btn from './Btn.vue'

function wrap(props: { hasSettings: boolean, pkgName: string }): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Btn, props)]) }
}

describe('extensionsSettingsBtn', () => {
  it('links to the settings page when the extension has settings', async () => {
    const wrapper = await mountSuspended(wrap({ hasSettings: true, pkgName: 'p' }))
    expect(wrapper.find('a').attributes('href')).toBe('/browse/extensions/p/settings')
    expect(wrapper.findComponent(VChip).exists()).toBe(false)
  })

  it('shows a disabled chip and no link when there are no settings', async () => {
    const wrapper = await mountSuspended(wrap({ hasSettings: false, pkgName: 'p' }))
    expect(wrapper.find('a').exists()).toBe(false)
    expect(wrapper.findComponent(VChip).exists()).toBe(true)
  })
})
