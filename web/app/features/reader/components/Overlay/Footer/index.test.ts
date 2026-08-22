// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { DetailedChapter } from '~/features/chapters/types'
import type { Comic } from '~/features/comics/types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp, VBtn } from 'vuetify/components'
import { ASURA_SOURCE_NAME } from '~/constants'
import Footer from './index.vue'

const { getByComicId } = vi.hoisted(() => ({
  getByComicId: vi.fn(),
}))

function createChaptersApiStub(): { getByComicId: typeof getByComicId } {
  return { getByComicId }
}

mockNuxtImport('createChaptersApi', () => createChaptersApiStub)

const MenuStub = defineComponent({
  name: 'ReaderOverlayFooterChaptersMenu',
  props: {
    comicId: { type: String, required: true },
    currentChapter: { type: Object, required: true },
  },
  template: '<div data-test="chapters-menu" />',
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
    chapterCount: 3,
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
    pageUrls: ['/p1'],
    next: { id: 'ch-3', title: 'Three', number: 3 },
    previous: { id: 'ch-1', title: 'One', number: 1 },
    ...overrides,
  }
}

function buttons(wrapper: VueWrapper): VueWrapper<InstanceType<typeof VBtn>>[] {
  return wrapper.findAllComponents(VBtn)
}

async function mount(value: DetailedChapter = chapter()): Promise<VueWrapper> {
  return mountSuspended({
    render: () => h(VApp, () => [h(Footer, { comic: comic(), chapter: value })]),
  }, {
    global: {
      stubs: { ReaderOverlayFooterChaptersMenu: MenuStub },
    },
  })
}

beforeEach(() => {
  getByComicId.mockReset()
  getByComicId.mockResolvedValue({ success: true, data: [] })
})

describe('readerOverlayFooter', () => {
  it('links to the previous and next chapters', async () => {
    const wrapper = await mount()

    expect(wrapper.find('a[href="/comic/c1/ch-1"]').exists()).toBe(true)
    expect(wrapper.find('a[href="/comic/c1/ch-3"]').exists()).toBe(true)
    expect(buttons(wrapper).find(b => b.text().includes('Prev'))?.props('disabled')).toBe(false)
    expect(buttons(wrapper).find(b => b.text().includes('Next'))?.props('disabled')).toBe(false)
  })

  it('disables previous and next at the ends of the comic', async () => {
    const wrapper = await mount(chapter({ next: undefined, previous: undefined }))

    expect(buttons(wrapper).find(b => b.text().includes('Prev'))?.props('disabled')).toBe(true)
    expect(buttons(wrapper).find(b => b.text().includes('Next'))?.props('disabled')).toBe(true)
  })
})
