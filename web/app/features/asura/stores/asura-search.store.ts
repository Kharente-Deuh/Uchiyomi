// SPDX-License-Identifier: AGPL-3.0-or-later

import type { AsuraSearchItem, AsuraSort } from '../types'
import type { ComicStatus, ComicType } from '~/features/comics/types'

const DEFAULT_SORT: AsuraSort = 'popular'
const DEFAULT_OFFSET = 1

export interface AsuraSearchStore {
  comics: Ref<AsuraSearchItem[]>
  setComics: (comics: AsuraSearchItem[]) => void
  clearComics: () => void

  accumulatedComics: Ref<AsuraSearchItem[]>
  setAccumulatedComics: (comics: AsuraSearchItem[]) => void
  clearAccumulatedComics: () => void

  setComicInternalId: (slug: string, internalId: string | undefined) => void

  loading: Ref<boolean>
  setLoading: (isLoading: boolean) => void
  clearLoading: () => void

  search: Ref<string | undefined>
  setSearch: (value: string) => void
  clearSearch: () => void

  sort: Ref<AsuraSort>
  setSort: (value: AsuraSort) => void
  clearSort: () => void

  status: Ref<ComicStatus | undefined>
  setStatus: (value: ComicStatus) => void
  clearStatus: () => void

  type: Ref<ComicType | undefined>
  setType: (value: ComicType) => void
  clearType: () => void

  offset: Ref<number>
  setOffset: (value: number) => void
  clearOffset: () => void

  invalidate: () => void
}

export const useAsuraSearchStore = defineStore('asura-search', (): AsuraSearchStore => {
  const search = ref<string>()
  const sort = ref<AsuraSort>(DEFAULT_SORT)
  const status = ref<ComicStatus>()
  const type = ref<ComicType>()
  const offset = ref<number>(DEFAULT_OFFSET)
  const loading = ref<boolean>(false)

  const comics = ref<AsuraSearchItem[]>([])
  const accumulatedComics = ref<AsuraSearchItem[]>([])

  function setComics(value: AsuraSearchItem[]): void {
    comics.value = [...value]
  }

  function clearComics(): void {
    comics.value = []
  }

  function setAccumulatedComics(value: AsuraSearchItem[]): void {
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

  function setSort(value: AsuraSort): void {
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

  function setOffset(value: number): void {
    offset.value = value
  }

  function clearOffset(): void {
    offset.value = DEFAULT_OFFSET
  }

  function invalidate(): void {
    clearSearch()
    clearSort()
    clearStatus()
    clearType()
    clearOffset()
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

    offset,
    setOffset,
    clearOffset,

    invalidate,
  }
})
