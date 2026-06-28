// SPDX-License-Identifier: AGPL-3.0-or-later

import type { MangaChapterSummaryDto } from '#shared/dto/catalogue/manga-chapter-summary.dto'
import { createExtensionsApi } from '../api/extensions.api'

export type ChapterSummaryStatus = 'queued' | 'loading' | 'success' | 'error'

interface SyncContext {
  pkgName: string
  sourceId: string
}

// Bounded worker pool. Bounding does not speed up the total (Suwayomi serialises
// the scrapes) but keeps each in-flight request short enough to avoid the BFF
// timeout pileup, and gives per-request cancellation granularity.
const CONCURRENCY = 5

export const useChapterSummariesStore = defineStore('chapter-summaries', () => {
  const api = createExtensionsApi()

  // Successes only; key = mangaId (globally unique across sources → safe cross-filter).
  const cache = ref(new Map<string, MangaChapterSummaryDto>())
  const status = ref(new Map<string, ChapterSummaryStatus>())
  const queue = ref<string[]>([])

  // Internal control state — not reactive, never rendered.
  const inFlight = new Map<string, AbortController>()
  // Queue only ever holds ids for the current ctx (sync prunes on every change),
  // so reading the latest ctx in a worker is always the right one.
  let ctx: SyncContext | undefined

  function summaryOf(id: string): MangaChapterSummaryDto | undefined {
    return cache.value.get(id)
  }

  function statusOf(id: string): ChapterSummaryStatus | undefined {
    return status.value.get(id)
  }

  function sync(ids: string[], context: SyncContext): void {
    ctx = context
    const wanted = new Set(ids)

    // Cancel in-flight ids that are no longer wanted, freeing their workers.
    for (const [id, controller] of inFlight) {
      if (!wanted.has(id)) {
        controller.abort()
        inFlight.delete(id)
        status.value.delete(id)
      }
    }

    // Drop queued ids that are no longer wanted.
    queue.value = queue.value.filter(id => wanted.has(id))

    // Enqueue wanted ids with no fresh state yet (cached/loading/queued are skipped).
    for (const id of ids) {
      const current = status.value.get(id)
      if (current === 'success' || current === 'loading' || current === 'queued') {
        continue
      }

      status.value.set(id, 'queued')
      queue.value.push(id)
    }

    pump()
  }

  function pump(): void {
    while (inFlight.size < CONCURRENCY && queue.value.length > 0) {
      const id = queue.value.shift()!
      void runWorker(id)
    }
  }

  async function runWorker(id: string): Promise<void> {
    if (!ctx) {
      return
    }

    const controller = new AbortController()
    inFlight.set(id, controller)
    status.value.set(id, 'loading')

    const res = await api.getMangaChapterSummary(ctx.pkgName, ctx.sourceId, id, { signal: controller.signal })

    // Cancelled mid-flight: sync already removed it from inFlight and cleared its
    // status. Leave it alone — it is no longer relevant.
    if (controller.signal.aborted) {
      return
    }

    inFlight.delete(id)

    if (res.success) {
      cache.value.set(id, res.data)
      status.value.set(id, 'success')
    } else {
      // Not cached → retried if this id re-enters a later sync.
      status.value.set(id, 'error')
    }

    pump()
  }

  return { cache, status, queue, sync, summaryOf, statusOf }
})
