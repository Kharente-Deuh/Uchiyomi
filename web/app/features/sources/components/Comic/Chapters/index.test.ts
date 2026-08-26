// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { SourceComicChapter } from '../../../types'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp } from 'vuetify/components'
import { ASURA_SOURCE_NAME } from '~/constants'
import Chapters from './index.vue'

const { sort, chapters, loading, fetchChapters } = await vi.hoisted(async () => {
  const { ref } = await import('vue')

  return {
    sort: ref<'asc' | 'desc'>('desc'),
    chapters: ref<SourceComicChapter[]>([]),
    loading: ref(false),
    fetchChapters: vi.fn(),
  }
})

vi.mock('~/features/sources/composables/sources-chapters.composable', () => ({
  useSourceChapters: () => ({
    sort,
    chapters,
    loading,
    fetchChapters,
  }),
}))

const ItemStub = defineComponent({
  name: 'SourcesComicChaptersItem',
  props: {
    source: { type: String, required: true },
    chapter: { type: Object, required: true },
    comicOriginUrl: { type: String, required: true },
  },
  template: '<div data-test="chapter">{{ chapter.number }}</div>',
})

function chapter(overrides: Partial<SourceComicChapter> = {}): SourceComicChapter {
  return {
    id: 'ch-1',
    title: 'One',
    number: 1,
    publishedAt: new Date('2026-01-01'),
    ...overrides,
  }
}

async function mount(): Promise<VueWrapper> {
  return mountSuspended(
    {
      render: () => h(VApp, () => [h(Chapters, {
        source: ASURA_SOURCE_NAME,
        slug: 'solo-leveling',
        comicOriginUrl: 'https://asurascans.com/series/solo-leveling',
      })]),
    },
    {
      global: {
        stubs: {
          SourcesComicChaptersItem: ItemStub,
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

describe('sourcesComicChapters', () => {
  it('fetches chapters on mount', async () => {
    await mount()
    expect(fetchChapters).toHaveBeenCalledWith('solo-leveling')
  })

  it('shows an empty state when there are no chapters', async () => {
    const wrapper = await mount()
    expect(wrapper.text()).toContain('No chapters')
  })

  it('renders a row per chapter', async () => {
    chapters.value = [chapter({ id: 'ch-2', number: 2 }), chapter()]
    const wrapper = await mount()

    expect(wrapper.findAll('[data-test="chapter"]').map(n => n.text())).toEqual(['2', '1'])
  })

  it('toggles sort between latest and oldest', async () => {
    chapters.value = [chapter({ id: 'ch-2', number: 2 }), chapter()]
    const wrapper = await mount()

    expect(wrapper.text()).toContain('Latest')

    await wrapper.findAll('button').find(b => b.text().includes('Latest'))!.trigger('click')

    expect(sort.value).toBe('asc')
    expect(wrapper.text()).toContain('Oldest')
  })
})
