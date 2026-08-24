// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { AsuraScansComicChapter } from '../../../types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp } from 'vuetify/components'
import Chapters from './index.vue'

const { sort, chapters, loading, fetchChapters } = await vi.hoisted(async () => {
  const { ref: vueRef } = await import('vue')

  return {
    sort: vueRef<'asc' | 'desc'>('desc'),
    chapters: vueRef<AsuraScansComicChapter[]>([]),
    loading: vueRef(false),
    fetchChapters: vi.fn(),
  }
})

vi.mock('~/features/asurascans/composables/asurascans-chapters.composable', () => ({
  useAsuraScansChapters: () => ({
    sort,
    chapters,
    loading,
    fetchChapters,
  }),
}))

const ItemStub = defineComponent({
  name: 'AsuraScansComicChaptersItem',
  props: {
    chapter: { type: Object, required: true },
    comicOriginUrl: { type: String, required: true },
  },
  template: '<div data-test="chapter">{{ chapter.number }}</div>',
})

function chapter(number: number): AsuraScansComicChapter {
  return {
    id: `ch-${number}`,
    title: `Chapter ${number}`,
    number,
    publishedAt: new Date('2026-01-01'),
  }
}

async function mount(slug = 'solo-leveling'): Promise<VueWrapper> {
  return mountSuspended(
    {
      render: () => h(VApp, () => [h(Chapters, {
        slug,
        comicOriginUrl: 'https://asurascans.com/series/solo-leveling',
      })]),
    },
    {
      global: {
        stubs: {
          AsuraScansComicChaptersItem: ItemStub,
          VVirtualScroll: false,
        },
      },
    },
  )
}

beforeEach(() => {
  sort.value = 'desc'
  chapters.value = []
  loading.value = false
  fetchChapters.mockReset()
})

describe('asuraScansComicChapters', () => {
  it('fetches chapters on mount', async () => {
    await mount()
    expect(fetchChapters).toHaveBeenCalled()
  })

  it('shows an empty state when there are no chapters', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('No chapters')
  })

  it('shows a spinner while loading', async () => {
    loading.value = true
    const wrapper = await mount()
    expect(wrapper.find('.v-progress-circular').exists()).toBe(true)
  })

  it('toggles sort between latest and oldest', async () => {
    chapters.value = [chapter(2), chapter(1)]
    const wrapper = await mount()

    expect(wrapper.text()).toContain('Latest')

    const buttons = wrapper.findAll('button')
    await buttons[0]!.trigger('click')

    expect(sort.value).toBe('asc')
    expect(wrapper.text()).toContain('Oldest')
  })
})
