// SPDX-License-Identifier: AGPL-3.0-or-later

import type { AsuraComicChapter } from '../types'
import { useAsuraChaptersStore } from '../stores/asura-chapters.store'

const POLL_INTERVAL_MS = 2000

export interface AsuraChaptersComposable {
  sort: Ref<'asc' | 'desc'>
  chapters: Ref<AsuraComicChapter[]>
  loading: Ref<boolean>
  fetchChapters: () => Promise<void>
}

export function isChapterDownloadInProgress(chapter: AsuraComicChapter): boolean {
  return chapter.download !== undefined && chapter.download >= 0 && chapter.download < 100
}

export function useAsuraChapters(): AsuraChaptersComposable {
  const store = useAsuraChaptersStore()
  const api = createAsuraApi()
  const toast = useToast()
  const { t } = useI18n()
  const route = useRoute('browse-sources-asura-slug')

  const sort = ref<'asc' | 'desc'>('desc')
  const chapters = computed(() => store.chapters.toSorted((a, b) => {
    if (sort.value === 'asc') {
      return a.number - b.number
    }

    return b.number - a.number
  }))

  const loading = ref(false)
  let isInFlight = false

  async function loadChapters(shouldNotify: boolean): Promise<void> {
    const res = await api.getSeriesChapters(route.params.slug)

    if (res.success) {
      store.setChapters(res.data)

      return
    }

    console.error('api.getSeriesChapters', res.error)

    if (!shouldNotify) {
      return
    }

    if (res.error.status === 404) {
      toast.error(t('sources.asura.comic.chapters.error.fetch'))
    } else {
      toast.error(t('error.unknown'))
    }
  }

  async function pollChapters(): Promise<void> {
    if (isInFlight || loading.value) {
      return
    }

    isInFlight = true
    await loadChapters(false)
    isInFlight = false
    syncPolling()
  }

  const { pause, resume } = useIntervalFn(() => {
    void pollChapters()
  }, POLL_INTERVAL_MS, { immediate: false })

  function syncPolling(): void {
    if (store.chapters.some(chapter => isChapterDownloadInProgress(chapter))) {
      resume()
    } else {
      pause()
    }
  }

  async function fetchChapters(): Promise<void> {
    loading.value = true
    await loadChapters(true)
    loading.value = false
    syncPolling()
  }

  onBeforeRouteLeave(() => {
    pause()
    store.invalidate()
  })

  return {
    sort,
    chapters,
    loading,
    fetchChapters,
  }
}
