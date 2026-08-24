// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { DetailedChapter } from '~/features/chapters/types'
import type { Comic } from '~/features/comics/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it, vi } from 'vitest'
import { h, ref } from 'vue'
import { VApp, VImg, VVirtualScroll } from 'vuetify/components'
import { ASURA_SOURCE_NAME } from '~/constants'
import Img from './Img.vue'
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
  page: ReturnType<typeof ref<number>>
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

  return { wrapper, page, showOverlay }
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

  it('follows the topmost intersecting page', async () => {
    const { wrapper, page } = await mount()
    const imgs = wrapper.findAllComponents(Img)

    await imgs[1]!.vm.$emit('intersecting', true)
    await imgs[2]!.vm.$emit('intersecting', true)

    expect(page.value).toBe(1)
  })

  it('does not let leftover intersecting pages overwrite page while moving to another page', async () => {
    const { wrapper, page } = await mount({ page: 2 })
    const imgs = wrapper.findAllComponents(Img)

    await imgs[0]!.vm.$emit('intersecting', true)
    await imgs[1]!.vm.$emit('intersecting', true)

    expect(page.value).toBe(2)
  })

  it('does not let the target intersecting unlock page updates before the scroll ends', async () => {
    const { wrapper, page } = await mount({ page: 2 })
    const imgs = wrapper.findAllComponents(Img)

    await imgs[2]!.vm.$emit('intersecting', true)
    await imgs[2]!.vm.$emit('intersecting', false)
    await imgs[1]!.vm.$emit('intersecting', true)

    expect(page.value).toBe(2)
  })

  it('does not hide the overlay on scroll after the target page intersects but before scroll ends', async () => {
    const { wrapper, page, showOverlay } = await mount()
    const el = wrapper.find('.v-virtual-scroll')
    setScrollMetrics(el.element as HTMLElement, { scrollHeight: 1000, scrollTop: 40, clientHeight: 100 })
    await el.trigger('scroll')

    page.value = 2
    await wrapper.vm.$nextTick()
    await wrapper.findAllComponents(Img)[2]!.vm.$emit('intersecting', true)
    await el.trigger('scroll')

    expect(showOverlay.value).toBe(true)
  })

  it('follows intersecting pages again after the rail scroll ends', async () => {
    const { wrapper, page } = await mount({ page: 2 })
    const imgs = wrapper.findAllComponents(Img)

    await imgs[2]!.vm.$emit('intersecting', true)
    await wrapper.find('.v-virtual-scroll').trigger('scrollend')
    await imgs[2]!.vm.$emit('intersecting', false)
    await imgs[0]!.vm.$emit('intersecting', true)

    expect(page.value).toBe(0)
  })

  it('scrolls the list when the page model changes from the rail', async () => {
    const { wrapper, page } = await mount()
    const scrollToIndex = vi.spyOn(wrapper.findComponent(VVirtualScroll).vm, 'scrollToIndex')

    page.value = 2
    await wrapper.vm.$nextTick()

    expect(scrollToIndex).toHaveBeenCalledWith(2)
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
