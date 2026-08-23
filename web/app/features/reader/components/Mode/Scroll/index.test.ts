// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { DetailedChapter } from '~/features/chapters/types'
import type { Comic } from '~/features/comics/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { h, ref } from 'vue'
import { VApp } from 'vuetify/components'
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

async function mount(opts: { page?: number } = {}): Promise<VueWrapper> {
  const page = ref(opts.page ?? 0)
  const showOverlay = ref(true)

  return await mountSuspended({
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
}

describe('readerModeScroll', () => {
  it('constrains the virtual list to the viewport so scrollToIndex can move', async () => {
    const wrapper = await mount({ page: 2 })

    expect(wrapper.find('.v-virtual-scroll').classes()).toContain('h-100')
  })
})
