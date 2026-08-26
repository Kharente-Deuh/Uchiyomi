// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ComicSource } from '~/features/comics/types'
import type { SourceComicChapter, SourceComicInfos, SourceSearchParams, SourceSearchResponse } from '../types'
import type { ApiResponse } from '~/utils/api'
import { ApiError, initApi } from '~/utils/api'

export interface SourceApi {
  search: (params: SourceSearchParams) => Promise<ApiResponse<SourceSearchResponse>>
  getInfosBySlug: (slug: string) => Promise<ApiResponse<SourceComicInfos>>
  getSeriesChapters: (slug: string) => Promise<ApiResponse<SourceComicChapter[]>>
}

export function createSourceApi(sourceId: ComicSource): SourceApi {
  const api = initApi(`/sources/${sourceId}`)

  async function search(params: SourceSearchParams): Promise<ApiResponse<SourceSearchResponse>> {
    try {
      const response = await api<SourceSearchResponse>('/search', {
        method: 'GET',
        params: {
          ...(params.search && { search: params.search }),
          ...(params.sort && { sort: params.sort }),
          ...(params.status && { status: params.status }),
          ...(params.type && { type: params.type }),
          ...(params.artist && { artist: params.artist }),
          ...(params.page >= 1 && { page: params.page }),
          ...(params.minChapters && params.minChapters > 0 && { minChapters: params.minChapters }),
        },
      })

      return {
        success: true,
        data: {
          hasNextPage: response.hasNextPage,
          items: response.items.map(({ lastChapterAt, updatedAt, createdAt, latestChapters, ...rest }) => ({
            ...rest,
            ...(lastChapterAt && { lastChapterAt: new Date(lastChapterAt) }),
            updatedAt: new Date(updatedAt),
            createdAt: new Date(createdAt),
            ...(latestChapters && {
              latestChapters: latestChapters.map(({ earlyAccessUntil, publishedAt, ...chapterRest }) => ({
                ...chapterRest,
                ...(earlyAccessUntil && { earlyAccessUntil: new Date(earlyAccessUntil) }),
                publishedAt: new Date(publishedAt),
              })),
            }),
          })),
        },
      }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function getInfosBySlug(slug: string): Promise<ApiResponse<SourceComicInfos>> {
    try {
      const {
        lastChapterAt,
        updatedAt,
        createdAt,
        ...rest
      } = await api<SourceComicInfos>(`/series/${slug}`, { method: 'GET' })

      return {
        success: true,
        data: {
          ...rest,
          ...(lastChapterAt && { lastChapterAt: new Date(lastChapterAt) }),
          updatedAt: new Date(updatedAt),
          createdAt: new Date(createdAt),
        },
      }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function getSeriesChapters(slug: string): Promise<ApiResponse<SourceComicChapter[]>> {
    try {
      const response = await api<SourceComicChapter[]>(`/series/${slug}/chapters`, { method: 'GET' })

      return {
        success: true,
        data: response.map(({ earlyAccessUntil, publishedAt, ...rest }) => ({
          ...rest,
          ...(earlyAccessUntil && { earlyAccessUntil: new Date(earlyAccessUntil) }),
          publishedAt: new Date(publishedAt),
        })),
      }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  return {
    search,
    getInfosBySlug,
    getSeriesChapters,
  }
}
