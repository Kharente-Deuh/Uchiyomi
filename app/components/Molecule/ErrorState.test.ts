// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import ErrorState from './ErrorState.vue'

describe('errorState', () => {
  it('renders the provided message', async () => {
    const wrapper = await mountSuspended(ErrorState, { props: { message: 'Boom' } })
    expect(wrapper.text()).toContain('Boom')
  })

  it('emits retry when the retry button is clicked', async () => {
    const wrapper = await mountSuspended(ErrorState)
    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted('retry')).toHaveLength(1)
  })
})
