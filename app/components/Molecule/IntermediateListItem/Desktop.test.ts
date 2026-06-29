// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import Desktop from './Desktop.vue'

describe('moleculeIntermediateListItemDesktop', () => {
  it('renders the title', async () => {
    const wrapper = await mountSuspended(Desktop, {
      props: { title: 'Account' },
    })
    expect(wrapper.text()).toContain('Account')
  })

  it('renders the subtitle when provided', async () => {
    const wrapper = await mountSuspended(Desktop, {
      props: { title: 'Account', subtitle: 'Manage your profile' },
    })
    expect(wrapper.text()).toContain('Manage your profile')
  })

  it('renders an icon element when the icon prop is set', async () => {
    const wrapper = await mountSuspended(Desktop, {
      props: { title: 'Account', icon: 'fa6-solid:user' },
    })
    expect(wrapper.find('.v-icon').exists()).toBe(true)
  })
})
