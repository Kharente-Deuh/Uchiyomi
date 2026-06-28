// SPDX-License-Identifier: AGPL-3.0-or-later

import type { MangaChapterSummaryDto } from '#shared/dto/catalogue/manga-chapter-summary.dto'
import type { ApiResponse } from '~/utils/api'
import { flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const { getMangaChapterSummary } = vi.hoisted(() => ({ getMangaChapterSummary: vi.fn() }))
vi.mock('~/features/extensions/api/extensions.api', () => ({
  createExtensionsApi: () => ({ getMangaChapterSummary }),
}))

const { useChapterSummariesStore } = await import('~/features/extensions/store/chapter-summaries.store')

const CTX = { pkgName: 'p', sourceId: '9' }

function ok(data: MangaChapterSummaryDto): ApiResponse<MangaChapterSummaryDto> {
  return { success: true, data }
}

function deferred<T>(): { promise: Promise<T>, resolve: (v: T) => void } {
  let resolve!: (v: T) => void
  const promise = new Promise<T>((r) => {
    resolve = r
  })

  return { promise, resolve }
}

beforeEach(() => setActivePinia(createPinia()))
afterEach(() => getMangaChapterSummary.mockReset())

describe('useChapterSummariesStore', () => {
  it('enqueues, fetches, caches and exposes the summary', async () => {
    getMangaChapterSummary.mockResolvedValueOnce(ok({ chapterCount: 3, lastChapter: null }))
    const store = useChapterSummariesStore()

    store.sync(['1'], CTX)
    expect(store.statusOf('1')).toBe('loading')

    await flushPromises()
    expect(store.statusOf('1')).toBe('success')
    expect(store.summaryOf('1')).toEqual({ chapterCount: 3, lastChapter: null })
  })

  it('does not refetch an id already in cache', async () => {
    getMangaChapterSummary.mockResolvedValueOnce(ok({ chapterCount: 1, lastChapter: null }))
    const store = useChapterSummariesStore()

    store.sync(['1'], CTX)
    await flushPromises()
    store.sync(['1'], CTX)

    expect(getMangaChapterSummary).toHaveBeenCalledTimes(1)
  })

  it('never runs more than 5 workers at once', () => {
    getMangaChapterSummary.mockImplementation(() => deferred<ApiResponse<MangaChapterSummaryDto>>().promise)
    const store = useChapterSummariesStore()

    store.sync(['1', '2', '3', '4', '5', '6', '7'], CTX)

    // The 5 workers call the api synchronously up to their await; 6 and 7 wait.
    expect(getMangaChapterSummary).toHaveBeenCalledTimes(5)
    expect(store.statusOf('5')).toBe('loading')
    expect(store.statusOf('6')).toBe('queued')
    expect(store.statusOf('7')).toBe('queued')
  })

  it('starts a queued id when an in-flight one resolves', async () => {
    const gates = new Map<string, ReturnType<typeof deferred<ApiResponse<MangaChapterSummaryDto>>>>()
    getMangaChapterSummary.mockImplementation((_p: string, _s: string, id: string) => {
      const d = deferred<ApiResponse<MangaChapterSummaryDto>>()
      gates.set(id, d)

      return d.promise
    })
    const store = useChapterSummariesStore()

    store.sync(['1', '2', '3', '4', '5', '6'], CTX)
    expect(store.statusOf('6')).toBe('queued')

    gates.get('1')!.resolve(ok({ chapterCount: 1, lastChapter: null }))
    await flushPromises()

    expect(store.statusOf('1')).toBe('success')
    expect(store.statusOf('6')).toBe('loading')
  })

  it('aborts in-flight ids dropped from the new sync set', () => {
    const signals = new Map<string, AbortSignal>()
    getMangaChapterSummary.mockImplementation((_p: string, _s: string, id: string, opts: { signal: AbortSignal }) => {
      signals.set(id, opts.signal)

      return deferred<ApiResponse<MangaChapterSummaryDto>>().promise
    })
    const store = useChapterSummariesStore()

    store.sync(['1', '2'], CTX)
    expect(store.statusOf('1')).toBe('loading')

    store.sync(['2', '3'], CTX)

    expect(signals.get('1')!.aborted).toBe(true)
    expect(store.statusOf('1')).toBeUndefined()
    expect(store.statusOf('2')).toBe('loading')
    expect(store.statusOf('3')).toBe('loading')
  })

  it('marks a failed id error and retries it on the next sync', async () => {
    getMangaChapterSummary.mockResolvedValueOnce({ success: false, error: { status: 500 } } as ApiResponse<MangaChapterSummaryDto>)
    const store = useChapterSummariesStore()

    store.sync(['1'], CTX)
    await flushPromises()
    expect(store.statusOf('1')).toBe('error')

    getMangaChapterSummary.mockResolvedValueOnce(ok({ chapterCount: 2, lastChapter: null }))
    store.sync(['1'], CTX)
    await flushPromises()
    expect(store.statusOf('1')).toBe('success')
    expect(getMangaChapterSummary).toHaveBeenCalledTimes(2)
  })

  it('empties the queue and aborts everything on sync([])', () => {
    const signals = new Map<string, AbortSignal>()
    getMangaChapterSummary.mockImplementation((_p: string, _s: string, id: string, opts: { signal: AbortSignal }) => {
      signals.set(id, opts.signal)

      return deferred<ApiResponse<MangaChapterSummaryDto>>().promise
    })
    const store = useChapterSummariesStore()

    store.sync(['1', '2', '6', '7'], CTX)
    store.sync([], CTX)

    expect(store.queue).toEqual([])
    expect(signals.get('1')!.aborted).toBe(true)
    expect(store.statusOf('1')).toBeUndefined()
    expect(store.statusOf('7')).toBeUndefined()
  })

  it('an aborted worker never writes a status after its promise settles', async () => {
    const gates = new Map<string, ReturnType<typeof deferred<ApiResponse<MangaChapterSummaryDto>>>>()
    getMangaChapterSummary.mockImplementation((_p: string, _s: string, id: string) => {
      const d = deferred<ApiResponse<MangaChapterSummaryDto>>()
      gates.set(id, d)

      return d.promise
    })
    const store = useChapterSummariesStore()

    store.sync(['1', '2'], CTX)
    expect(store.statusOf('1')).toBe('loading')
    expect(store.statusOf('2')).toBe('loading')

    // Drop id '1' — its worker is aborted and its status is cleared.
    store.sync(['2'], CTX)
    expect(store.statusOf('1')).toBeUndefined()

    // Resolve the already-aborted promise for '1'.
    gates.get('1')!.resolve(ok({ chapterCount: 9, lastChapter: null }))
    await flushPromises()

    // The aborted worker must not write anything.
    expect(store.statusOf('1')).toBeUndefined()
    expect(store.summaryOf('1')).toBeUndefined()
  })

  it('aborting an in-flight id frees its slot for a queued id', async () => {
    const gates = new Map<string, ReturnType<typeof deferred<ApiResponse<MangaChapterSummaryDto>>>>()
    getMangaChapterSummary.mockImplementation((_p: string, _s: string, id: string) => {
      const d = deferred<ApiResponse<MangaChapterSummaryDto>>()
      gates.set(id, d)

      return d.promise
    })
    const store = useChapterSummariesStore()

    // 5 workers busy, '6' is queued.
    store.sync(['1', '2', '3', '4', '5', '6'], CTX)
    expect(store.statusOf('6')).toBe('queued')

    // Drop '1': its slot is freed and pump() should start '6'.
    store.sync(['2', '3', '4', '5', '6'], CTX)

    expect(store.statusOf('1')).toBeUndefined()
    expect(store.statusOf('6')).toBe('loading')
  })
})
