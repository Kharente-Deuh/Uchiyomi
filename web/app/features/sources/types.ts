// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ComicSource, ComicStatus, ComicType } from '../comics/types'

export type SourceSort = 'popular' | 'latest' | 'rating' | 'title' | 'newest'

export interface SourceSearchParams {
  search?: string
  sort?: SourceSort
  status?: ComicStatus
  type?: ComicType
  artist?: string
  page: number
  minChapters?: number
}

export interface SourceSearchItem {
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
  latestChapters?: SourceComicChapter[]
  chapterCount: number
  rating: number
  releaseYear?: number
}

export interface SourceSearchResponse {
  items: SourceSearchItem[]
  hasNextPage: boolean
}

export interface SourceComicChapter {
  id: string
  title: string
  number: number
  publishedAt: Date
  earlyAccessUntil?: Date
  internalId?: string
  download?: number
}

export interface SourceComicInfos {
  slug: string
  title: string
  description: string
  cover: string
  status: ComicStatus
  type: ComicType
  author: string
  artist: string
  sourceUrl: string
  publicUrl: string
  altTitles: string[]
  genres: string[]
  chapterCount: number
  rating: number
  releaseYear?: number
  lastChapterAt?: Date
  updatedAt: Date
  createdAt: Date
  internalId?: string
}

export interface SourceConfig {
  id: ComicSource
  nameKey: string
  url: string
  image: string
  color: string
  allowedSorts: SourceSort[]
  supportsMinChapters?: boolean
}
