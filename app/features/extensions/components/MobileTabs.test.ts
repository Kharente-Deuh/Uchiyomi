// SPDX-License-Identifier: AGPL-3.0-or-later

import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h } from 'vue'
import { VApp, VBtn } from 'vuetify/components'
import MobileTabs from './MobileTabs.vue'

type MobileModel = 'sources' | 'series'

function wrap(modelValue: MobileModel): { render: () => ReturnType<typeof h> } {
  return { render: () => h(VApp, () => [h(MobileTabs, { modelValue, 'onUpdate:modelValue': () => {} })]) }
}

describe('extensionsMobileTabs', () => {
  it('renders two buttons', async () => {
    const wrapper = await mountSuspended(wrap('sources'))
    expect(wrapper.findAllComponents(VBtn)).toHaveLength(2)
  })

  it('gives the sources button a flat primary style when modelValue is sources', async () => {
    const wrapper = await mountSuspended(wrap('sources'))
    const [sourcesBtn] = wrapper.findAllComponents(VBtn)
    expect(sourcesBtn!.props('color')).toBe('primary')
    expect(sourcesBtn!.props('variant')).toBe('flat')
  })

  it('gives the series button a text secondary style when modelValue is sources', async () => {
    const wrapper = await mountSuspended(wrap('sources'))
    const [, seriesBtn] = wrapper.findAllComponents(VBtn)
    expect(seriesBtn!.props('color')).toBe('secondary')
    expect(seriesBtn!.props('variant')).toBe('text')
  })

  it('emits update:modelValue with series when series button is clicked', async () => {
    const emitted: MobileModel[] = []
    const wrapper = await mountSuspended({
      render: () => h(VApp, () => [
        h(MobileTabs, {
          'modelValue': 'sources',
          'onUpdate:modelValue': (v: MobileModel) => emitted.push(v),
        }),
      ]),
    })
    const [, seriesBtn] = wrapper.findAllComponents(VBtn)
    await seriesBtn!.trigger('click')
    expect(emitted).toContain('series')
  })

  it('emits update:modelValue with sources when sources button is clicked', async () => {
    const emitted: MobileModel[] = []
    const wrapper = await mountSuspended({
      render: () => h(VApp, () => [
        h(MobileTabs, {
          'modelValue': 'series',
          'onUpdate:modelValue': (v: MobileModel) => emitted.push(v),
        }),
      ]),
    })
    const [sourcesBtn] = wrapper.findAllComponents(VBtn)
    await sourcesBtn!.trigger('click')
    expect(emitted).toContain('sources')
  })
})
