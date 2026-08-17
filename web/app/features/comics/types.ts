// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ASURA_SOURCE_NAME } from '~/constants'

export type ComicStatus = 'ongoing' | 'completed' | 'hiatus' | 'cancelled'
export type ComicType = 'manga' | 'manhua' | 'manhwa' | 'mangatoon'
export type ComicSource = typeof ASURA_SOURCE_NAME

export interface LightComic {
  id: string
  title: string
  slug: string
  source: string
  status: ComicStatus
  chapterCount: number
  cover: string
}

export type SearchComicResponse = {
  total: number
  items: LightComic[]
}

export type SearchComicSort = 'title' | 'addedAt'
export type SearchComicOrder = 'asc' | 'desc'

export type SearchComicParams = {
  source?: ComicSource
  type?: ComicType
  status?: ComicStatus
  search?: string
  sort: SearchComicSort
  order: SearchComicOrder
  limit: number
  offset: number
}
