// SPDX-License-Identifier: AGPL-3.0-or-later

import type { MangaChapterSummaryDto } from '#shared/dto/catalogue/manga-chapter-summary.dto'
import type { SourceSearchItemDto, SourceSearchQueryType, SourceSearchResultDto } from '#shared/dto/catalogue/source-search.dto'
import type { ChapterSummaryStatus } from '../store/chapter-summaries.store'
import { endomyLanguage } from '~/utils/languages.util'
import { createExtensionsApi } from '../api/extensions.api'

const DEFAULT_SEARCH_RESULT: SourceSearchResultDto = {
  hasNextPage: false,
  items: [],
}

interface UseSearchSeriesSourcesComposable {
  series: ComputedRef<SourceSearchItemDto[]>
  hasNextPage: ComputedRef<boolean>
  searchLoading: Ref<boolean>
  searchType: Ref<SourceSearchQueryType>
  searchFilter: Ref<string | undefined>
  page: Ref<number>
  sourcesItems: ComputedRef<{ value: string, title: string }[]>
  selectedSourceId: Ref<string | undefined>
  chapterStatusOf: (id: string) => ChapterSummaryStatus | undefined
  chapterSummaryOf: (id: string) => MangaChapterSummaryDto | undefined
}

export function useSearchSeriesSourcesComposable(): UseSearchSeriesSourcesComposable {
  const api = createExtensionsApi()
  const store = useSingleExtensionStore()
  const toast = useToast()
  const { t } = useI18n()

  const summaries = useChapterSummariesStore()

  const extension = computed(() => store.extension)
  const sourcesItems = computed(() => {
    if (!store.sources?.length) {
      return []
    }

    const ret: { value: string, title: string }[] = []

    for (const source of store.sources) {
      if (!source.isEnabled) {
        continue
      }

      ret.push({
        value: source.id,
        title: source.lang === 'all' ? t('sources.all') : endomyLanguage(source.lang),
      })
    }

    return ret
  })

  const searchFilter = ref<string>()
  const page = ref(1)
  const debouncedSearch = useDebounce(searchFilter, 300)
  const selectedSourceId = ref<string | undefined>(sourcesItems.value[0]?.value)
  const searchType = ref<SourceSearchQueryType>('popular')

  watch(sourcesItems, (newValue, oldValue) => {
    if (oldValue.length === 0 && newValue.length > 0) {
      selectedSourceId.value = newValue[0]?.value as string
    }
  }, { once: true })

  const queryKey = computed(() => [
    ...(extension.value ? [extension.value.pkgName] : []),
    ...(selectedSourceId.value ? [selectedSourceId.value] : []),
    ...(debouncedSearch.value ? [debouncedSearch.value] : []),
    page.value,
  ])

  const { data, isLoading } = useQuery({
    key: queryKey,
    query: async () => {
      if (!extension.value || !selectedSourceId.value) {
        return DEFAULT_SEARCH_RESULT
      }

      if (searchType.value === 'search' && !debouncedSearch.value) {
        return DEFAULT_SEARCH_RESULT
      }

      const res = await api.searchSeriesBySource(extension.value.pkgName, selectedSourceId.value, {
        q: debouncedSearch.value,
        type: searchType.value,
        page: page.value,
      })

      if (!res.success) {
        toast.error(t('sources.series.search.errors.loadFailed'))

        return DEFAULT_SEARCH_RESULT
      }

      return res.data
    },
    enabled: computed(() => !!extension.value && !!selectedSourceId.value),
  })

  watch([searchType, selectedSourceId], () => {
    page.value = 1
  })

  watch(searchType, (_newValue, oldValue) => {
    if (oldValue === 'search') {
      searchFilter.value = undefined
    }
  })

  const series = computed(() => data.value?.items ?? [])

  watch(series, (items) => {
    if (!extension.value || !selectedSourceId.value) {
      return
    }

    summaries.sync(items.map(item => item.id), { pkgName: extension.value.pkgName, sourceId: selectedSourceId.value })
  })

  return {
    series,
    hasNextPage: computed(() => data.value?.hasNextPage ?? false),
    searchLoading: isLoading,
    searchType,
    searchFilter,
    sourcesItems,
    page,
    selectedSourceId,
    chapterStatusOf: (id: string) => summaries.statusOf(id),
    chapterSummaryOf: (id: string) => summaries.summaryOf(id),
  }
}
