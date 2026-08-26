// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceSearchItem, SourceSort } from '../types'
import type { ComicSource, ComicStatus, ComicType } from '~/features/comics/types'

const DEFAULT_SORT: SourceSort = 'popular'
const DEFAULT_PAGE = 1

export interface SourceSearchStore {
  comics: Ref<SourceSearchItem[]>
  setComics: (comics: SourceSearchItem[]) => void
  clearComics: () => void

  accumulatedComics: Ref<SourceSearchItem[]>
  setAccumulatedComics: (comics: SourceSearchItem[]) => void
  clearAccumulatedComics: () => void

  setComicInternalId: (slug: string, internalId: string | undefined) => void

  loading: Ref<boolean>
  setLoading: (isLoading: boolean) => void
  clearLoading: () => void

  search: Ref<string | undefined>
  setSearch: (value: string) => void
  clearSearch: () => void

  sort: Ref<SourceSort>
  setSort: (value: SourceSort) => void
  clearSort: () => void

  status: Ref<ComicStatus | undefined>
  setStatus: (value: ComicStatus) => void
  clearStatus: () => void

  type: Ref<ComicType | undefined>
  setType: (value: ComicType) => void
  clearType: () => void

  page: Ref<number>
  setPage: (value: number) => void
  clearPage: () => void

  invalidate: () => void
}

type SearchStoreDefinition = ReturnType<typeof defineStore<string, SourceSearchStore>>

const storesMap = new Map<ComicSource, SearchStoreDefinition>()

function createSearchStoreDefinition(sourceId: ComicSource): SearchStoreDefinition {
  return defineStore(`sources-search-${sourceId}`, (): SourceSearchStore => {
    const search = ref<string>()
    const sort = ref<SourceSort>(DEFAULT_SORT)
    const status = ref<ComicStatus>()
    const type = ref<ComicType>()
    const page = ref<number>(DEFAULT_PAGE)
    const loading = ref<boolean>(false)

    const comics = ref<SourceSearchItem[]>([])
    const accumulatedComics = ref<SourceSearchItem[]>([])

    function setComics(value: SourceSearchItem[]): void {
      comics.value = [...value]
    }

    function clearComics(): void {
      comics.value = []
    }

    function setAccumulatedComics(value: SourceSearchItem[]): void {
      accumulatedComics.value = [...value]
    }

    function clearAccumulatedComics(): void {
      accumulatedComics.value = []
    }

    function setLoading(isLoading: boolean): void {
      loading.value = isLoading
    }

    function clearLoading(): void {
      loading.value = false
    }

    function setComicInternalId(slug: string, internalId: string | undefined): void {
      let i = comics.value.findIndex(c => c.slug === slug)
      if (i !== -1) {
        const { internalId: _, ...comic } = comics.value[i]!
        comics.value[i] = { ...comic, internalId }
      }

      i = accumulatedComics.value.findIndex(c => c.slug === slug)
      if (i !== -1) {
        const { internalId: _, ...comic } = accumulatedComics.value[i]!
        accumulatedComics.value[i] = { ...comic, internalId }
      }
    }

    function setSearch(value: string): void {
      search.value = value
    }

    function clearSearch(): void {
      search.value = undefined
    }

    function setSort(value: SourceSort): void {
      sort.value = value
    }

    function clearSort(): void {
      sort.value = DEFAULT_SORT
    }

    function setStatus(value: ComicStatus): void {
      status.value = value
    }

    function clearStatus(): void {
      status.value = undefined
    }

    function setType(value: ComicType): void {
      type.value = value
    }

    function clearType(): void {
      type.value = undefined
    }

    function setPage(value: number): void {
      page.value = value
    }

    function clearPage(): void {
      page.value = DEFAULT_PAGE
    }

    function invalidate(): void {
      clearSearch()
      clearSort()
      clearStatus()
      clearType()
      clearPage()
      clearLoading()
      clearComics()
      clearAccumulatedComics()
    }

    return {
      comics,
      setComics,
      clearComics,
      accumulatedComics,
      setAccumulatedComics,
      clearAccumulatedComics,
      setComicInternalId,
      loading,
      setLoading,
      clearLoading,
      search,
      setSearch,
      clearSearch,
      sort,
      setSort,
      clearSort,
      status,
      setStatus,
      clearStatus,
      type,
      setType,
      clearType,
      page,
      setPage,
      clearPage,
      invalidate,
    }
  })
}

export function useSourceSearchStore(sourceId: ComicSource): ReturnType<SearchStoreDefinition> {
  if (!storesMap.has(sourceId)) {
    storesMap.set(sourceId, createSearchStoreDefinition(sourceId))
  }

  return storesMap.get(sourceId)!()
}
