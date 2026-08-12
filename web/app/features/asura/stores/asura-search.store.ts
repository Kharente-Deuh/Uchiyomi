// SPDX-License-Identifier: AGPL-3.0-or-later

import type { AsuraSort, AsuraSortOrder, AsuraStatus, AsuraType } from '../types'

const DEFAULT_SORT: AsuraSort = 'popular'
const DEFAULT_SORT_ORDER: AsuraSortOrder = 'asc'

export interface AsuraSearchStore {
  search: Ref<string | undefined>
  setSearch: (value: string) => void
  clearSearch: () => void

  sort: Ref<AsuraSort>
  setSort: (value: AsuraSort) => void
  clearSort: () => void

  sortOrder: Ref<AsuraSortOrder>
  setSortOrder: (value: AsuraSortOrder) => void
  clearSortOrder: () => void

  status: Ref<AsuraStatus | undefined>
  setStatus: (value: AsuraStatus) => void
  clearStatus: () => void

  type: Ref<AsuraType | undefined>
  setType: (value: AsuraType) => void
  clearType: () => void

  offset: Ref<number>
  setOffset: (value: number) => void
  clearOffset: () => void

  invalidate: () => void
}

export const useAsuraSearchStore = defineStore('asura-search', (): AsuraSearchStore => {
  const search = ref<string>()
  const sort = ref<AsuraSort>(DEFAULT_SORT)
  const sortOrder = ref<AsuraSortOrder>(DEFAULT_SORT_ORDER)
  const status = ref<AsuraStatus>()
  const type = ref<AsuraType>()
  const offset = ref<number>(1)

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

  function setSortOrder(value: AsuraSortOrder): void {
    sortOrder.value = value
  }

  function clearSortOrder(): void {
    sortOrder.value = DEFAULT_SORT_ORDER
  }

  function setStatus(value: AsuraStatus): void {
    status.value = value
  }

  function clearStatus(): void {
    status.value = undefined
  }

  function setType(value: AsuraType): void {
    type.value = value
  }

  function clearType(): void {
    type.value = undefined
  }

  function setOffset(value: number): void {
    offset.value = value
  }

  function clearOffset(): void {
    offset.value = 1
  }

  function invalidate(): void {
    clearSearch()
    clearSort()
    clearSortOrder()
    clearStatus()
    clearType()
    clearOffset()
  }

  return {
    search,
    setSearch,
    clearSearch,

    sort,
    setSort,
    clearSort,

    sortOrder,
    setSortOrder,
    clearSortOrder,

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
