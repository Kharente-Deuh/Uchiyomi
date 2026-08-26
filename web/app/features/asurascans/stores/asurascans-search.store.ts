// SPDX-License-Identifier: AGPL-3.0-or-later

import type { AsuraScansSearchItem, AsuraScansSort } from '../types'
import type { ComicStatus, ComicType } from '~/features/comics/types'

const DEFAULT_SORT: AsuraScansSort = 'popular'
const DEFAULT_PAGE = 1

export interface AsuraScansSearchStore {
  comics: Ref<AsuraScansSearchItem[]>
  setComics: (comics: AsuraScansSearchItem[]) => void
  clearComics: () => void

  accumulatedComics: Ref<AsuraScansSearchItem[]>
  setAccumulatedComics: (comics: AsuraScansSearchItem[]) => void
  clearAccumulatedComics: () => void

  setComicInternalId: (slug: string, internalId: string | undefined) => void

  loading: Ref<boolean>
  setLoading: (isLoading: boolean) => void
  clearLoading: () => void

  search: Ref<string | undefined>
  setSearch: (value: string) => void
  clearSearch: () => void

  sort: Ref<AsuraScansSort>
  setSort: (value: AsuraScansSort) => void
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

export const useAsuraScansSearchStore = defineStore('asurascans-search', (): AsuraScansSearchStore => {
  const search = ref<string>()
  const sort = ref<AsuraScansSort>(DEFAULT_SORT)
  const status = ref<ComicStatus>()
  const type = ref<ComicType>()
  const page = ref<number>(DEFAULT_PAGE)
  const loading = ref<boolean>(false)

  const comics = ref<AsuraScansSearchItem[]>([])
  const accumulatedComics = ref<AsuraScansSearchItem[]>([])

  function setComics(value: AsuraScansSearchItem[]): void {
    comics.value = [...value]
  }

  function clearComics(): void {
    comics.value = []
  }

  function setAccumulatedComics(value: AsuraScansSearchItem[]): void {
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
    if (i === -1) {
      return
    }

    const { internalId: _, ...comic } = comics.value[i]!

    comics.value[i] = { ...comic, internalId }

    i = accumulatedComics.value.findIndex(c => c.slug === slug)
    if (i === -1) {
      return
    }

    accumulatedComics.value[i] = { ...comic, internalId }
  }

  function setSearch(value: string): void {
    search.value = value
  }

  function clearSearch(): void {
    search.value = undefined
  }

  function setSort(value: AsuraScansSort): void {
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
