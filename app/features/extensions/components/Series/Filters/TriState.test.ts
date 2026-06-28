// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceFilterDto } from '#shared/dto/catalogue/source-filters.dto'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import TriState from './TriState.vue'

const filter: Extract<SourceFilterDto, { type: 'tristate' }> = {
  type: 'tristate',
  position: 0,
  name: 'Action',
  default: 'IGNORE',
}

describe('extensionsSeriesFiltersTriState', () => {
  it('cycles from IGNORE to INCLUDE on click', async () => {
    const wrapper = await mountSuspended(TriState, { props: { filter, modelValue: 'IGNORE' } })
    await wrapper.find('div').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['INCLUDE'])
  })

  it('cycles from INCLUDE to EXCLUDE on click', async () => {
    const wrapper = await mountSuspended(TriState, { props: { filter, modelValue: 'INCLUDE' } })
    await wrapper.find('div').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['EXCLUDE'])
  })

  it('cycles from EXCLUDE to IGNORE on click', async () => {
    const wrapper = await mountSuspended(TriState, { props: { filter, modelValue: 'EXCLUDE' } })
    await wrapper.find('div').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['IGNORE'])
  })
})
