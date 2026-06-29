// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceFilterDto } from '#shared/dto/catalogue/source-filters.dto'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { VBtn, VSelect } from 'vuetify/components'
import Sort from './Sort.vue'

const filter: Extract<SourceFilterDto, { type: 'sort' }> = {
  type: 'sort',
  position: 0,
  name: 'Sort',
  default: null,
  values: ['Title', 'Date'],
}

describe('extensionsSeriesFiltersSort', () => {
  it('toggles ascending to false when direction button is clicked', async () => {
    const wrapper = await mountSuspended(Sort, { props: { filter, modelValue: { ascending: true, index: 0 } } })
    await wrapper.findComponent(VBtn).trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([{ ascending: false, index: 0 }])
  })

  it('updates the index via VSelect while preserving ascending', async () => {
    const wrapper = await mountSuspended(Sort, { props: { filter, modelValue: { ascending: true, index: 0 } } })
    wrapper.findComponent(VSelect).vm.$emit('update:model-value', 1)
    await wrapper.vm.$nextTick()
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([{ ascending: true, index: 1 }])
  })

  it('falls back to default state when modelValue is undefined and toggles ascending', async () => {
    const wrapper = await mountSuspended(Sort, { props: { filter, modelValue: undefined } })
    await wrapper.findComponent(VBtn).trigger('click')
    // state defaults to { ascending: false, index: 0 }, toggle → { ascending: true, index: 0 }
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([{ ascending: true, index: 0 }])
  })
})
