// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { PageScale } from '~/features/reader/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it, vi } from 'vitest'
import { h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import PageScaleInput from './PageScale.vue'

function buttonByText(wrapper: VueWrapper, text: string): ReturnType<VueWrapper['find']> {
  return wrapper.findAll('button').find(b => b.text().includes(text))!
}

async function mount(
  scale: PageScale = 'fit-width',
  disabledOptions: PageScale[] = [],
): Promise<{
  wrapper: VueWrapper
  model: ReturnType<typeof ref<PageScale>>
  disabledOptions: ReturnType<typeof ref<PageScale[]>>
}> {
  const model = ref(scale)
  const disabledOptionsRef = ref(disabledOptions)
  const wrapper = await mountSuspended({
    setup: () => ({ model, disabledOptionsRef }),
    render: () => h(VApp, () => [
      h(PageScaleInput, {
        'modelValue': model.value,
        'disabledOptions': disabledOptionsRef.value,
        'onUpdate:modelValue': (value: PageScale) => {
          model.value = value
        },
      }),
    ]),
  })

  return { wrapper, model, disabledOptions: disabledOptionsRef }
}

describe('readerInputPageScale', () => {
  it('offers each scale', async () => {
    const { wrapper } = await mount()

    expect(wrapper.text()).toContain('Fit width')
    expect(wrapper.text()).toContain('Fit height')
    expect(wrapper.text()).toContain('Fit screen')
  })

  it('selects a scale on click', async () => {
    const { wrapper, model } = await mount('fit-width')

    await buttonByText(wrapper, 'Fit height').trigger('click')

    expect(model.value).toBe('fit-height')
  })

  it('disables listed options', async () => {
    const { wrapper } = await mount('fit-width', ['fit-height', 'fit-screen'])

    expect(buttonByText(wrapper, 'Fit height').attributes('disabled')).toBeDefined()
    expect(buttonByText(wrapper, 'Fit screen').attributes('disabled')).toBeDefined()
    expect(buttonByText(wrapper, 'Fit width').attributes('disabled')).toBeUndefined()
  })

  it('moves off a newly disabled current option', async () => {
    const { wrapper, model, disabledOptions } = await mount('fit-height')

    disabledOptions.value = ['fit-height', 'fit-screen']
    await wrapper.vm.$forceUpdate()
    await wrapper.vm.$nextTick()

    await vi.waitFor(() => expect(model.value).toBe('fit-width'))
  })
})
