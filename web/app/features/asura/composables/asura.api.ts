// SPDX-License-Identifier: AGPL-3.0-or-later

import type { AsuraComicChapter, AsuraComicInfos, AsuraSearchParams, AsuraSearchResponse } from '../types'
import type { ApiResponse } from '~/utils/api'
import { ApiError, initApi } from '~/utils/api'

export interface AsuraApi {
  search: (params: AsuraSearchParams) => Promise<ApiResponse<AsuraSearchResponse>>
  getInfosBySlug: (slug: string) => Promise<ApiResponse<AsuraComicInfos>>
  getSeriesChapters: (slug: string) => Promise<ApiResponse<AsuraComicChapter[]>>
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

      return {
        success: true,
        data: {
          total: response.total,
          items: response.items.map(({ lastChapterAt, updatedAt, createdAt, latestChapters, ...rest }) => ({
            ...rest,
            ...(lastChapterAt && { lastChapterAt: new Date(lastChapterAt) }),
            updatedAt: new Date(updatedAt),
            createdAt: new Date(createdAt),
            latestChapters: latestChapters.map(({ earlyAccessUntil, publishedAt, ...rest }) => ({
              ...rest,
              ...(earlyAccessUntil && { earlyAccessUntil: new Date(earlyAccessUntil) }),
              publishedAt: new Date(publishedAt),
            })),
          })),
        },
      }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function getInfosBySlug(slug: string): Promise<ApiResponse<AsuraComicInfos>> {
    try {
      const {
        lastChapterAt,
        updatedAt,
        createdAt,
        ...rest
      } = await api<AsuraComicInfos>(`/series/${slug}`, { method: 'GET' })

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

  async function getSeriesChapters(slug: string): Promise<ApiResponse<AsuraComicChapter[]>> {
    try {
      const response = await api<AsuraComicChapter[]>(`/series/${slug}/chapters`, { method: 'GET' })

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
