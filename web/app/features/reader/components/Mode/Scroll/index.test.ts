// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { DetailedChapter } from '~/features/chapters/types'
import type { Comic } from '~/features/comics/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h, ref } from 'vue'
import { VApp, VImg } from 'vuetify/components'
import { ASURA_SOURCE_NAME } from '~/constants'
import Scroll from './index.vue'

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

async function mount(opts: { page?: number } = {}): Promise<{
  wrapper: VueWrapper
  showOverlay: ReturnType<typeof ref<boolean>>
}> {
  const page = ref(opts.page ?? 0)
  const showOverlay = ref(true)

  const wrapper = await mountSuspended({
    setup: () => ({ page, showOverlay }),
    render: () => h(VApp, () => [
      h(Scroll, {
        'comic': comic(),
        'chapter': chapter(),
        'page': page.value,
        'showOverlay': showOverlay.value,
        'onUpdate:page': (value: number) => {
          page.value = value
        },
        'onUpdate:showOverlay': (isShown: boolean) => {
          showOverlay.value = isShown
        },
      }),
    ]),
  })

  return { wrapper, showOverlay }
}

function setScrollMetrics(el: HTMLElement, metrics: { scrollHeight: number, scrollTop: number, clientHeight: number }): void {
  Object.defineProperties(el, {
    scrollHeight: { configurable: true, value: metrics.scrollHeight },
    scrollTop: { configurable: true, value: metrics.scrollTop },
    clientHeight: { configurable: true, value: metrics.clientHeight },
  })
}

describe('readerModeScroll', () => {
  it('constrains the virtual list to the viewport so scrollToIndex can move', async () => {
    const { wrapper } = await mount({ page: 2 })

    expect(wrapper.find('.v-virtual-scroll').classes()).toContain('h-100')
  })

  it('renders a page image per url', async () => {
    const { wrapper } = await mount()

    expect(wrapper.findAllComponents(VImg).map(img => img.props('src'))).toEqual(['/p1', '/p2', '/p3'])
  })

  it('ignores the first scroll so restore does not hide the overlay', async () => {
    const { wrapper, showOverlay } = await mount()
    const el = wrapper.find('.v-virtual-scroll').element as HTMLElement
    setScrollMetrics(el, { scrollHeight: 1000, scrollTop: 40, clientHeight: 100 })

    await wrapper.find('.v-virtual-scroll').trigger('scroll')

    expect(showOverlay.value).toBe(true)
  })

  it('hides the overlay on a later scroll that is not at the bottom', async () => {
    const { wrapper, showOverlay } = await mount()
    const el = wrapper.find('.v-virtual-scroll').element as HTMLElement
    setScrollMetrics(el, { scrollHeight: 1000, scrollTop: 40, clientHeight: 100 })

    await wrapper.find('.v-virtual-scroll').trigger('scroll')
    await wrapper.find('.v-virtual-scroll').trigger('scroll')

    expect(showOverlay.value).toBe(false)
  })

  it('shows the overlay when scrolled to the bottom', async () => {
    const { wrapper, showOverlay } = await mount()
    const el = wrapper.find('.v-virtual-scroll').element as HTMLElement
    setScrollMetrics(el, { scrollHeight: 1000, scrollTop: 40, clientHeight: 100 })
    await wrapper.find('.v-virtual-scroll').trigger('scroll')
    await wrapper.find('.v-virtual-scroll').trigger('scroll')
    expect(showOverlay.value).toBe(false)

    setScrollMetrics(el, { scrollHeight: 1000, scrollTop: 900, clientHeight: 100 })
    await wrapper.find('.v-virtual-scroll').trigger('scroll')

    expect(showOverlay.value).toBe(true)
  })
})
