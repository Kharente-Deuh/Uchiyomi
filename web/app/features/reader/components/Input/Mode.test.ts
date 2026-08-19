// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { ReadingMode } from '~/features/reader/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import Mode from './Mode.vue'

function buttonByText(wrapper: VueWrapper, text: string): ReturnType<VueWrapper['find']> {
  return wrapper.findAll('button').find(b => b.text().includes(text))!
}

async function mount(mode: ReadingMode = 'webtoon', isDisabled = false): Promise<{
  wrapper: VueWrapper
  model: ReturnType<typeof ref<ReadingMode>>
}> {
  const model = ref(mode)
  const wrapper = await mountSuspended({
    setup: () => ({ model, isDisabled }),
    render: () => h(VApp, () => [
      h(Mode, {
        'modelValue': model.value,
        'disabled': isDisabled,
        'onUpdate:modelValue': (value: ReadingMode) => {
          model.value = value
        },
      }),
    ]),
  })

  return { wrapper, model }
}

describe('readerInputMode', () => {
  it('offers each reading mode', async () => {
    const { wrapper } = await mount()

    expect(wrapper.text()).toContain('Left to right')
    expect(wrapper.text()).toContain('Right to left')
    expect(wrapper.text()).toContain('Continous scroll')
  })

  it('selects a mode on click', async () => {
    const { wrapper, model } = await mount('webtoon')

    await buttonByText(wrapper, 'Left to right').trigger('click')

    expect(model.value).toBe('paged-ltr')
  })

  it('disables every option while loading', async () => {
    const { wrapper } = await mount('webtoon', true)

    expect(wrapper.findAll('button').every(b => b.attributes('disabled') !== undefined)).toBe(true)
  })
})
