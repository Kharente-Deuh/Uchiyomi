// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import EmptyState from './EmptyState.vue'

describe('emptyState', () => {
  it('renders the title and description', async () => {
    const wrapper = await mountSuspended(EmptyState, {
      props: { title: 'No series', description: 'Add one to get started' },
    })
    expect(wrapper.text()).toContain('No series')
    expect(wrapper.text()).toContain('Add one to get started')
  })

  it('renders the action slot', async () => {
    const wrapper = await mountSuspended(EmptyState, {
      slots: { action: () => 'CTA' },
    })
    expect(wrapper.text()).toContain('CTA')
  })
})
