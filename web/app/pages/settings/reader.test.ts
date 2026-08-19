// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { ReaderSettings } from '~/features/reader/types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp } from 'vuetify/components'
import { useToast } from '~/composables/toast.composable'
import ReaderPage from './reader.vue'

const { getReaderSettings } = await vi.hoisted(async () => ({
  getReaderSettings: vi.fn(),
}))

function readerSettingsApiStub(): { getReaderSettings: typeof getReaderSettings } {
  return { getReaderSettings }
}

mockNuxtImport('createReaderSettingsApi', () => readerSettingsApiStub)

const CardStub = defineComponent({
  name: 'ReaderCard',
  props: {
    modelValue: { type: Object, required: true },
    type: { type: String, required: true },
  },
  template: '<div data-test="reader-card">{{ type }} {{ modelValue.readingMode }}</div>',
})

const manhwa: ReaderSettings = {
  type: 'manhwa',
  readingMode: 'webtoon',
  pageScale: 'fit-width',
  doublePage: false,
}

const manga: ReaderSettings = {
  type: 'manga',
  readingMode: 'paged-rtl',
  pageScale: 'fit-screen',
  doublePage: false,
}

async function mount(): Promise<VueWrapper> {
  return mountSuspended(
    { render: () => h(VApp, () => [h(ReaderPage)]) },
    { global: { stubs: { ReaderCard: CardStub } } },
  )
}

beforeEach(() => {
  getReaderSettings.mockReset()
  getReaderSettings.mockResolvedValue({ success: true, data: [manhwa, manga] })
  useToast().messages.value.length = 0
})

describe('settings/reader', () => {
  it('loads the settings on mount', async () => {
    await mount()

    await vi.waitFor(() => expect(getReaderSettings).toHaveBeenCalledTimes(1))
  })

  it('renders the card for the current comic type', async () => {
    const wrapper = await mount()

    await vi.waitFor(() => expect(wrapper.find('[data-test="reader-card"]').text()).toBe('manhwa webtoon'))
  })

  it('keeps the card hidden until settings for the type are in', async () => {
    getReaderSettings.mockResolvedValue({ success: true, data: [manga] })
    const wrapper = await mount()

    await vi.waitFor(() => expect(getReaderSettings).toHaveBeenCalled())
    expect(wrapper.find('[data-test="reader-card"]').exists()).toBe(false)
  })

  it('toasts when the settings cannot be loaded', async () => {
    getReaderSettings.mockResolvedValue({
      success: false,
      error: { status: 500, message: 'boom' },
    })

    await mount()

    await vi.waitFor(() => expect(useToast().messages.value).toEqual([
      { text: 'Unknown error', color: 'error' },
    ]))
  })
})
