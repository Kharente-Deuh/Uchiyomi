import type { ComicSource, ComicStatus, ComicType } from '../comics/types'

export interface FeedItem {
  id: string
  title: string
  slug: string
  cover: string
  source: ComicSource
  status: ComicStatus
  type: ComicType
  latestChapters: FeedChapter[]
}

export interface FeedChapter {
  id: string
  title?: string
  publishedAt: Date
  earlyAccessUntil: Date
  number: number
  download: number
}

export interface FeedResponse {
  items: FeedItem[]
  total: number
}

export interface FeedParams {
  offset: number
  limit: number
  type?: ComicType
  source?: ComicSource
}
