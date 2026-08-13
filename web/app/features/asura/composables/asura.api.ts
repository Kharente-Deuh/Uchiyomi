// SPDX-License-Identifier: AGPL-3.0-or-later

import type { AsuraSort } from '../types'
import type { ComicStatus, ComicType } from '~/features/comics/types'
import type { ApiResponse } from '~/utils/api'
import { ApiError, initApi } from '~/utils/api'

export interface AsuraApi {
  search: (params: AsuraSearchParams) => Promise<ApiResponse<AsuraSearchResponse>>
}

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
  latestChapters: AsuraSearchItemChapter[]
  chapterCount: number
  rating: number
  releaseYear: number
}

export interface AsuraSearchItemChapter {
  earlyAccessUntil: Date
  publishedAt: Date
  title: string
  id: string
  number: number
}

export function createAsuraApi(): AsuraApi {
  const api = initApi('/sources/asura')

  async function search(params: AsuraSearchParams): Promise<ApiResponse<AsuraSearchResponse>> {
    try {
      const response = await api<AsuraSearchResponse>('/search', { method: 'GET', params: {
        ...(params.search && { search: params.search }),
        ...(params.sort && { sort: params.sort }),
        ...(params.status && { status: params.status }),
        ...(params.type && { type: params.type }),
        ...(params.artist && { artist: params.artist }),
        ...(params.offset >= 0 && { offset: params.offset }),
        ...(params.limit > 0 && { limit: params.limit }),
        ...(params.minChapters && params.minChapters > 0 && { minChapters: params.minChapters }),
      } })

      return { success: true, data: response }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  return { search }
}
