// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { PageMode } from './PageMode.vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it, vi } from 'vitest'
import { h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import PageModeInput from './PageMode.vue'

function buttonByText(wrapper: VueWrapper, text: string): ReturnType<VueWrapper['find']> {
  return wrapper.findAll('button').find(b => b.text().includes(text))!
}

async function mount(
  mode: PageMode = 'single',
  disabledOptions: PageMode[] = [],
): Promise<{
  wrapper: VueWrapper
  model: ReturnType<typeof ref<PageMode>>
  disabledOptions: ReturnType<typeof ref<PageMode[]>>
}> {
  const model = ref(mode)
  const disabledOptionsRef = ref(disabledOptions)
  const wrapper = await mountSuspended({
    setup: () => ({ model, disabledOptionsRef }),
    render: () => h(VApp, () => [
      h(PageModeInput, {
        'modelValue': model.value,
        'disabledOptions': disabledOptionsRef.value,
        'onUpdate:modelValue': (value: PageMode) => {
          model.value = value
        },
      }),
    ]),
  })

  return { wrapper, model, disabledOptions: disabledOptionsRef }
}

describe('readerInputPageMode', () => {
  it('offers single and double page', async () => {
    const { wrapper } = await mount()

    expect(wrapper.text()).toContain('Single page')
    expect(wrapper.text()).toContain('Double page')
  })

  it('selects a mode on click', async () => {
    const { wrapper, model } = await mount('single')

    await buttonByText(wrapper, 'Double page').trigger('click')

    expect(model.value).toBe('double')
  })

  it('disables listed options', async () => {
    const { wrapper } = await mount('single', ['double'])

    expect(buttonByText(wrapper, 'Double page').attributes('disabled')).toBeDefined()
    expect(buttonByText(wrapper, 'Single page').attributes('disabled')).toBeUndefined()
  })

  it('moves off a newly disabled current option', async () => {
    const { wrapper, model, disabledOptions } = await mount('double')

    disabledOptions.value = ['double']
    await wrapper.vm.$forceUpdate()
    await wrapper.vm.$nextTick()

    await vi.waitFor(() => expect(model.value).toBe('single'))
  })
})
