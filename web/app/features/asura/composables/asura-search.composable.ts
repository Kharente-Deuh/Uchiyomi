// SPDX-License-Identifier: AGPL-3.0-or-later

import type { AsuraSort } from '../types'
import type { ComicStatus, ComicType } from '~/features/comics/types'
import { createComicsApi } from '~/features/comics/composables/comics.api'

const PAGE_SIZE = 20
const ASURA_SOURCE_NAME = 'asurascans'

export interface AsuraSearchComposable {
  search: Ref<string | undefined>
  sort: Ref<AsuraSort>
  status: Ref<ComicStatus | undefined>
  type: Ref<ComicType | undefined>
  page: Ref<number>
  maxPage: Ref<number>
  series: Ref<AsuraSearchItem[]>
  isLoading: Ref<boolean>
  resetFilters: () => void

  removeComicFromLibrary: (comic: AsuraSearchItem) => Promise<void>
  addComicInLibrary: (comic: AsuraSearchItem) => Promise<void>
  addComicInLibraryLoading: Ref<Record<string, boolean>>
}

export function useAsuraSearch(opts: { doSearch: boolean }): AsuraSearchComposable {
  const api = createAsuraApi()
  const store = useAsuraSearchStore()
  const comicsApi = createComicsApi()
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
    set: (value: AsuraSearchItem[]) => store.setComics(value),
  })

  const accumulatedComics = computed({
    get: () => store.accumulatedComics,
    set: (value: AsuraSearchItem[]) => store.setAccumulatedComics(value),
  })

  function resetFilters(): void {
    store.invalidate()
  }

  const maxPage = ref(0)
  const addComicInLibraryLoading = ref<Record<string, boolean>>({})

  const debouncedSearchFn = useDebounceFn(async () => {
    isLoading.value = true

    const res = await api.search({
      ...(debouncedSearch.value && { search: debouncedSearch.value }),
      ...(sort.value && { sort: sort.value }),
      ...(status.value && { status: status.value }),
      ...(type.value && { type: type.value }),
      offset: offset.value,
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

  watch([search, sort, status, type], () => {
    offset.value = 0
    accumulatedComics.value = []
    comics.value = []
  })

  watch([search, sort, status, type, offset], () => {
    if (opts.doSearch) {
      debouncedSearchFn()
    }
  }, { immediate: true })

  async function removeComicFromLibrary(comic: AsuraSearchItem): Promise<void> {
    if (!comic.internalId) {
      return
    }

    const res = await comicsApi.deleteById(comic.internalId)
    if (!res.success) {
      console.error('comicsApi.deleteById failed', res.error)
      toast.error(t('error.unknown'))

      return
    }

    store.setComicInternalId(comic.slug, undefined)
  }

  async function addComicInLibrary(comic: AsuraSearchItem): Promise<void> {
    if (Object.hasOwn(addComicInLibraryLoading.value, comic.slug) || comic.internalId) {
      return
    }

    addComicInLibraryLoading.value[comic.slug] = true

    const res = await comicsApi.create({ source: ASURA_SOURCE_NAME, slug: comic.slug })
    if (!res.success) {
      if (res.error.status === 409) {
        toast.error(t('comics.create.error.alreadyExists'))

        delete addComicInLibraryLoading.value[comic.slug]

        return
      }

      console.error('comicsApi.create failed', res.error)
      toast.error(t('error.unknown'))

      delete addComicInLibraryLoading.value[comic.slug]

      return
    }

    store.setComicInternalId(comic.slug, res.data.id)

    delete addComicInLibraryLoading.value[comic.slug]
  }

  return {
    search: debouncedSearch,
    sort,
    status,
    type,
    page: offset,
    maxPage,
    series: computed(() => smAndDown.value ? accumulatedComics.value : comics.value),
    isLoading,
    resetFilters,

    addComicInLibrary,
    addComicInLibraryLoading,
    removeComicFromLibrary,
  }
}
