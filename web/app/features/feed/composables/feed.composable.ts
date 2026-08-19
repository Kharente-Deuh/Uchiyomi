import type { FeedItem } from '../types'
import type { ComicSource, ComicType } from '~/features/comics/types'

const DEFAULT_LIMIT = 15
const DEFAULT_PAGE = 1

export interface FeedComposable {
  items: Ref<FeedItem[]>
  isLoading: Ref<boolean>

  page: Ref<number>
  maxPage: Ref<number>
  source: Ref<ComicSource | undefined>
  type: Ref<ComicType | undefined>
}

export function useFeed(): FeedComposable {
  const api = createFeedApi()
  const chaptersApi = createChaptersApi()

  const { t } = useI18n()
  const toast = useToast()
  const { smAndDown } = useDisplay()

  const items = ref<FeedItem[]>([])

  const isLoading = ref(false)

  const page = ref(1)
  const maxPage = ref(1)
  const source = ref<ComicSource | undefined>(undefined)
  const type = ref<ComicType | undefined>(undefined)

  const getFeedDebounced = useDebounceFn(async (opts?: { clearItems?: boolean }): Promise<void> => {
    isLoading.value = true

    const res = await api.getFeed({
      offset: (page.value - 1) * DEFAULT_LIMIT,
      limit: DEFAULT_LIMIT,
      type: type.value,
      source: source.value,
    })

    if (!res.success) {
      console.error(res.error)
      toast.error(t('feed.error.fetch'))

      items.value = []

      isLoading.value = false

      return
    }

    if (smAndDown.value && !opts?.clearItems) {
      items.value = [...items.value, ...res.data.items]
    } else {
      items.value = res.data.items
    }

    maxPage.value = Math.ceil(res.data.total / DEFAULT_LIMIT)

    isLoading.value = false

    syncPolling()
  }, 200)

  watch([type, source], () => {
    page.value = DEFAULT_PAGE
  })

  watch([page, source, type], (newValues, oldValues) => {
    const [_newPage, newType, newSource] = newValues
    const [_oldPage, oldType, oldSource] = oldValues

    getFeedDebounced({
      clearItems: oldType !== newType || oldSource !== newSource,
    })
  }, { immediate: true })

  const isInFlight = ref(false)

  const { pause, resume } = useIntervalFn(() => {
    void pollChapters()
  }, 2000, { immediate: false })

  function syncPolling(): void {
    if (getChaptersIdsToPoll().length > 0) {
      resume()
    } else {
      pause()
    }
  }

  async function pollChapters(): Promise<void> {
    if (isInFlight.value || isLoading.value) {
      return
    }

    const chaptersIds = getChaptersIdsToPoll()
    if (chaptersIds.length === 0) {
      return
    }

    isInFlight.value = true

    const res = await chaptersApi.getByIds(chaptersIds)
    if (res.success) {
      for (const chapter of res.data) {
        updateChapterDownload(chapter.comicId, chapter.id, chapter.download)
      }
    }

    isInFlight.value = false
    syncPolling()
  }

  function getChaptersIdsToPoll(): string[] {
    const chaptersIds: string[] = []
    for (const item of items.value) {
      for (const chapter of item.latestChapters) {
        if (chapter.download >= 0 && chapter.download < 100) {
          chaptersIds.push(chapter.id)
        }
      }
    }

    return chaptersIds
  }

  function updateChapterDownload(comicId: string, chapterId: string, download: number): void {
    const itemIdx = items.value.findIndex(item => item.id === comicId)
    if (itemIdx === -1) {
      return
    }

    const chapterIdx = items.value[itemIdx]!.latestChapters.findIndex(c => c.id === chapterId)
    if (chapterIdx === -1) {
      return
    }

    if (items.value[itemIdx]!.latestChapters[chapterIdx]!.download === download) {
      return
    }

    items.value[itemIdx]!.latestChapters[chapterIdx]!.download = download
  }

  return {
    items,
    isLoading,
    page,
    maxPage,
    source,
    type,
  }
}
