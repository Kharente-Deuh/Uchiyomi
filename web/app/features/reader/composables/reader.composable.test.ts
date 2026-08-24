// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { VueWrapper } from '@vue/test-utils'
import type { Ref } from 'vue'
import type { ReaderComposable } from './reader.composable'
import type { DetailedChapter } from '~/features/chapters/types'
import type { Comic } from '~/features/comics/types'
import type { ReaderSettings } from '~/features/reader/types'
import { mockNuxtImport, mountSuspended } from '@nuxt/test-utils/runtime'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { useToast } from '~/composables/toast.composable'
import { ASURA_SOURCE_NAME } from '~/constants'
import { useReader } from './reader.composable'

const {
  getById,
  retryDownload,
  saveProgress,
  getComicById,
  getReaderSettings,
  navigateTo,
} = vi.hoisted(() => ({
  getById: vi.fn(),
  retryDownload: vi.fn(),
  saveProgress: vi.fn(),
  getComicById: vi.fn(),
  getReaderSettings: vi.fn(),
  navigateTo: vi.fn(),
}))

function createChaptersApiStub(): {
  getById: typeof getById
  retryDownload: typeof retryDownload
  saveProgress: typeof saveProgress
} {
  return { getById, retryDownload, saveProgress }
}

function createComicsApiStub(): { getById: typeof getComicById } {
  return { getById: getComicById }
}

function createReaderSettingsApiStub(): { getReaderSettings: typeof getReaderSettings } {
  return { getReaderSettings }
}

mockNuxtImport('createChaptersApi', () => createChaptersApiStub)
mockNuxtImport('createComicsApi', () => createComicsApiStub)
mockNuxtImport('createReaderSettingsApi', () => createReaderSettingsApiStub)
mockNuxtImport('navigateTo', () => navigateTo)

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

function apiError(status: number): { success: false, error: { status: number, message: string } } {
  return { success: false, error: { status, message: 'boom' } }
}

function ok<T>(data: T): { success: true, data: T } {
  return { success: true, data }
}

function chapterResponse(overrides: Partial<DetailedChapter> = {}): { success: true, data: DetailedChapter } {
  return ok(chapter(overrides))
}

async function waitUntilIdle(reader: ReaderComposable): Promise<void> {
  await vi.waitFor(() => {
    expect(reader.isLoading.value).toBe(false)
  })
}

const wrappers: VueWrapper[] = []

async function setup(
  chapterId = 'ch-2',
  opts: { ignoreProgress?: boolean, onAfterProgressIgnored?: () => void } = {},
): Promise<{ reader: ReaderComposable, chapterId: Ref<string> }> {
  const id = ref(chapterId)
  const ignoreProgress = ref(opts.ignoreProgress ?? false)
  let reader: ReaderComposable | undefined

  const wrapper = await mountSuspended({
    setup: () => {
      reader = useReader({
        comicId: 'c1',
        chapterId: id,
        ignoreProgress,
        onAfterProgressIgnored: opts.onAfterProgressIgnored,
      })

      return { reader }
    },
    render: () => null,
  })
  wrappers.push(wrapper)

  if (!reader) {
    throw new Error('useReader returned undefined')
  }

  return { reader, chapterId: id }
}

beforeEach(() => {
  useToast().messages.value.length = 0
  vi.spyOn(console, 'error').mockImplementation(() => {})
  getById.mockReset()
  retryDownload.mockReset()
  saveProgress.mockReset()
  getComicById.mockReset()
  getReaderSettings.mockReset()
  navigateTo.mockReset()
  getById.mockImplementation(async (id: string) => chapterResponse({ id, sourceChapterSlug: id, title: id }))
  getComicById.mockResolvedValue(ok(comic()))
  getReaderSettings.mockResolvedValue(ok([settings()]))
  saveProgress.mockResolvedValue(ok({ page: 0, updatedAt: new Date('2026-08-22') }))
  retryDownload.mockResolvedValue(ok(undefined))
})

afterEach(() => {
  while (wrappers.length > 0) {
    wrappers.pop()?.unmount()
  }
})

describe('useReader', () => {
  it('loads the chapter, comic and matching reader settings', async () => {
    const { reader } = await setup()

    await waitUntilIdle(reader)

    expect(getById).toHaveBeenCalledWith('ch-2')
    expect(getComicById).toHaveBeenCalledWith('c1')
    expect(reader.chapter.value?.id).toBe('ch-2')
    expect(reader.comic.value?.id).toBe('c1')
    expect(reader.readerSettings.value).toEqual(settings())
  })

  it('restores the saved page after load', async () => {
    getById.mockResolvedValue(chapterResponse({
      progress: { page: 2, updatedAt: new Date('2026-08-22') },
    }))

    const { reader } = await setup()
    await waitUntilIdle(reader)

    expect(reader.page.value).toBe(1)
  })

  it('skips restored progress when ignoreProgress is set', async () => {
    const onAfterProgressIgnored = vi.fn()
    getById.mockResolvedValue(chapterResponse({
      progress: { page: 2, updatedAt: new Date('2026-08-22') },
    }))

    const { reader } = await setup('ch-2', { ignoreProgress: true, onAfterProgressIgnored })
    await waitUntilIdle(reader)

    expect(reader.page.value).toBe(0)
    expect(onAfterProgressIgnored).toHaveBeenCalledOnce()
  })

  it('redirects to the library when the comic is missing', async () => {
    getComicById.mockResolvedValue(apiError(404))

    await setup()
    await vi.waitFor(() => expect(navigateTo).toHaveBeenCalledWith('/library'))

    expect(useToast().messages.value).toEqual([{ text: 'Comic not found', color: 'error' }])
  })

  it('redirects to the comic when the chapter is missing', async () => {
    getById.mockResolvedValue(apiError(404))

    await setup()
    await vi.waitFor(() => expect(navigateTo).toHaveBeenCalledWith('/comic/c1'))

    expect(useToast().messages.value).toEqual([{ text: 'Chapter not found', color: 'error' }])
  })

  it('redirects to the comic when no settings match the comic type', async () => {
    getReaderSettings.mockResolvedValue(ok([settings({ type: 'manga' })]))

    await setup()
    await vi.waitFor(() => expect(navigateTo).toHaveBeenCalledWith('/comic/c1'))
  })

  it('prefills the next chapter without changing the current one', async () => {
    const { reader } = await setup()
    await waitUntilIdle(reader)
    getById.mockClear()

    reader.fetchNextChapter()
    await vi.waitFor(() => expect(reader.nextChapter.value?.id).toBe('ch-3'))

    expect(getById).toHaveBeenCalledWith('ch-3')
    expect(reader.chapter.value?.id).toBe('ch-2')
  })

  it('opens a prefetched previous chapter on its last page', async () => {
    getById.mockImplementation(async (id: string) => {
      return chapterResponse({
        id,
        sourceChapterSlug: id,
        title: id,
        pagesNb: 5,
        pageUrls: ['/p1', '/p2', '/p3', '/p4', '/p5'],
      })
    })
    const { reader, chapterId } = await setup()
    await waitUntilIdle(reader)

    reader.fetchPreviousChapter()
    await vi.waitFor(() => expect(reader.previousChapter.value?.id).toBe('ch-1'))

    reader.startEnd.value = true
    chapterId.value = 'ch-1'
    await vi.waitFor(() => expect(reader.chapter.value?.id).toBe('ch-1'))

    expect(reader.page.value).toBe(4)
  })

  it('retries a failed download and reloads the chapter', async () => {
    getById
      .mockResolvedValueOnce(chapterResponse({ download: -1 }))
      .mockResolvedValueOnce(chapterResponse({ download: 0 }))
    const { reader } = await setup()
    await waitUntilIdle(reader)

    await reader.retryDownload()

    expect(retryDownload).toHaveBeenCalledWith('ch-2')
    expect(reader.chapter.value?.download).toBe(0)
  })

  it('saves progress when the page advances past the stored value', async () => {
    getById.mockResolvedValue(chapterResponse({
      progress: { page: 0, updatedAt: new Date('2026-08-22') },
    }))
    const { reader } = await setup()
    await waitUntilIdle(reader)
    saveProgress.mockClear()

    reader.page.value = 1
    await vi.waitFor(() => expect(saveProgress).toHaveBeenCalledWith({ id: 'ch-2', page: 2 }))
  })

  it('saves the right-hand page when double page is on', async () => {
    getReaderSettings.mockResolvedValue(ok([settings({ doublePage: true })]))
    getById.mockResolvedValue(chapterResponse({
      pagesNb: 4,
      pageUrls: ['/p1', '/p2', '/p3', '/p4'],
      progress: { page: 0, updatedAt: new Date('2026-08-22') },
    }))
    const { reader } = await setup()
    await waitUntilIdle(reader)
    saveProgress.mockClear()

    reader.page.value = 2
    await vi.waitFor(() => expect(saveProgress).toHaveBeenCalledWith({ id: 'ch-2', page: 4 }))
  })

  it('evicts the oldest chapter once more than three are loaded', async () => {
    getById.mockImplementation(async (id: string) => {
      const n = Number(id.replace('ch-', ''))

      return chapterResponse({
        id,
        sourceChapterSlug: id,
        title: id,
        number: n,
        next: { id: `ch-${n + 1}`, title: `ch-${n + 1}`, number: n + 1 },
        previous: n > 1 ? { id: `ch-${n - 1}`, title: `ch-${n - 1}`, number: n - 1 } : undefined,
      })
    })
    const { reader, chapterId } = await setup('ch-1')
    await waitUntilIdle(reader)

    reader.fetchNextChapter()
    await vi.waitFor(() => expect(reader.nextChapter.value?.id).toBe('ch-2'))
    chapterId.value = 'ch-2'
    await vi.waitFor(() => expect(reader.chapter.value?.id).toBe('ch-2'))

    reader.fetchNextChapter()
    await vi.waitFor(() => expect(reader.nextChapter.value?.id).toBe('ch-3'))
    chapterId.value = 'ch-3'
    await vi.waitFor(() => expect(reader.chapter.value?.id).toBe('ch-3'))

    reader.fetchNextChapter()
    await vi.waitFor(() => expect(reader.nextChapter.value?.id).toBe('ch-4'))

    getById.mockClear()
    chapterId.value = 'ch-1'
    await vi.waitFor(() => expect(getById).toHaveBeenCalledWith('ch-1'))
  })
})

describe('useReader polling', () => {
  it('polls the current chapter while download is in progress', async () => {
    getById.mockResolvedValue(chapterResponse({ download: 10 }))
    const { reader } = await setup()
    await waitUntilIdle(reader)
    getById.mockClear()
    getById.mockResolvedValue(chapterResponse({ download: 40 }))

    await vi.waitFor(() => expect(getById).toHaveBeenCalledWith('ch-2'), { timeout: 3000 })
    await vi.waitFor(() => expect(reader.chapter.value?.download).toBe(40))
  })

  it('does not reset the page when a poll refreshes the same chapter', async () => {
    getById.mockResolvedValue(chapterResponse({
      download: 10,
      progress: { page: 2, updatedAt: new Date('2026-08-22') },
    }))
    const { reader } = await setup()
    await waitUntilIdle(reader)
    expect(reader.page.value).toBe(1)
    getById.mockResolvedValue(chapterResponse({ download: 40 }))

    await vi.waitFor(() => expect(reader.chapter.value?.download).toBe(40), { timeout: 3000 })
    expect(reader.page.value).toBe(1)
  })

  it('does not poll once the download is complete', async () => {
    const { reader } = await setup()
    await waitUntilIdle(reader)
    getById.mockClear()

    await new Promise(resolve => setTimeout(resolve, 2500))

    expect(getById).not.toHaveBeenCalled()
  })
})
