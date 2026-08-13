// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ComicStatus, ComicType } from '../comics/types'

export type AsuraSort = 'popular' | 'latest' | 'rating' | 'title' | 'newest'

export interface AsuraSearchParams {
  search?: string
  sort?: AsuraSort
  status?: ComicStatus
  type?: ComicType
  artist?: string
  offset: number
  limit: number
  minChapters?: number
}

export interface AsuraSearchResponse {
  items: AsuraSearchItem[]
  total: number
}

export interface AsuraSearchItem {
  internalId?: string
  lastChapterAt: Date
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
  latestChapters: AsuraComicChapter[]
  chapterCount: number
  rating: number
  releaseYear: number
}

export interface AsuraComicChapter {
  earlyAccessUntil: Date
  publishedAt: Date
  id: string
  title: string
  number: number
}

export interface AsuraComicInfos {
  lastChapterAt: Date
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
}
