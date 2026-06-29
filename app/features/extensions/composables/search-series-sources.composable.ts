// SPDX-License-Identifier: AGPL-3.0-or-later

import type { MangaChapterSummaryDto } from '#shared/dto/catalogue/manga-chapter-summary.dto'
import type { SourceFilterDto } from '#shared/dto/catalogue/source-filters.dto'
import type { SourceSearchItemDto, SourceSearchQueryType, SourceSearchResultDto } from '#shared/dto/catalogue/source-search.dto'
import type { ChapterSummaryStatus } from '../store/chapter-summaries.store'
import type { FilterDraftValue } from './source-filters.composable'
import { endomyLanguage } from '~/utils/languages.util'
import { createExtensionsApi } from '../api/extensions.api'
import { useSourceFilters } from './source-filters.composable'

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
  sourcesItems: ComputedRef<{ value: string, title: string, supportsLatest: boolean }[]>
  selectedSourceId: Ref<string | undefined>
  chapterStatusOf: (id: string) => ChapterSummaryStatus | undefined
  chapterSummaryOf: (id: string) => MangaChapterSummaryDto | undefined
  definitions: Ref<SourceFilterDto[]>
  filtersDraft: Ref<Record<string, FilterDraftValue>>
  filtersActiveCount: ComputedRef<number>
  filtersLoading: Ref<boolean>
  applyFilters: () => void
  resetFilters: () => void
}

export function useSearchSeriesSourcesComposable(): UseSearchSeriesSourcesComposable {
  const api = createExtensionsApi()
  const store = useSingleExtensionStore()
  const toast = useToast()
  const { t } = useI18n()

  const summaries = useChapterSummariesStore()
  const { mobile } = useDisplay()

  const extension = computed(() => store.extension)
  const sourcesItems = computed(() => {
    if (!store.sources?.length) {
      return []
    }

    const ret: { value: string, title: string, supportsLatest: boolean }[] = []

    for (const source of store.sources) {
      if (!source.isEnabled) {
        continue
      }

      ret.push({
        value: source.id,
        title: source.lang === 'all' ? t('sources.all') : endomyLanguage(source.lang),
        supportsLatest: source.supportsLatest,
      })
    }

    return ret
  })

  const searchFilter = ref<string>()
  const page = ref(1)
  const debouncedSearch = useDebounce(searchFilter, 300)
  const selectedSourceId = ref<string | undefined>(sourcesItems.value[0]?.value)
  const searchType = ref<SourceSearchQueryType>('popular')

  const sourceFilters = useSourceFilters({
    pkgName: computed(() => extension.value?.pkgName),
    sourceId: selectedSourceId,
    enabled: computed(() => searchType.value === 'search'),
  })

  watch(sourcesItems, (newValue, oldValue) => {
    if (oldValue.length === 0 && newValue.length > 0) {
      selectedSourceId.value = newValue[0]?.value as string
    }
  }, { once: true })

  watch(selectedSourceId, (newValue) => {
    if (!newValue) {
      selectedSourceId.value = sourcesItems.value[0]?.value as string
    }
  })

  const queryKey = computed(() => [
    searchType.value,
    ...(extension.value ? [extension.value.pkgName] : []),
    ...(selectedSourceId.value ? [selectedSourceId.value] : []),
    ...(debouncedSearch.value ? [debouncedSearch.value] : []),
    ...(searchType.value === 'search' ? [JSON.stringify(sourceFilters.appliedChanges.value)] : []),
    page.value,
  ])

  const { data, isLoading } = useQuery({
    key: queryKey,
    query: async () => {
      if (!extension.value || !selectedSourceId.value) {
        return DEFAULT_SEARCH_RESULT
      }

      if (searchType.value === 'search' && !debouncedSearch.value && sourceFilters.appliedChanges.value.length === 0) {
        return DEFAULT_SEARCH_RESULT
      }

      const res = await api.searchSeriesBySource(extension.value.pkgName, selectedSourceId.value, {
        q: debouncedSearch.value,
        type: searchType.value,
        page: page.value,
        filters: searchType.value === 'search' ? sourceFilters.appliedChanges.value : undefined,
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
    searchFilter.value = undefined
  })

  watch(debouncedSearch, () => {
    page.value = 1
  })

  // On mobile we append each new page to the already-loaded results (infinite
  // scroll); on desktop the pagination footer drives `page`, so each page
  // replaces the previous one. Page 1 always resets the accumulator (filter,
  // source or search-type change resets `page` to 1 first).
  const accumulatedSeries = ref<SourceSearchItemDto[]>([])

  watch(data, (value) => {
    const items = value?.items ?? []

    accumulatedSeries.value = mobile.value && page.value > 1
      ? [...accumulatedSeries.value, ...items]
      : items
  }, { immediate: true })

  const series = computed(() => accumulatedSeries.value)

  watch(series, (items) => {
    if (!extension.value || !selectedSourceId.value) {
      return
    }

    summaries.sync(items.map(item => item.id), { pkgName: extension.value.pkgName, sourceId: selectedSourceId.value })
  })

  function applyFilters(): void {
    searchType.value = 'search'
    page.value = 1
    sourceFilters.apply()
  }

  function resetFilters(): void {
    sourceFilters.reset()
  }

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
    definitions: sourceFilters.definitions,
    filtersDraft: sourceFilters.draft,
    filtersActiveCount: sourceFilters.activeCount,
    filtersLoading: sourceFilters.filtersLoading,
    applyFilters,
    resetFilters,
  }
}
