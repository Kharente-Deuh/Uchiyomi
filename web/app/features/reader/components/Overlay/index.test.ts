// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { DetailedChapter } from '~/features/chapters/types'
import type { Comic } from '~/features/comics/types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp, VSlider } from 'vuetify/components'
import { ASURA_SOURCE_NAME } from '~/constants'
import Overlay from './index.vue'
import Rail from './Rail.vue'

const HeaderStub = defineComponent({
  name: 'ReaderOverlayHeader',
  props: {
    comic: { type: Object, required: true },
    chapter: { type: Object, required: true },
  },
  template: '<div data-test="header" />',
})

const FooterStub = defineComponent({
  name: 'ReaderOverlayFooter',
  props: {
    comic: { type: Object, required: true },
    chapter: { type: Object, required: true },
  },
  template: '<div data-test="footer" />',
})

function comic(): Comic {
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
  }
}

function chapter(): DetailedChapter {
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
  }
}

async function mount(isOpen = true, page = 1): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [
      h(Overlay, {
        'comic': comic(),
        'chapter': chapter(),
        'modelValue': isOpen,
        'page': page,
        'onUpdate:modelValue': () => {},
        'onUpdate:page': () => {},
      }),
    ]),
  }, {
    global: {
      stubs: {
        ReaderOverlayHeader: HeaderStub,
        ReaderOverlayFooter: FooterStub,
      },
    },
  })
}

describe('readerOverlay', () => {
  it('shows the page rail with the current page when the overlay is open', async () => {
    const wrapper = await mount(true, 1)

    expect(wrapper.text()).toContain('2')
    expect(wrapper.text()).toContain('3')
    expect(wrapper.findComponent(VSlider).exists()).toBe(true)
  })

  it('hides the page rail when the overlay is closed', async () => {
    const wrapper = await mount(false, 1)
    const rail = wrapper.findComponent(Rail)

    expect(rail.exists()).toBe(true)
    expect((rail.element as HTMLElement).style.display).toBe('none')
  })
})
