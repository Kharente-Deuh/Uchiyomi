import type { AsuraComicChapter } from '../types'
import { useAsuraChaptersStore } from '../stores/asura-chapters.store'

export interface AsuraChaptersComposable {
  sort: Ref<'asc' | 'desc'>
  chapters: Ref<AsuraComicChapter[]>
  loading: Ref<boolean>
  fetchChapters: () => Promise<void>
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
  async function fetchChapters(): Promise<void> {
    loading.value = true

    const res = await api.getSeriesChapters(route.params.slug)

    if (res.success) {
      store.setChapters(res.data)
    } else {
      console.error('api.getSeriesChapters', res.error)

      if (res.error.status === 404) {
        toast.error(t('sources.asura.comic.chapters.error.fetch'))
      } else {
        toast.error(t('error.unknown'))
      }
    }

    loading.value = false
  }

  onBeforeRouteLeave(() => {
    store.invalidate()
  })

  return {
    sort,
    chapters,
    loading,
    fetchChapters,
  }
}
