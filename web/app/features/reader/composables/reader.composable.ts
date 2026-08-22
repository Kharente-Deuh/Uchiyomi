import type { ReaderSettings } from '../types'
import type { DetailedChapter } from '~/features/chapters/types'
import type { Comic } from '~/features/comics/types'

const MAX_LOADED_CHAPTERS = 3

export interface ReaderComposable {
  page: Ref<number>
  chapter: Ref<DetailedChapter | undefined>
  comic: Ref<Comic | undefined>
  readerSettings: Ref<ReaderSettings | undefined>
  isLoading: Ref<boolean>
  fetchNextChapter: () => void
  fetchPreviousChapter: () => void
  retryDownload: () => void
  retryDownloadLoading: Ref<boolean>
  nextChapter: Ref<DetailedChapter | undefined>
  previousChapter: Ref<DetailedChapter | undefined>
  startEnd: Ref<boolean>
}

export function useReader(comicId: string, chapterId: Ref<string>): ReaderComposable {
  const api = createChaptersApi()
  const comicsApi = createComicsApi()
  const readerSettingsApi = createReaderSettingsApi()
  const { t } = useI18n()
  const toast = useToast()

  const loadedChapters = ref<DetailedChapter[]>([])
  const chapter = computed(() => loadedChapters.value.find(({ id }) => id === chapterId.value))
  const comic = ref<Comic>()
  const isLoading = ref(true)
  const fetchChapterLoading = ref(false)

  const readerSettings = ref<ReaderSettings>()
  const page = ref(0)
  const startEnd = ref(false)

  async function fetchChapter(id: string): Promise<void> {
    fetchChapterLoading.value = true
    const res = await api.getById(id)
    if (!res.success) {
      console.error('api.getById', res.error)

      if (res.error.status === 404 || res.error.status === 403) {
        toast.error(t('sources.asura.chapter.notFound'))
      } else {
        toast.error(t('error.unknown'))
      }

      fetchChapterLoading.value = false

      return
    }

    const index = loadedChapters.value.findIndex(c => c.id === id)
    if (index === -1) {
      loadedChapters.value.push(res.data)
      if (loadedChapters.value.length > MAX_LOADED_CHAPTERS) {
        loadedChapters.value.shift()
      }
    } else {
      loadedChapters.value[index] = res.data
    }

    fetchChapterLoading.value = false
  }

  async function fetchReaderSettings(): Promise<ReaderSettings[]> {
    const res = await readerSettingsApi.getReaderSettings()
    if (!res.success) {
      console.error('readerSettingsApi.get', res.error)
      toast.error(t('error.unknown'))

      return []
    }

    return res.data
  }

  async function fetchComic(): Promise<void> {
    const res = await comicsApi.getById(comicId)
    if (!res.success) {
      console.error('comicsApi.getById', res.error)

      if (res.error.status === 404 || res.error.status === 403) {
        toast.error(t('sources.asura.comic.notFound'))
      } else {
        toast.error(t('error.unknown'))
      }

      return
    }

    comic.value = res.data
  }

  onMounted(async () => {
    const [_, __, settings] = await Promise.all([
      fetchChapter(chapterId.value),
      fetchComic(),
      fetchReaderSettings(),
    ])

    if (!comic.value) {
      navigateTo('/library')

      return
    }

    if (!chapter.value) {
      navigateTo(`/comic/${comicId}`)

      return
    }

    readerSettings.value = settings.find(({ type }) => type === comic.value?.type)
    if (!readerSettings.value) {
      navigateTo(`/comic/${comicId}`)

      return
    }

    if (chapter.value.progress) {
      page.value = chapter.value.progress.page
    }

    isLoading.value = false
  })

  const retryDownloadLoading = ref(false)

  async function retryDownload(): Promise<void> {
    retryDownloadLoading.value = true
    const res = await api.retryDownload(chapterId.value)
    if (res.success) {
      await fetchChapter(chapterId.value)
    } else {
      console.error('api.retryDownload', res.error)
      if (res.error.status === 404 || res.error.status === 403) {
        toast.error(t('sources.asura.chapter.notFound'))
      } else {
        toast.error(t('error.unknown'))
      }
    }

    retryDownloadLoading.value = false
  }

  watch(chapterId, async (): Promise<void> => {
    const loadedChapter = loadedChapters.value.find(({ id }) => id === chapterId.value)
    if (loadedChapter || fetchChapterLoading.value) {
      return
    }

    isLoading.value = true

    await fetchChapter(chapterId.value)

    isLoading.value = false
  })

  watch(chapter, (newValue) => {
    if (!newValue) {
      return
    }

    if (startEnd.value) {
      page.value = newValue.pagesNb - 1

      startEnd.value = false
    } else {
      page.value = 0
    }
  })

  const isInFlight = ref(false)

  const { pause, resume } = useIntervalFn(async () => {
    if (isInFlight.value || isLoading.value || fetchChapterLoading.value) {
      return
    }

    isInFlight.value = true
    await fetchChapter(chapterId.value)
    isInFlight.value = false
  }, 2000, { immediate: false })

  watch(chapterId, () => {
    syncPolling()
  })

  watch(fetchChapterLoading, (value) => {
    if (value) {
      return
    }

    if (loadedChapters.value.at(-1)?.id === chapterId.value) {
      syncPolling()
    }
  })

  function syncPolling(): void {
    if (!chapter.value) {
      pause()

      return
    }

    if (chapter.value.download >= 0 && chapter.value.download < 100) {
      resume()
    } else {
      pause()
    }
  }

  function fetchNextChapter(): void {
    if (!chapter.value?.next) {
      return
    }

    fetchChapter(chapter.value.next.id)
  }

  function fetchPreviousChapter(): void {
    if (!chapter.value?.previous) {
      return
    }

    fetchChapter(chapter.value.previous.id)
  }

  const previousChapter = computed(() => {
    if (!chapter.value?.previous) {
      return
    }

    return loadedChapters.value.find(({ id }) => id === chapter.value?.previous?.id)
  })

  const nextChapter = computed(() => {
    if (!chapter.value?.next) {
      return
    }

    return loadedChapters.value.find(({ id }) => id === chapter.value?.next?.id)
  })

  watch(page, (value) => {
    if (!chapter.value || !readerSettings.value) {
      return
    }

    const actualPage = readerSettings.value.doublePage && value + 1 < chapter.value.pagesNb
      ? value + 1
      : value

    if (!chapter.value.progress || chapter.value.progress.page < actualPage) {
      saveProgress(chapter.value.id, actualPage)
    }
  })

  async function saveProgress(chapterId: string, page: number): Promise<void> {
    const res = await api.saveProgress({
      id: chapterId,
      page,
    })

    if (!res.success) {
      console.error('api.saveProgress', res.error)

      return
    }

    const index = loadedChapters.value.findIndex(({ id }) => id === chapterId)
    if (index === -1) {
      return
    }

    loadedChapters.value[index]!.progress = res.data
  }

  return {
    page,
    chapter,
    comic,
    readerSettings,

    isLoading: computed(() => isLoading.value || (!chapter.value && fetchChapterLoading.value)),

    fetchNextChapter,
    fetchPreviousChapter,

    retryDownload,
    retryDownloadLoading,

    nextChapter,
    previousChapter,

    startEnd,
  }
}
