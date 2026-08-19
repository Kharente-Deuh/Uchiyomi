// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { ComicType } from '~/features/comics/types'
import type { ReaderSettings } from '~/features/reader/types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { h, ref } from 'vue'
import { VApp } from 'vuetify/components'
import { useToast } from '~/composables/toast.composable'
import Card from './Card.vue'

const { updateReaderSettings } = await vi.hoisted(async () => ({
  updateReaderSettings: vi.fn(),
}))

function readerSettingsApiStub(): { updateReaderSettings: typeof updateReaderSettings } {
  return { updateReaderSettings }
}

mockNuxtImport('createReaderSettingsApi', () => readerSettingsApiStub)

function buttonByText(wrapper: VueWrapper, text: string): ReturnType<VueWrapper['find']> | undefined {
  return wrapper.findAll('button').find(b => b.text().includes(text))
}

function manhwaDefaults(): ReaderSettings {
  return {
    type: 'manhwa',
    readingMode: 'webtoon',
    pageScale: 'fit-width',
    doublePage: false,
  }
}

async function mount(settings: ReaderSettings = manhwaDefaults()): Promise<{
  wrapper: VueWrapper
  model: ReturnType<typeof ref<ReaderSettings>>
  type: ReturnType<typeof ref<ComicType>>
}> {
  const model = ref({ ...settings })
  const type = ref(settings.type)
  const wrapper = await mountSuspended({
    setup: () => ({ model, type }),
    render: () => h(VApp, () => [
      h(Card, {
        'modelValue': model.value,
        'type': type.value,
        'onUpdate:modelValue': (value: ReaderSettings) => {
          model.value = value
        },
        'onUpdate:type': (value: ComicType) => {
          type.value = value
        },
      }),
    ]),
  })

  return { wrapper, model, type }
}

beforeEach(() => {
  updateReaderSettings.mockReset()
  useToast().messages.value.length = 0
})

describe('readerCard', () => {
  it('keeps save disabled while nothing changed', async () => {
    const { wrapper } = await mount()

    expect(buttonByText(wrapper, 'Save')?.attributes('disabled')).toBeDefined()
  })

  it('hides reset when the saved settings already match the type defaults', async () => {
    const { wrapper } = await mount()

    expect(buttonByText(wrapper, 'Reset')).toBeUndefined()
  })

  it('enables save once a field changed', async () => {
    const { wrapper } = await mount()

    await buttonByText(wrapper, 'Left to right')?.trigger('click')

    await vi.waitFor(() => expect(buttonByText(wrapper, 'Save')?.attributes('disabled')).toBeUndefined())
  })

  it('saves the edited settings', async () => {
    updateReaderSettings.mockResolvedValue({
      success: true,
      data: { ...manhwaDefaults(), readingMode: 'paged-ltr' },
    })
    const { wrapper, model } = await mount()

    await buttonByText(wrapper, 'Left to right')?.trigger('click')
    await vi.waitFor(() => expect(buttonByText(wrapper, 'Save')?.attributes('disabled')).toBeUndefined())
    await buttonByText(wrapper, 'Save')?.trigger('click')

    await vi.waitFor(() => expect(updateReaderSettings).toHaveBeenCalledWith({
      type: 'manhwa',
      readingMode: 'paged-ltr',
      pageScale: 'fit-width',
      doublePage: false,
    }))
    expect(model.value.readingMode).toBe('paged-ltr')
  })

  it('toasts when save fails', async () => {
    updateReaderSettings.mockResolvedValue({
      success: false,
      error: { status: 500, message: 'boom' },
    })
    const { wrapper } = await mount()

    await buttonByText(wrapper, 'Left to right')?.trigger('click')
    await vi.waitFor(() => expect(buttonByText(wrapper, 'Save')?.attributes('disabled')).toBeUndefined())
    await buttonByText(wrapper, 'Save')?.trigger('click')

    await vi.waitFor(() => expect(useToast().messages.value).toEqual([
      { text: 'Unknown error', color: 'error' },
    ]))
  })

  it('offers reset when the saved settings differ from the type defaults', async () => {
    const { wrapper } = await mount({
      type: 'manhwa',
      readingMode: 'paged-rtl',
      pageScale: 'fit-screen',
      doublePage: true,
    })

    expect(buttonByText(wrapper, 'Reset')?.exists()).toBe(true)
  })

  it('restores the type defaults', async () => {
    const restored = manhwaDefaults()
    updateReaderSettings.mockResolvedValue({ success: true, data: restored })
    const { wrapper, model } = await mount({
      type: 'manhwa',
      readingMode: 'paged-rtl',
      pageScale: 'fit-screen',
      doublePage: true,
    })

    await buttonByText(wrapper, 'Reset')?.trigger('click')

    await vi.waitFor(() => expect(updateReaderSettings).toHaveBeenCalledWith(restored))
    expect(model.value).toEqual(restored)
  })

  it('blocks double page and non-width scales in webtoon mode', async () => {
    const { wrapper } = await mount()

    expect(buttonByText(wrapper, 'Double page')?.attributes('disabled')).toBeDefined()
    expect(buttonByText(wrapper, 'Fit height')?.attributes('disabled')).toBeDefined()
    expect(buttonByText(wrapper, 'Fit screen')?.attributes('disabled')).toBeDefined()
    expect(buttonByText(wrapper, 'Fit width')?.attributes('disabled')).toBeUndefined()
    expect(buttonByText(wrapper, 'Single page')?.attributes('disabled')).toBeUndefined()
  })
})
