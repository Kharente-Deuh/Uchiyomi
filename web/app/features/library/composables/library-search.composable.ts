import type { ComicSource, ComicStatus, ComicType, LightComic, SearchComicSort } from '~/features/comics/types'

const PAGE_SIZE = 20

export interface LibrarySearchComposable {
  search: Ref<string | undefined>
  sort: Ref<SearchComicSort>
  source: Ref<ComicSource | undefined>
  status: Ref<ComicStatus | undefined>
  type: Ref<ComicType | undefined>
  page: Ref<number>
  maxPage: Ref<number>
  comics: Ref<LightComic[]>
  isLoading: Ref<boolean>
  resetFilters: () => void
}

export interface LibrarySearchOptions {
  doSearch: boolean
}

export function useLibrarySearch({ doSearch }: LibrarySearchOptions): LibrarySearchComposable {
  const store = useLibraryStore()
  const api = createComicsApi()
  const { t } = useI18n()
  const toast = useToast()
  const { smAndDown } = useDisplay()

  const search = computed({
    get: () => store.search,
    set: (value?: string) => value ? store.setSearch(value) : store.clearSearch(),
  })

  const sort = computed({
    get: () => store.sort,
    set: (value: SearchComicSort) => store.setSort(value),
  })

  const source = computed({
    get: () => store.source,
    set: (value?: ComicSource) => value ? store.setSource(value) : store.clearSource(),
  })

  const status = computed({
    get: () => store.status,
    set: (value?: ComicStatus) => value ? store.setStatus(value) : store.clearStatus(),
  })

  const type = computed({
    get: () => store.type,
    set: (value?: ComicType) => value ? store.setType(value) : store.clearType(),
  })

  const offset = computed({
    get: () => store.offset,
    set: (value: number) => store.setOffset(value),
  })

  const isLoading = computed({
    get: () => store.loading,
    set: (isLoadingValue: boolean) => store.setLoading(isLoadingValue),
  })

  const comics = computed({
    get: () => store.comics,
    set: (value: LightComic[]) => store.setComics(value),
  })

  const accumulatedComics = computed({
    get: () => store.accumulatedComics,
    set: (value: LightComic[]) => store.setAccumulatedComics(value),
  })

  function resetFilters(): void {
    store.invalidate()
  }

  const maxPage = ref(0)

  const debouncedSearchFn = useDebounceFn(async () => {
    isLoading.value = true

    const res = await api.search({
      ...(search.value && { search: search.value }),
      ...(status.value && { status: status.value }),
      ...(type.value && { type: type.value }),
      sort: sort.value,
      order: sort.value === 'latest' ? 'desc' : 'asc',
      offset: (offset.value - 1) * PAGE_SIZE,
      limit: PAGE_SIZE,
    })

    isLoading.value = false

    if (!res.success) {
      toast.error(t('sources.asurascans.search.error'))

      accumulatedComics.value = []
      maxPage.value = 0
      comics.value = []

      return
    }

    if (smAndDown.value) {
      accumulatedComics.value = [...accumulatedComics.value, ...res.data.items]
    }

    maxPage.value = Math.ceil(res.data.total / PAGE_SIZE)
    comics.value = res.data.items
  }, 200)

  watch([search, sort, status, type, source], () => {
    store.clearOffset()
  })

  watch([search, sort, status, type, offset, source], () => {
    if (doSearch) {
      debouncedSearchFn()
    }
  }, { immediate: true })

  return {
    search,
    sort,
    status,
    type,
    source,
    page: offset,
    maxPage,
    comics: computed(() => smAndDown.value ? accumulatedComics.value : comics.value),
    isLoading,
    resetFilters,
  }
}
