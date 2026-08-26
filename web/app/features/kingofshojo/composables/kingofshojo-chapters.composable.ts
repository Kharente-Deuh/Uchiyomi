// SPDX-License-Identifier: AGPL-3.0-or-later

import type { KingOfShojoComicChapter } from '../types'
import { useKingOfShojoChaptersStore } from '../stores/kingofshojo-chapters.store'

const POLL_INTERVAL_MS = 2000

export interface KingOfShojoChaptersComposable {
  sort: Ref<'asc' | 'desc'>
  chapters: Ref<KingOfShojoComicChapter[]>
  loading: Ref<boolean>
  fetchChapters: () => Promise<void>
  retryDownload: (chapterId: string) => Promise<void>
  retryDownloadLoading: Ref<boolean>
}

export function useKingOfShojoChapters(): KingOfShojoChaptersComposable {
  const store = useKingOfShojoChaptersStore()
  const api = createKingOfShojoApi()
  const chaptersApi = createChaptersApi()
  const toast = useToast()
  const { t } = useI18n()
  const route = useRoute('browse-sources-kingofshojo-slug')

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
      toast.error(t('sources.kingofshojo.comic.chapters.error.fetch'))
    } else {
      toast.error(t('error.unknown'))
    }
  }

  const retryDownloadLoading = ref(false)
  async function retryDownload(chapterId: string): Promise<void> {
    if (!chapterId) {
      return
    }

    retryDownloadLoading.value = true

    const res = await chaptersApi.retryDownload(chapterId)
    if (res.success) {
      await loadChapters(false)
      syncPolling()

      retryDownloadLoading.value = false

      return
    }

    console.error('chaptersApi.retryDownload', res.error)

    switch (res.error.status) {
      case 404:
        toast.error(t('sources.kingofshojo.comic.chapters.error.retry.notFound'))
        break
      case 403:
        toast.error(t('sources.kingofshojo.comic.chapters.error.retry.forbidden'))
        break
      case 409:
        toast.error(t('sources.kingofshojo.comic.chapters.error.retry.conflict'))
        break
      default:
        toast.error(t('error.unknown'))
        break
    }

    retryDownloadLoading.value = false
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
    retryDownload,
    retryDownloadLoading,
  }
}
