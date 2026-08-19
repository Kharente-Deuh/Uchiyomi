import type { ComicType } from '../comics/types'

export type ReadingMode = 'paged-rtl' | 'paged-ltr' | 'webtoon'
export type PageScale = 'fit-width' | 'fit-height' | 'fit-screen'

export interface ReaderSettings {
  type: ComicType
  readingMode: ReadingMode
  pageScale: PageScale
  doublePage: boolean
}

export interface GetReaderSettingsResponse {
  items: ReaderSettings[]
}
