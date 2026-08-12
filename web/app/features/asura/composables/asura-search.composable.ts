// SPDX-License-Identifier: AGPL-3.0-or-later

import type { AsuraSort, AsuraSortOrder, AsuraStatus, AsuraType } from '../types'

const PAGE_SIZE = 20

export interface AsuraSearchComposable {
  search: Ref<string | undefined>
  sort: Ref<AsuraSort>
  sortOrder: Ref<AsuraSortOrder>
  status: Ref<AsuraStatus | undefined>
  type: Ref<AsuraType | undefined>
  offset: Ref<number>
  maxPage: Ref<number>
  series: Ref<AsuraSearchItem[]>
  isLoading: Ref<boolean>
  resetFilters: () => void
}

export function useAsuraSearch(opts: { doSearch: boolean }): AsuraSearchComposable {
  const api = createAsuraApi()
  const store = useAsuraSearchStore()
  const { t } = useI18n()
  const toast = useToast()
  const { smAndDown } = useDisplay()

  const search = computed({
    get: () => store.search,
    set: (value?: string) => value ? store.setSearch(value) : store.clearSearch(),
  })

  const debouncedSearch = useDebounce(search, 200)

  const sort = computed({
    get: () => store.sort,
    set: (value: AsuraSort) => store.setSort(value),
  })

  const sortOrder = computed({
    get: () => store.sortOrder,
    set: (value: AsuraSortOrder) => store.setSortOrder(value),
  })

  const status = computed({
    get: () => store.status,
    set: (value?: AsuraStatus) => value ? store.setStatus(value) : store.clearStatus(),
  })

  const type = computed({
    get: () => store.type,
    set: (value?: AsuraType) => value ? store.setType(value) : store.clearType(),
  })

  const offset = computed({
    get: () => store.offset,
    set: (value: number) => store.setOffset(value),
  })

  function resetFilters(): void {
    store.invalidate()
  }

  const series = ref<AsuraSearchItem[]>([])
  const accumulatedSeries = ref<AsuraSearchItem[]>([])
  const maxPage = ref(0)

  const isLoading = ref(false)
  const debouncedSearchFn = useDebounceFn(async () => {
    const res = await api.search({
      ...(debouncedSearch.value && { search: debouncedSearch.value }),
      ...(sort.value && { sort: sort.value }),
      ...(sortOrder.value && { sortOrder: sortOrder.value }),
      ...(status.value && { status: status.value }),
      ...(type.value && { type: type.value }),
      offset: offset.value,
      limit: PAGE_SIZE,
    })

    if (!res.success) {
      toast.error(t('sources.asurascans.search.error'))

      accumulatedSeries.value = []
      maxPage.value = 0
      series.value = []

      return
    }

    if (smAndDown.value) {
      accumulatedSeries.value = [...accumulatedSeries.value, ...res.data.items]
    }

    maxPage.value = res.data.total / PAGE_SIZE
    series.value = res.data.items
  }, 200)

  watch([search, sort, sortOrder, status, type], () => {
    offset.value = 0
    accumulatedSeries.value = []
    series.value = []
  })

  watch([search, sort, sortOrder, status, type, offset], () => {
    if (opts.doSearch) {
      debouncedSearchFn()
    }
  }, { immediate: true })

  return {
    search: debouncedSearch,
    sort,
    sortOrder,
    status,
    type,
    offset,
    maxPage,
    series: computed(() => smAndDown.value ? accumulatedSeries.value : series.value),
    isLoading,
    resetFilters,
  }
}
