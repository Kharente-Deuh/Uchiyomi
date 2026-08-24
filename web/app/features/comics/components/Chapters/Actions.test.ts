// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { Chapter } from '~/features/chapters/types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { VApp, VListItem } from 'vuetify/components'
import { useToast } from '~/composables/toast.composable'
import Actions from './Actions.vue'

const { setChaptersProgress, retryChaptersDownload } = vi.hoisted(() => ({
  setChaptersProgress: vi.fn(),
  retryChaptersDownload: vi.fn(),
}))

function comicsApiStub(): {
  setChaptersProgress: typeof setChaptersProgress
  retryChaptersDownload: typeof retryChaptersDownload
} {
  return {
    setChaptersProgress,
    retryChaptersDownload,
  }
}

mockNuxtImport('createComicsApi', () => comicsApiStub)

const MenuStub = defineComponent({
  name: 'VMenu',
  inheritAttrs: false,
  setup(_, { slots }) {
    return () => slots.default?.()
  },
})

function chapter(overrides: Partial<Chapter> = {}): Chapter {
  return {
    id: 'ch-1',
    comicId: 'c1',
    title: 'Chapter 1',
    number: 1,
    publishedAt: new Date('2026-08-18T12:00:00.000Z'),
    sourceChapterSlug: '1',
    pagesNb: 20,
    download: 100,
    ...overrides,
  }
}

async function mount(initialChapters: Chapter[]): Promise<{
  wrapper: VueWrapper
  chaptersRef: ReturnType<typeof ref<Chapter[]>>
  onRefetchChapters: ReturnType<typeof vi.fn>
}> {
  const chaptersRef = ref<Chapter[]>([...initialChapters])
  const onRefetchChapters = vi.fn()

  const wrapper = await mountSuspended({
    render: () => h(VApp, () => [h(Actions, {
      'comicId': 'c1',
      'modelValue': chaptersRef.value,
      'onUpdate:modelValue': (val: Chapter[]) => {
        chaptersRef.value = val
      },
      'onRefetchChapters': onRefetchChapters,
    })]),
  }, {
    global: {
      stubs: { VMenu: MenuStub },
    },
  })

  return { wrapper, chaptersRef, onRefetchChapters }
}

beforeEach(() => {
  useToast().messages.value.length = 0
  vi.spyOn(console, 'error').mockImplementation(() => {})
  setChaptersProgress.mockReset()
  retryChaptersDownload.mockReset()
  setChaptersProgress.mockResolvedValue({ success: true, data: undefined })
  retryChaptersDownload.mockResolvedValue({ success: true, data: undefined })
})

describe('comicsChaptersActions', () => {
  it('disables actions when corresponding chapters are empty', async () => {
    const { wrapper } = await mount([
      chapter({ id: 'ch-1', pagesNb: 20, progress: { page: 20, updatedAt: new Date() }, download: 100 }),
    ])

    const items = wrapper.findAllComponents(VListItem)
    expect(items).toHaveLength(3)

    // retryable: disabled because no download === -1
    expect(items[0]!.props('disabled')).toBe(true)
    // setRead: disabled because all chapters are already read
    expect(items[1]!.props('disabled')).toBe(true)
    // setUnread: enabled because ch-1 is read
    expect(items[2]!.props('disabled')).toBe(false)
  })

  it('triggers retryDownload on action click, clears chapters model and emits refetchChapters(false)', async () => {
    const { wrapper, chaptersRef, onRefetchChapters } = await mount([
      chapter({ id: 'ch-1', download: -1 }),
      chapter({ id: 'ch-2', download: 100 }),
    ])

    const items = wrapper.findAllComponents(VListItem)
    expect(items[0]!.props('disabled')).toBe(false)

    await items[0]!.trigger('click')

    expect(retryChaptersDownload).toHaveBeenCalledWith('c1', ['ch-1'])
    expect(chaptersRef.value).toEqual([])
    expect(onRefetchChapters).toHaveBeenCalledWith(false)
  })

  it('triggers setRead on action click, clears chapters model and emits refetchChapters(true)', async () => {
    const { wrapper, chaptersRef, onRefetchChapters } = await mount([
      chapter({ id: 'ch-1', pagesNb: 20, progress: undefined }),
      chapter({ id: 'ch-2', pagesNb: 20, progress: { page: 20, updatedAt: new Date() } }),
    ])

    const items = wrapper.findAllComponents(VListItem)
    expect(items[1]!.props('disabled')).toBe(false)

    await items[1]!.trigger('click')

    expect(setChaptersProgress).toHaveBeenCalledWith({
      comicId: 'c1',
      chapterIds: ['ch-1'],
      read: true,
    })
    expect(chaptersRef.value).toEqual([])
    expect(onRefetchChapters).toHaveBeenCalledWith(true)
  })

  it('triggers setUnread on action click, clears chapters model and emits refetchChapters(true)', async () => {
    const { wrapper, chaptersRef, onRefetchChapters } = await mount([
      chapter({ id: 'ch-1', pagesNb: 20, progress: undefined }),
      chapter({ id: 'ch-2', pagesNb: 20, progress: { page: 20, updatedAt: new Date() } }),
    ])

    const items = wrapper.findAllComponents(VListItem)
    expect(items[2]!.props('disabled')).toBe(false)

    await items[2]!.trigger('click')

    expect(setChaptersProgress).toHaveBeenCalledWith({
      comicId: 'c1',
      chapterIds: ['ch-2'],
      read: false,
    })
    expect(chaptersRef.value).toEqual([])
    expect(onRefetchChapters).toHaveBeenCalledWith(true)
  })

  it('displays a toast error when setChaptersProgress fails', async () => {
    setChaptersProgress.mockResolvedValue({
      success: false,
      error: { status: 500, message: 'Server error' },
    })

    const { wrapper } = await mount([
      chapter({ id: 'ch-1', pagesNb: 20, progress: undefined }),
    ])

    const items = wrapper.findAllComponents(VListItem)
    await items[1]!.trigger('click')

    expect(useToast().messages.value).toEqual([{ text: 'Unknown error', color: 'error' }])
  })
})
