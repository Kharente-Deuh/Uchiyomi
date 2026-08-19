// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { Chapter } from '~/features/chapters/types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { VApp } from 'vuetify/components'
import Chapters from './index.vue'

const { getByComicId, getByIds, retryDownload, pause, resume } = vi.hoisted(() => ({
  getByComicId: vi.fn(),
  getByIds: vi.fn(),
  retryDownload: vi.fn(),
  pause: vi.fn(),
  resume: vi.fn(),
}))

function chaptersApiStub(): {
  getByComicId: typeof getByComicId
  getByIds: typeof getByIds
  retryDownload: typeof retryDownload
} {
  return { getByComicId, getByIds, retryDownload }
}

function intervalStub(_fn: () => void): { pause: typeof pause, resume: typeof resume } {
  return { pause, resume }
}

mockNuxtImport('createChaptersApi', () => chaptersApiStub)
mockNuxtImport('useIntervalFn', () => intervalStub)

const ItemStub = defineComponent({
  name: 'ComicsChaptersItem',
  props: {
    chapter: { type: Object, required: true },
    retryLoading: { type: Boolean, default: false },
  },
  emits: ['retry'],
  template: '<button data-test="chapter" @click="$emit(\'retry\')">{{ chapter.number }}</button>',
})

function chapter(overrides: Partial<Chapter> = {}): Chapter {
  return {
    id: 'ch-1',
    comicId: 'c1',
    title: 'One',
    number: 1,
    publishedAt: new Date('2026-01-01'),
    sourceChapterSlug: '1',
    pagesNb: 10,
    download: 100,
    ...overrides,
  }
}

async function mount(id = 'c1'): Promise<VueWrapper> {
  return mountSuspended(
    { render: () => h(VApp, () => [h(Chapters, { id })]) },
    {
      global: {
        stubs: {
          ComicsChaptersItem: ItemStub,
          VVirtualScroll: false,
        },
      },
    },
  )
}

beforeEach(() => {
  getByComicId.mockReset()
  getByIds.mockReset()
  retryDownload.mockReset()
  pause.mockReset()
  resume.mockReset()
  getByComicId.mockResolvedValue({ success: true, data: [] })
  getByIds.mockResolvedValue({ success: true, data: [] })
  retryDownload.mockResolvedValue({ success: true, data: undefined })
})

describe('comicsChapters', () => {
  it('fetches chapters on mount', async () => {
    await mount()
    expect(getByComicId).toHaveBeenCalledWith('c1')
  })

  it('shows an empty state when there are no chapters', async () => {
    const wrapper = await mount()
    await vi.waitFor(() => expect(wrapper.text()).toContain('No chapters'))
  })

  it('renders a row per chapter', async () => {
    getByComicId.mockResolvedValue({ success: true, data: [chapter({ id: 'ch-2', number: 2 }), chapter()] })
    const wrapper = await mount()

    await vi.waitFor(() => expect(wrapper.findAll('[data-test="chapter"]').map(n => n.text())).toEqual(['2', '1']))
  })

  it('toggles sort between latest and oldest', async () => {
    getByComicId.mockResolvedValue({ success: true, data: [chapter({ id: 'ch-2', number: 2 }), chapter()] })
    const wrapper = await mount()
    await vi.waitFor(() => expect(wrapper.findAll('[data-test="chapter"]').length).toBe(2))

    expect(wrapper.text()).toContain('Latest')

    await wrapper.findAll('button').at(0)!.trigger('click')

    expect(wrapper.text()).toContain('Oldest')
    expect(wrapper.findAll('[data-test="chapter"]').map(n => n.text())).toEqual(['1', '2'])
  })

  it('resumes polling when a download is in progress', async () => {
    getByComicId.mockResolvedValue({ success: true, data: [chapter({ download: 40 })] })
    await mount()

    await vi.waitFor(() => expect(resume).toHaveBeenCalled())
  })

  it('retries a failed download from the chapter row', async () => {
    getByComicId.mockResolvedValue({ success: true, data: [chapter({ download: -1 })] })
    const wrapper = await mount()
    await vi.waitFor(() => expect(wrapper.find('[data-test="chapter"]').exists()).toBe(true))

    await wrapper.find('[data-test="chapter"]').trigger('click')

    expect(retryDownload).toHaveBeenCalledWith('ch-1')
  })
})
