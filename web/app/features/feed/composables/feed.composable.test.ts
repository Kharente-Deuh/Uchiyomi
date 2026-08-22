// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { FeedItem } from '../types'
import type { Chapter } from '~/features/chapters/types'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope } from 'vue'
import { useToast } from '~/composables/toast.composable'
import { useFeed } from './feed.composable'

const { getFeed, getByIds, smAndDown } = vi.hoisted(() => ({
  getFeed: vi.fn(),
  getByIds: vi.fn(),
  smAndDown: { value: false },
}))

vi.mock('./feed.api', () => ({
  createFeedApi: () => ({ getFeed }),
}))

function i18nStub(): { t: (key: string) => string } {
  return { t: (key: string) => key }
}

function displayStub(): { smAndDown: { value: boolean } } {
  return { smAndDown }
}

function debounceStub<T extends (...args: never[]) => unknown>(fn: T): T {
  return fn
}

function createChaptersApiStub(): { getByIds: (ids: string[]) => Promise<{ success: true, data: Chapter[] }> } {
  return { getByIds }
}

mockNuxtImport('useI18n', () => i18nStub)
mockNuxtImport('useDisplay', () => displayStub)
mockNuxtImport('useDebounceFn', debounceStub)
mockNuxtImport('createChaptersApi', () => createChaptersApiStub)

function item(overrides: Partial<FeedItem> = {}): FeedItem {
  const defaults: FeedItem = {
    id: 'comic-1',
    title: 'Solo Leveling',
    slug: 'solo-leveling',
    cover: '/cover',
    source: 'asurascans',
    status: 'ongoing',
    type: 'manhwa',
    hasProgress: true,
    latestChapters: [{
      id: 'ch-1',
      number: 1,
      download: 100,
      publishedAt: new Date('2026-01-01'),
    }],
  }

  return { ...defaults, ...overrides }
}

function chapter(overrides: Partial<Chapter> = {}): Chapter {
  return {
    id: 'ch-1',
    comicId: 'comic-1',
    publishedAt: new Date('2026-01-01'),
    sourceChapterSlug: 'ch-1',
    title: 'One',
    number: 1,
    pagesNb: 10,
    download: 40,
    ...overrides,
  }
}

function apiError(status: number): { success: false, error: { status: number, message: string } } {
  return { success: false, error: { status, message: 'boom' } }
}

const scopes: ReturnType<typeof effectScope>[] = []

function setup(): ReturnType<typeof useFeed> {
  const scope = effectScope()
  scopes.push(scope)
  const composable = scope.run(() => useFeed())
  if (!composable) {
    throw new Error('useFeed returned undefined')
  }

  return composable
}

beforeEach(() => {
  useToast().messages.value.length = 0
  vi.spyOn(console, 'error').mockImplementation(() => {})
  getFeed.mockReset()
  getByIds.mockReset()
  smAndDown.value = false
  getFeed.mockResolvedValue({ success: true, data: { items: [], total: 0 } })
  vi.useFakeTimers()
})

afterEach(() => {
  while (scopes.length > 0) {
    scopes.pop()?.stop()
  }

  vi.useRealTimers()
})

describe('useFeed', () => {
  it('fetches the first page on setup', async () => {
    getFeed.mockResolvedValue({ success: true, data: { items: [item()], total: 1 } })

    const feed = setup()
    await vi.waitFor(() => expect(feed.items.value).toEqual([item()]))

    expect(getFeed).toHaveBeenCalledWith({
      offset: 0,
      limit: 15,
      type: undefined,
      source: undefined,
    })
    expect(feed.maxPage.value).toBe(1)
    expect(feed.isLoading.value).toBe(false)
  })

  it('converts page 2 to a 15-item offset', async () => {
    getFeed.mockResolvedValue({ success: true, data: { items: [], total: 30 } })
    const feed = setup()
    await vi.waitFor(() => expect(getFeed).toHaveBeenCalled())
    getFeed.mockClear()

    feed.page.value = 2
    await vi.waitFor(() => expect(getFeed).toHaveBeenCalled())

    expect(getFeed).toHaveBeenCalledWith({
      offset: 15,
      limit: 15,
      type: undefined,
      source: undefined,
    })
    expect(feed.maxPage.value).toBe(2)
  })

  it('forwards type and source filters and resets to page 1', async () => {
    getFeed.mockResolvedValue({ success: true, data: { items: [], total: 0 } })
    const feed = setup()
    await vi.waitFor(() => expect(getFeed).toHaveBeenCalled())
    feed.page.value = 2
    await vi.waitFor(() => expect(getFeed).toHaveBeenCalledTimes(2))
    getFeed.mockClear()

    feed.type.value = 'manhwa'
    feed.source.value = 'asurascans'
    await vi.waitFor(() => expect(getFeed).toHaveBeenCalled())

    expect(feed.page.value).toBe(1)
    expect(getFeed).toHaveBeenLastCalledWith({
      offset: 0,
      limit: 15,
      type: 'manhwa',
      source: 'asurascans',
    })
  })

  it('replaces items on desktop when paging', async () => {
    getFeed
      .mockResolvedValueOnce({ success: true, data: { items: [item({ id: 'c1' })], total: 30 } })
      .mockResolvedValueOnce({ success: true, data: { items: [item({ id: 'c2' })], total: 30 } })

    const feed = setup()
    await vi.waitFor(() => expect(feed.items.value.map(i => i.id)).toEqual(['c1']))

    feed.page.value = 2
    await vi.waitFor(() => expect(feed.items.value.map(i => i.id)).toEqual(['c2']))
  })

  it('accumulates pages on small screens', async () => {
    smAndDown.value = true
    getFeed
      .mockResolvedValueOnce({ success: true, data: { items: [item({ id: 'c1' })], total: 30 } })
      .mockResolvedValueOnce({ success: true, data: { items: [item({ id: 'c2' })], total: 30 } })

    const feed = setup()
    await vi.waitFor(() => expect(feed.items.value.map(i => i.id)).toEqual(['c1']))

    feed.page.value = 2
    await vi.waitFor(() => expect(feed.items.value.map(i => i.id)).toEqual(['c1', 'c2']))
  })

  it('replaces accumulated items when filters change on small screens', async () => {
    smAndDown.value = true
    getFeed
      .mockResolvedValueOnce({ success: true, data: { items: [item({ id: 'c1' })], total: 2 } })
      .mockResolvedValueOnce({ success: true, data: { items: [item({ id: 'c2', type: 'manga' })], total: 1 } })

    const feed = setup()
    await vi.waitFor(() => expect(feed.items.value.map(i => i.id)).toEqual(['c1']))

    feed.type.value = 'manga'
    await vi.waitFor(() => expect(feed.items.value.map(i => i.id)).toEqual(['c2']))
  })

  it('toasts and clears results when fetch fails', async () => {
    getFeed.mockResolvedValue(apiError(500))

    const feed = setup()
    await vi.waitFor(() => expect(useToast().messages.value.length).toBe(1))

    expect(useToast().messages.value).toEqual([{ text: 'feed.error.fetch', color: 'error' }])
    expect(feed.items.value).toEqual([])
    expect(feed.isLoading.value).toBe(false)
  })
})

describe('useFeed polling', () => {
  it('does not poll when no chapter is in progress', async () => {
    getFeed.mockResolvedValue({
      success: true,
      data: { items: [item({ latestChapters: [{ id: 'ch-1', number: 1, download: 100, publishedAt: new Date('2026-01-01') }] })], total: 1 },
    })

    setup()
    await vi.waitFor(() => expect(getFeed).toHaveBeenCalled())
    await vi.advanceTimersByTimeAsync(4000)

    expect(getByIds).not.toHaveBeenCalled()
  })

  it('polls in-progress chapter downloads every 2s', async () => {
    getFeed.mockResolvedValue({
      success: true,
      data: {
        items: [item({
          latestChapters: [{ id: 'ch-1', number: 1, download: 10, publishedAt: new Date('2026-01-01') }],
        })],
        total: 1,
      },
    })
    getByIds.mockResolvedValue({ success: true, data: [chapter({ download: 40 })] })

    const feed = setup()
    await vi.waitFor(() => expect(feed.items.value[0]?.latestChapters[0]?.download).toBe(10))

    await vi.advanceTimersByTimeAsync(2000)

    expect(getByIds).toHaveBeenCalledWith(['ch-1'])
    expect(feed.items.value[0]?.latestChapters[0]?.download).toBe(40)
  })

  it('skips a tick while a poll is already in flight', async () => {
    getFeed.mockResolvedValue({
      success: true,
      data: {
        items: [item({
          latestChapters: [{ id: 'ch-1', number: 1, download: 10, publishedAt: new Date('2026-01-01') }],
        })],
        total: 1,
      },
    })
    const poll = Promise.withResolvers<{ success: true, data: Chapter[] }>()
    getByIds.mockImplementationOnce(() => poll.promise)

    setup()
    await vi.waitFor(() => expect(getFeed).toHaveBeenCalled())
    await vi.advanceTimersByTimeAsync(2000)
    await vi.advanceTimersByTimeAsync(2000)

    expect(getByIds).toHaveBeenCalledTimes(1)

    poll.resolve({ success: true, data: [chapter({ download: 20 })] })
    await Promise.resolve()
  })

  it('does not rewrite download when the poll returns the same value', async () => {
    const feedItem = item({
      latestChapters: [{ id: 'ch-1', number: 1, download: 10, publishedAt: new Date('2026-01-01') }],
    })
    getFeed.mockResolvedValue({ success: true, data: { items: [feedItem], total: 1 } })
    getByIds.mockResolvedValue({ success: true, data: [chapter({ download: 10 })] })

    const feed = setup()
    await vi.waitFor(() => expect(feed.items.value[0]?.latestChapters[0]?.download).toBe(10))
    const before = feed.items.value[0]!.latestChapters[0]

    await vi.advanceTimersByTimeAsync(2000)

    expect(feed.items.value[0]!.latestChapters[0]).toBe(before)
  })

  it('stops polling once every chapter is complete', async () => {
    getFeed.mockResolvedValue({
      success: true,
      data: {
        items: [item({
          latestChapters: [{ id: 'ch-1', number: 1, download: 90, publishedAt: new Date('2026-01-01') }],
        })],
        total: 1,
      },
    })
    getByIds.mockResolvedValue({ success: true, data: [chapter({ download: 100 })] })

    setup()
    await vi.waitFor(() => expect(getFeed).toHaveBeenCalled())
    await vi.advanceTimersByTimeAsync(2000)
    getByIds.mockClear()
    await vi.advanceTimersByTimeAsync(4000)

    expect(getByIds).not.toHaveBeenCalled()
  })
})
