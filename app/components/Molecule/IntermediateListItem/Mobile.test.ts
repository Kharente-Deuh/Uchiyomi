// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import Mobile from './Mobile.vue'

describe('moleculeIntermediateListItemMobile', () => {
  it('renders the title', async () => {
    const wrapper = await mountSuspended(Mobile, {
      props: { title: 'Security' },
    })
    expect(wrapper.text()).toContain('Security')
  })

  it('renders the subtitle when provided', async () => {
    const wrapper = await mountSuspended(Mobile, {
      props: { title: 'Security', subtitle: 'Password and 2FA' },
    })
    expect(wrapper.text()).toContain('Password and 2FA')
  })

  it('applies the bottom border class when last is true', async () => {
    const wrapper = await mountSuspended(Mobile, {
      props: { title: 'Security', last: true },
    })
    expect(wrapper.find('div').classes()).toContain('border-b-thin')
  })

  it('does not apply the bottom border class when last is false', async () => {
    const wrapper = await mountSuspended(Mobile, {
      props: { title: 'Security', last: false },
    })
    expect(wrapper.find('div').classes()).not.toContain('border-b-thin')
  })
})
