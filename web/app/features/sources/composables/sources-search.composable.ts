// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceComicInfos, SourceSearchItem, SourceSort } from '../types'
import type { ComicSource, ComicStatus, ComicType } from '~/features/comics/types'
import { createComicsApi } from '~/features/comics/composables/comics.api'
import { getSourceConfig } from '../config/sources.config'
import { useSourceSearchStore } from '../stores/sources-search.store'
import { createSourceApi } from './sources.api'

export interface SourceSearchComposable {
  search: Ref<string | undefined>
  sort: Ref<SourceSort>
  status: Ref<ComicStatus | undefined>
  type: Ref<ComicType | undefined>
  page: Ref<number>
  hasNextPage: Ref<boolean>
  series: Ref<SourceSearchItem[]>
  isLoading: Ref<boolean>
  resetFilters: () => void

  removeComicFromLibrary: (comic: SourceSearchItem) => Promise<boolean>
  addComicInLibrary: (comic: SourceSearchItem | SourceComicInfos) => Promise<string | undefined>
  addComicInLibraryLoading: Ref<Record<string, boolean>>
  infosLoading: ComputedRef<Record<string, boolean>>
}

export function useSourceSearch(sourceId: ComicSource, opts: { doSearch: boolean }): SourceSearchComposable {
  const api = createSourceApi(sourceId)
  const store = useSourceSearchStore(sourceId)
  const comicsApi = createComicsApi()
  const { t } = useI18n()
  const toast = useToast()
  const { smAndDown } = useDisplay()

  const search = computed({
    get: () => store.search,
    set: (value?: string) => value ? store.setSearch(value) : store.clearSearch(),
  })

  const sort = computed({
    get: () => store.sort,
    set: (value: SourceSort) => store.setSort(value),
  })

  const status = computed({
    get: () => store.status,
    set: (value?: ComicStatus) => value ? store.setStatus(value) : store.clearStatus(),
  })

  const type = computed({
    get: () => store.type,
    set: (value?: ComicType) => value ? store.setType(value) : store.clearType(),
  })

  const page = computed({
    get: () => store.page,
    set: (value: number) => store.setPage(value),
  })

  const isLoading = computed({
    get: () => store.loading,
    set: (isLoadingValue: boolean) => store.setLoading(isLoadingValue),
  })

  const comics = computed({
    get: () => store.comics,
    set: (value: SourceSearchItem[]) => store.setComics(value),
  })

  const accumulatedComics = computed({
    get: () => store.accumulatedComics,
    set: (value: SourceSearchItem[]) => store.setAccumulatedComics(value),
  })

  function resetFilters(): void {
    store.invalidate()
  }

  const hasNextPage = ref(false)
  const addComicInLibraryLoading = ref<Record<string, boolean>>({})
  let enrichGeneration = 0

  const infosLoading = computed(() => store.infosLoading)

  async function enrichSearchItems(items: SourceSearchItem[], generation: number): Promise<void> {
    await Promise.all(items.map(async (item) => {
      store.setInfosLoading(item.slug, true)
      const infos = await api.getInfosBySlug(item.slug)
      if (generation !== enrichGeneration) {
        return
      }

      store.setInfosLoading(item.slug, false)
      if (!infos.success) {
        console.error('api.getInfosBySlug', infos.error)

        return
      }

      store.patchComic(item.slug, {
        status: infos.data.status,
        type: infos.data.type,
        chapterCount: infos.data.chapterCount,
      })
    }))
  }

  const debouncedSearchFn = useDebounceFn(async () => {
    isLoading.value = true

    const res = await api.search({
      ...(search.value && { search: search.value }),
      ...(sort.value && { sort: sort.value }),
      ...(status.value && { status: status.value }),
      ...(type.value && { type: type.value }),
      page: page.value,
    })

    isLoading.value = false

    if (!res.success) {
      toast.error(t('sources.search.error'))

      accumulatedComics.value = []
      hasNextPage.value = false
      comics.value = []

      return
    }

    if (smAndDown.value) {
      accumulatedComics.value = [...accumulatedComics.value, ...res.data.items]
    }

    hasNextPage.value = res.data.hasNextPage
    comics.value = res.data.items

    const isNewResultSet = page.value === 1 || !smAndDown.value
    if (isNewResultSet) {
      enrichGeneration += 1
      store.clearInfosLoading()
    }

    if (getSourceConfig(sourceId)?.enrichSearchFromSeries) {
      void enrichSearchItems(res.data.items, enrichGeneration)
    }
  }, 200)

  watch([search, sort, status, type], () => {
    page.value = 1
    accumulatedComics.value = []
    comics.value = []
  })

  watch([search, sort, status, type, page], () => {
    if (opts.doSearch) {
      debouncedSearchFn()
    }
  }, { immediate: true })

  async function removeComicFromLibrary(comic: SourceSearchItem): Promise<boolean> {
    if (!comic.internalId) {
      return false
    }

    const res = await comicsApi.deleteById(comic.internalId)
    if (!res.success) {
      console.error('comicsApi.deleteById failed', res.error)
      toast.error(t('error.unknown'))

      return false
    }

    store.setComicInternalId(comic.slug, undefined)

    return true
  }

  async function addComicInLibrary(comic: SourceSearchItem | SourceComicInfos): Promise<string | undefined> {
    if (Object.hasOwn(addComicInLibraryLoading.value, comic.slug) || comic.internalId) {
      return
    }

    addComicInLibraryLoading.value[comic.slug] = true

    const res = await comicsApi.create({ source: sourceId, slug: comic.slug })
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

    return res.data.id
  }

  return {
    search,
    sort,
    status,
    type,
    page,
    hasNextPage,
    series: computed(() => smAndDown.value ? accumulatedComics.value : comics.value),
    isLoading,
    resetFilters,

    addComicInLibrary,
    addComicInLibraryLoading,
    removeComicFromLibrary,
    infosLoading,
  }
}
