// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import LoadingState from './LoadingState.vue'

describe('loadingState', () => {
  it('renders the provided message', async () => {
    const wrapper = await mountSuspended(LoadingState, { props: { message: 'Fetching series' } })
    expect(wrapper.text()).toContain('Fetching series')
  })
})
