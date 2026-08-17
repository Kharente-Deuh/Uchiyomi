import type { ComicSource, ComicStatus, ComicType, LightComic, SearchComicSort } from '~/features/comics/types'

const DEFAULT_SORT: SearchComicSort = 'title'
const DEFAULT_OFFSET = 1

export interface LibraryStore {
  comics: Ref<LightComic[]>
  setComics: (comics: LightComic[]) => void
  clearComics: () => void

  accumulatedComics: Ref<LightComic[]>
  setAccumulatedComics: (comics: LightComic[]) => void
  clearAccumulatedComics: () => void

  loading: Ref<boolean>
  setLoading: (isLoading: boolean) => void
  clearLoading: () => void

  search: Ref<string | undefined>
  setSearch: (value: string) => void
  clearSearch: () => void

  sort: Ref<SearchComicSort>
  setSort: (value: SearchComicSort) => void

  status: Ref<ComicStatus | undefined>
  setStatus: (value: ComicStatus) => void
  clearStatus: () => void

  type: Ref<ComicType | undefined>
  setType: (value: ComicType) => void
  clearType: () => void

  source: Ref<ComicSource | undefined>
  setSource: (source: ComicSource) => void
  clearSource: () => void

  offset: Ref<number>
  setOffset: (value: number) => void
  clearOffset: () => void

  invalidate: () => void
}

export const useLibraryStore = defineStore('library', (): LibraryStore => {
  const comics = ref<LightComic[]>([])
  const accumulatedComics = ref<LightComic[]>([])

  const loading = ref<boolean>(false)
  const search = ref<string>()
  const sort = ref<SearchComicSort>(DEFAULT_SORT)
  const status = ref<ComicStatus>()
  const type = ref<ComicType>()
  const offset = ref<number>(DEFAULT_OFFSET)
  const source = ref<ComicSource>()

  function setComics(value: LightComic[]): void {
    comics.value = [...value]
  }

  function clearComics(): void {
    comics.value = []
  }

  function setAccumulatedComics(value: LightComic[]): void {
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

  function setSearch(value: string): void {
    search.value = value
  }

  function clearSearch(): void {
    search.value = undefined
  }

  function setSort(value: SearchComicSort): void {
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

  function setSource(value: ComicSource): void {
    source.value = value
  }

  function clearSource(): void {
    source.value = undefined
  }

  function invalidate(): void {
    clearComics()
    clearAccumulatedComics()
    clearLoading()
    clearSearch()
    clearSort()
    clearStatus()
    clearType()
    clearOffset()
    clearSource()
  }

  return {
    comics,
    setComics,
    clearComics,

    accumulatedComics,
    setAccumulatedComics,
    clearAccumulatedComics,

    loading,
    setLoading,
    clearLoading,

    search,
    setSearch,
    clearSearch,

    sort,
    setSort,

    status,
    setStatus,
    clearStatus,

    type,
    setType,
    clearType,

    source,
    setSource,
    clearSource,

    offset,
    setOffset,
    clearOffset,

    invalidate,
  }
})
