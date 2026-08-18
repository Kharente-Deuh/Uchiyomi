// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VImg } from 'vuetify/components'
import ProjectLogo from './ProjectLogo.vue'

describe('atomProjectLogo', () => {
  it('renders a full-width logo', async () => {
    const wrapper = await mountSuspended({
      render: () => h(VApp, () => [h(ProjectLogo)]),
    })

    expect(wrapper.findComponent(VImg).exists()).toBe(true)
    expect(wrapper.findComponent(VImg).props('aspectRatio')).toBeUndefined()
  })

  it('uses a square aspect ratio when compact', async () => {
    const wrapper = await mountSuspended({
      render: () => h(VApp, () => [h(ProjectLogo, { compact: true })]),
    })

    expect(wrapper.findComponent(VImg).props('aspectRatio')).toBe('1/1')
  })
})
