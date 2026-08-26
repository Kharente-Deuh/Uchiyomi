// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ComicStatus, ComicType } from '../comics/types'

export type KingOfShojoSort = 'popular' | 'latest' | 'title' | 'newest'
export interface KingOfShojoSearchParams {
  search?: string
  sort?: KingOfShojoSort
  status?: ComicStatus
  type?: ComicType
  artist?: string
  page: number
}

export interface KingOfShojoSearchResponse {
  items: KingOfShojoSearchItem[]
  hasNextPage: boolean
}

export interface KingOfShojoSearchItem {
  lastChapterAt?: Date
  updatedAt: Date
  createdAt: Date
  internalId?: string
  description: string
  slug: string
  status: ComicStatus
  type: ComicType
  author: string
  artist: string
  sourceUrl: string
  cover: string
  title: string
  publicUrl: string
  genres: string[]
  altTitles: string[]
  chapterCount: number
  rating: number
  releaseYear: number
}

export interface KingOfShojoComicInfos {
  lastChapterAt?: Date
  updatedAt: Date
  createdAt: Date
  internalId?: string
  author: string
  publicUrl: string
  status: ComicStatus
  type: ComicType
  title: string
  artist: string
  sourceUrl: string
  cover: string
  slug: string
  description: string
  genres: string[]
  altTitles: string[]
  chapterCount: number
  rating: number
}

export interface KingOfShojoComicChapter {
  earlyAccessUntil?: Date
  publishedAt: Date
  internalId?: string
  download?: number
  id: string
  title: string
  number: number
}
