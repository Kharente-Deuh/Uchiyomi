// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { DetailedChapter } from '~/features/chapters/types'
import type { Comic } from '~/features/comics/types'
import type { ReaderSettings } from '~/features/reader/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h, ref } from 'vue'
import { VApp, VImg } from 'vuetify/components'
import { ASURA_SOURCE_NAME } from '~/constants'
import Paged from './Paged.vue'

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
    chapterCount: 3,
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
    next: { id: 'ch-3', title: 'Three', number: 3 },
    previous: { id: 'ch-1', title: 'One', number: 1 },
    ...overrides,
  }
}

function clickZones(wrapper: VueWrapper): ReturnType<VueWrapper['findAll']> {
  return wrapper.findAll('.position-absolute')
}

async function mount(opts: {
  settings?: Partial<ReaderSettings>
  chapter?: Partial<DetailedChapter>
  page?: number
  showOverlay?: boolean
} = {}): Promise<{
  wrapper: VueWrapper
  page: ReturnType<typeof ref<number>>
  showOverlay: ReturnType<typeof ref<boolean>>
  events: { previousChapter: number, nextChapter: number, fetchNextChapter: number, fetchPreviousChapter: number }
}> {
  const page = ref(opts.page ?? 0)
  const showOverlay = ref(opts.showOverlay ?? true)
  const events = { previousChapter: 0, nextChapter: 0, fetchNextChapter: 0, fetchPreviousChapter: 0 }
  const wrapper = await mountSuspended({
    setup: () => ({ page, showOverlay }),
    render: () => h(VApp, () => [
      h(Paged, {
        'comic': comic(),
        'chapter': chapter(opts.chapter),
        'settings': settings(opts.settings),
        'page': page.value,
        'showOverlay': showOverlay.value,
        'onUpdate:page': (value: number) => {
          page.value = value
        },
        'onUpdate:showOverlay': (isShown: boolean) => {
          showOverlay.value = isShown
        },
        'onPreviousChapter': () => {
          events.previousChapter++
        },
        'onNextChapter': () => {
          events.nextChapter++
        },
        'onFetchNextChapter': () => {
          events.fetchNextChapter++
        },
        'onFetchPreviousChapter': () => {
          events.fetchPreviousChapter++
        },
      }),
    ]),
  })

  return { wrapper, page, showOverlay, events }
}

describe('readerModePaged', () => {
  it('shows the current page image', async () => {
    const { wrapper } = await mount()

    expect(wrapper.findAllComponents(VImg).map(img => img.props('src'))).toEqual(['/p1'])
  })

  it('advances on a right click in left-to-right mode', async () => {
    const { wrapper, page } = await mount()

    await clickZones(wrapper)[1]?.trigger('click')

    expect(page.value).toBe(1)
  })

  it('goes back on a left click in left-to-right mode', async () => {
    const { wrapper, page } = await mount({ page: 1 })

    await clickZones(wrapper)[0]?.trigger('click')

    expect(page.value).toBe(0)
  })

  it('mirrors left and right clicks in right-to-left mode', async () => {
    const { wrapper, page } = await mount({ settings: { readingMode: 'paged-rtl' } })

    await clickZones(wrapper)[0]?.trigger('click')

    expect(page.value).toBe(1)
  })

  it('toggles the overlay from the center', async () => {
    const { wrapper, showOverlay } = await mount({ showOverlay: true })

    await wrapper.find('.h-screen.w-screen').trigger('click')

    expect(showOverlay.value).toBe(false)
  })

  it('snaps an odd page down when double page is on', async () => {
    const { page } = await mount({ settings: { doublePage: true }, page: 1 })

    expect(page.value).toBe(0)
  })

  it('shows two images in double page mode', async () => {
    const { wrapper } = await mount({ settings: { doublePage: true } })

    expect(wrapper.findAllComponents(VImg).map(img => img.props('src'))).toEqual(['/p1', '/p2'])
  })

  it('steps by two pages when double page is on', async () => {
    const { wrapper, page } = await mount({
      settings: { doublePage: true },
      chapter: { pagesNb: 4, pageUrls: ['/p1', '/p2', '/p3', '/p4'] },
    })

    await clickZones(wrapper)[1]?.trigger('click')

    expect(page.value).toBe(2)
  })

  it('shows the between-chapters card after the last page and prefetches the next chapter', async () => {
    const { wrapper, events } = await mount({ page: 2 })

    await clickZones(wrapper)[1]?.trigger('click')

    expect(wrapper.text()).toContain('Current chapter')
    expect(wrapper.text()).toContain('Next chapter')
    expect(events.fetchNextChapter).toBe(1)
  })

  it('emits next chapter on a second click past the last page', async () => {
    const { wrapper, events } = await mount({ page: 2 })

    await clickZones(wrapper)[1]?.trigger('click')
    await clickZones(wrapper)[1]?.trigger('click')

    expect(events.nextChapter).toBe(1)
  })

  it('prefetches the previous chapter when going back from the first page', async () => {
    const { wrapper, events } = await mount()

    await clickZones(wrapper)[0]?.trigger('click')

    expect(wrapper.text()).toContain('Previous chapter')
    expect(events.fetchPreviousChapter).toBe(1)
  })
})
