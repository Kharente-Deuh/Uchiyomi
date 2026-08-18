// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { ComicStatus } from '../../types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp } from 'vuetify/components'
import Status from './Status.vue'

async function mount(status: ComicStatus): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Status, { status })]),
  })
}

describe('comicsChipStatus', () => {
  it('renders the status label', async () => {
    const wrapper = await mount('ongoing')
    expect(wrapper.text()).toContain('Ongoing')
  })

  it.each([
    ['ongoing', 'border-thin-primary'],
    ['completed', 'border-thin'],
    ['hiatus', 'border-thin-warning'],
    ['cancelled', 'border-thin-error'],
  ] as const)('uses %s border class %s', async (status, borderClass) => {
    const wrapper = await mount(status)
    expect(wrapper.find(`.${borderClass}`).exists()).toBe(true)
  })
})
