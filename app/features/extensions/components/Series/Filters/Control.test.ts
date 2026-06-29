// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceFilterDto } from '#shared/dto/catalogue/source-filters.dto'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import Control from './Control.vue'

describe('extensionsSeriesFiltersControl', () => {
  it('renders a checkbox control for a checkbox filter', async () => {
    const filter: SourceFilterDto = { type: 'checkbox', position: 0, name: 'Completed', default: false }
    const wrapper = await mountSuspended(Control, { props: { filter, modelValue: false } })
    expect(wrapper.find('.v-checkbox').exists()).toBe(true)
  })

  it('renders a divider for a separator filter', async () => {
    const filter: SourceFilterDto = { type: 'separator', position: 1, name: '' }
    const wrapper = await mountSuspended(Control, { props: { filter } })
    expect(wrapper.find('.v-divider').exists()).toBe(true)
  })

  it('renders the name as a header for a header filter', async () => {
    const filter: SourceFilterDto = { type: 'header', position: 2, name: 'Section' }
    const wrapper = await mountSuspended(Control, { props: { filter } })
    expect(wrapper.text()).toContain('Section')
  })
})
