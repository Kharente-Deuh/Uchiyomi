// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ComicStatus, ComicType } from '../comics/types'

export type AsuraScansSort = 'popular' | 'latest' | 'rating' | 'title' | 'newest'

export interface AsuraScansSearchParams {
  search?: string
  sort?: AsuraScansSort
  status?: ComicStatus
  type?: ComicType
  artist?: string
  page: number
  minChapters?: number
}

export interface AsuraScansSearchResponse {
  items: AsuraScansSearchItem[]
  hasNextPage: boolean
}

export interface AsuraScansSearchItem {
  internalId?: string
  lastChapterAt?: Date
  updatedAt: Date
  createdAt: Date
  publicUrl: string
  sourceUrl: string
  cover: string
  status: ComicStatus
  type: ComicType
  author: string
  artist: string
  description: string
  slug: string
  title: string
  altTitles: string[]
  genres: string[]
  latestChapters: AsuraScansComicChapter[]
  chapterCount: number
  rating: number
  releaseYear: number
}

export interface AsuraScansComicChapter {
  earlyAccessUntil?: Date
  publishedAt: Date
  id: string
  title: string
  number: number
  internalId?: string
  download?: number
}

export interface AsuraScansComicInfos {
  lastChapterAt?: Date
  updatedAt: Date
  createdAt: Date
  description: string
  title: string
  cover: string
  status: ComicStatus
  type: ComicType
  author: string
  artist: string
  sourceUrl: string
  publicUrl: string
  slug: string
  altTitles: string[]
  genres: string[]
  chapterCount: number
  rating: number
  internalId?: string
}
