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

  return {
    items,
    isLoading,
    page,
    maxPage,
    source,
    type,
  }
}
