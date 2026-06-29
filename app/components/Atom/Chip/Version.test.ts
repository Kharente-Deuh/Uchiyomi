// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VChip } from 'vuetify/components'
import Version from './Version.vue'

type VersionProps = InstanceType<typeof Version>['$props']

function wrap(props: VersionProps): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(Version, props)]) }
}

describe('atomChipVersion', () => {
  it('renders the version string with a "v" prefix', async () => {
    const wrapper = await mountSuspended(wrap({ version: '1.2.3', hasUpdate: false }))
    expect(wrapper.findComponent(VChip).props('text')).toBe('v1.2.3')
  })

  it('uses the secondary colour when there is no update', async () => {
    const wrapper = await mountSuspended(wrap({ version: '1.0.0', hasUpdate: false }))
    expect(wrapper.findComponent(VChip).props('color')).toBe('secondary')
  })

  it('uses the warning colour when an update is available', async () => {
    const wrapper = await mountSuspended(wrap({ version: '1.0.0', hasUpdate: true }))
    expect(wrapper.findComponent(VChip).props('color')).toBe('warning')
  })
})
