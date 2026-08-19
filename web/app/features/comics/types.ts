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
export type SearchComicParams = {
  source?: ComicSource
  type?: ComicType
  status?: ComicStatus
  search?: string
  sort: SearchComicSort
  order: 'asc' | 'desc'
  limit: number
  offset: number
}

export interface Comic {
  id: string
  artist: string
  type: ComicType
  description: string
  source: ComicSource
  author: string
  status: ComicStatus
  slug: string
  title: string
  cover: string
  genres: string[]
  altTitles: string[]
  chapterCount: number
}
