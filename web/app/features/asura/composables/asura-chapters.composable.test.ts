// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { AsuraComicChapter } from '../types'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope } from 'vue'
import { useToast } from '~/composables/toast.composable'
import { useAsuraChaptersStore } from '../stores/asura-chapters.store'
import { isChapterDownloadInProgress, useAsuraChapters } from './asura-chapters.composable'

const { getSeriesChapters } = vi.hoisted(() => ({
  getSeriesChapters: vi.fn(),
}))

vi.mock('./asura.api', () => ({
  createAsuraApi: () => ({ getSeriesChapters }),
}))

function i18nStub(): { t: (key: string) => string } {
  return { t: (key: string) => key }
}

function routeStub(): { params: { slug: string } } {
  return { params: { slug: 'solo-leveling' } }
}

mockNuxtImport('useI18n', () => i18nStub)
mockNuxtImport('useRoute', () => routeStub)

function chapter(overrides: Partial<AsuraComicChapter> = {}): AsuraComicChapter {
  return {
    id: 'ch-1',
    title: 'One',
    number: 1,
    publishedAt: new Date('2026-01-01'),
    earlyAccessUntil: new Date(0),
    ...overrides,
  }
}

function apiError(status: number): { success: false, error: { status: number, message: string } } {
  return { success: false, error: { status, message: 'boom' } }
}

const scopes: ReturnType<typeof effectScope>[] = []

function setup(): ReturnType<typeof useAsuraChapters> {
  const scope = effectScope()
  scopes.push(scope)
  const composable = scope.run(() => useAsuraChapters())
  if (!composable) {
    throw new Error('useAsuraChapters returned undefined')
  }

  return composable
}

beforeEach(() => {
  setActivePinia(createPinia())
  useToast().messages.value.length = 0
  vi.spyOn(console, 'error').mockImplementation(() => {})
  getSeriesChapters.mockReset()
  vi.useFakeTimers()
})

afterEach(() => {
  while (scopes.length > 0) {
    scopes.pop()?.stop()
  }

  vi.useRealTimers()
})

describe('isChapterDownloadInProgress', () => {
  it('is true only for a defined progress between 0 and 99', () => {
    expect(isChapterDownloadInProgress(chapter())).toBe(false)
    expect(isChapterDownloadInProgress(chapter({ download: -1 }))).toBe(false)
    expect(isChapterDownloadInProgress(chapter({ download: 0 }))).toBe(true)
    expect(isChapterDownloadInProgress(chapter({ download: 42 }))).toBe(true)
    expect(isChapterDownloadInProgress(chapter({ download: 99 }))).toBe(true)
    expect(isChapterDownloadInProgress(chapter({ download: 100 }))).toBe(false)
  })
})

describe('useAsuraChapters polling', () => {
  it('toasts on the initial fetch and does not poll', async () => {
    getSeriesChapters.mockResolvedValue(apiError(500))

    await setup().fetchChapters()
    await vi.advanceTimersByTimeAsync(4000)

    expect(getSeriesChapters).toHaveBeenCalledTimes(1)
    expect(useToast().messages.value).toEqual([{ text: 'error.unknown', color: 'error' }])
  })

  it('does not poll when no chapter is in progress', async () => {
    getSeriesChapters.mockResolvedValue({
      success: true,
      data: [
        chapter({ download: 100 }),
        chapter({ id: 'ch-2', number: 2, download: -1 }),
        chapter({ id: 'ch-3', number: 3 }),
      ],
    })

    await setup().fetchChapters()
    await vi.advanceTimersByTimeAsync(4000)

    expect(getSeriesChapters).toHaveBeenCalledTimes(1)
  })

  it('polls every 2s without loading or toasts while a download is in progress', async () => {
    getSeriesChapters
      .mockResolvedValueOnce({ success: true, data: [chapter({ download: 10 })] })
      .mockResolvedValueOnce({ success: true, data: [chapter({ download: 40 })] })

    const asura = setup()
    await asura.fetchChapters()
    expect(asura.loading.value).toBe(false)

    await vi.advanceTimersByTimeAsync(2000)

    expect(getSeriesChapters).toHaveBeenCalledTimes(2)
    expect(asura.loading.value).toBe(false)
    expect(useAsuraChaptersStore().chapters[0]?.download).toBe(40)
    expect(useToast().messages.value).toEqual([])
  })

  it('keeps the last list and does not toast when a poll fails', async () => {
    getSeriesChapters
      .mockResolvedValueOnce({ success: true, data: [chapter({ download: 10 })] })
      .mockResolvedValueOnce(apiError(500))

    await setup().fetchChapters()
    useToast().messages.value.length = 0

    await vi.advanceTimersByTimeAsync(2000)

    expect(useAsuraChaptersStore().chapters[0]?.download).toBe(10)
    expect(useToast().messages.value).toEqual([])
  })

  it('stops polling once no chapter is in progress', async () => {
    getSeriesChapters
      .mockResolvedValueOnce({ success: true, data: [chapter({ download: 90 })] })
      .mockResolvedValueOnce({ success: true, data: [chapter({ download: 100 })] })

    await setup().fetchChapters()
    await vi.advanceTimersByTimeAsync(2000)
    await vi.advanceTimersByTimeAsync(4000)

    expect(getSeriesChapters).toHaveBeenCalledTimes(2)
  })

  it('skips a tick while a fetch is already in flight', async () => {
    getSeriesChapters.mockResolvedValueOnce({ success: true, data: [chapter({ download: 10 })] })

    const poll = Promise.withResolvers<{ success: true, data: AsuraComicChapter[] }>()
    getSeriesChapters.mockImplementationOnce(() => poll.promise)

    await setup().fetchChapters()
    await vi.advanceTimersByTimeAsync(2000)
    await vi.advanceTimersByTimeAsync(2000)

    expect(getSeriesChapters).toHaveBeenCalledTimes(2)

    poll.resolve({ success: true, data: [chapter({ download: 20 })] })
    await Promise.resolve()
  })
})
