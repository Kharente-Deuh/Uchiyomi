// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { DetailedChapter } from '~/features/chapters/types'
import type { Comic } from '~/features/comics/types'
import type { ReaderSettings } from '~/features/reader/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it, vi } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { VApp, VBtn, VProgressCircular } from 'vuetify/components'
import { ASURA_SOURCE_NAME } from '~/constants'
import Reader from './index.vue'

const OverlayStub = defineComponent({
  name: 'ReaderOverlay',
  props: {
    comic: { type: Object, required: true },
    chapter: { type: Object, required: true },
    modelValue: { type: Boolean, required: true },
    page: { type: Number, required: true },
    doublePage: { type: Boolean, default: false },
  },
  template: '<div data-test="overlay" />',
})

const PagedStub = defineComponent({
  name: 'ReaderModePaged',
  props: {
    comic: { type: Object, required: true },
    chapter: { type: Object, required: true },
    settings: { type: Object, required: true },
    page: { type: Number, required: true },
    showOverlay: { type: Boolean, required: true },
  },
  template: '<div data-test="paged" />',
})

const ScrollStub = defineComponent({
  name: 'ReaderModeScroll',
  props: {
    comic: { type: Object, required: true },
    chapter: { type: Object, required: true },
    page: { type: Number, required: true },
    showOverlay: { type: Boolean, required: true },
  },
  template: '<div data-test="scroll" />',
})

function comic(overrides: Partial<Comic> = {}): Comic {
  return {
    id: 'c1',
    title: 'Solo Leveling',
    slug: 'solo-leveling',
    source: ASURA_SOURCE_NAME,
    status: 'ongoing',
    type: 'manhwa',
    author: 'Chugong',
    artist: 'Jang',
    description: 'A hunter',
    cover: '/cover',
    genres: [],
    altTitles: [],
    chapterCount: 1,
    ...overrides,
  }
}

function settings(overrides: Partial<ReaderSettings> = {}): ReaderSettings {
  return {
    type: 'manhwa',
    readingMode: 'paged-ltr',
    pageScale: 'fit-width',
    doublePage: false,
    ...overrides,
  }
}

function chapter(overrides: Partial<DetailedChapter> = {}): DetailedChapter {
  return {
    id: 'ch-2',
    comicId: 'c1',
    publishedAt: new Date('2026-01-01'),
    sourceChapterSlug: 'ch-2',
    title: 'Two',
    number: 2,
    pagesNb: 3,
    download: 100,
    pageUrls: ['/p1', '/p2', '/p3'],
    ...overrides,
  }
}

async function mount(opts: {
  chapter?: Partial<DetailedChapter>
  settings?: Partial<ReaderSettings>
  retryDownloadLoading?: boolean
} = {}): Promise<{ wrapper: VueWrapper, retry: ReturnType<typeof vi.fn> }> {
  const page = ref(0)
  const retry = vi.fn()
  const wrapper = await mountSuspended({
    setup: () => ({ page }),
    render: () => h(VApp, () => [
      h(Reader, {
        'comic': comic(),
        'chapter': chapter(opts.chapter),
        'settings': settings(opts.settings),
        'retryDownloadLoading': opts.retryDownloadLoading ?? false,
        'page': page.value,
        'onUpdate:page': (value: number) => {
          page.value = value
        },
        'onRetryDownload': retry,
      }),
    ]),
  }, {
    global: {
      stubs: {
        ReaderOverlay: OverlayStub,
        ReaderModePaged: PagedStub,
        ReaderModeScroll: ScrollStub,
      },
    },
  })

  return { wrapper, retry }
}

describe('reader', () => {
  it('shows paged mode once the chapter is downloaded', async () => {
    const { wrapper } = await mount()

    expect(wrapper.find('[data-test="paged"]').exists()).toBe(true)
    expect(wrapper.findComponent(VProgressCircular).exists()).toBe(false)
  })

  it('shows download progress while the chapter is downloading', async () => {
    const { wrapper } = await mount({ chapter: { download: 40 } })

    expect(wrapper.text()).toContain('Chapter downloading...')
    expect(wrapper.findComponent(VProgressCircular).exists()).toBe(true)
  })

  it('offers retry when the download failed', async () => {
    const { wrapper, retry } = await mount({ chapter: { download: -1 } })

    expect(wrapper.text()).toContain('Chapter download error')
    const retryBtn = wrapper.findAllComponents(VBtn).find(b => b.text().includes('Retry'))
    await retryBtn?.trigger('click')

    expect(retry).toHaveBeenCalledOnce()
  })

  it('does not mount paged mode for webtoon settings', async () => {
    const { wrapper } = await mount({ settings: { readingMode: 'webtoon' } })

    expect(wrapper.find('[data-test="paged"]').exists()).toBe(false)
  })

  it('shows scroll mode for webtoon settings', async () => {
    const { wrapper } = await mount({ settings: { readingMode: 'webtoon' } })

    expect(wrapper.find('[data-test="scroll"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="paged"]').exists()).toBe(false)
  })
})
