// SPDX-License-Identifier: AGPL-3.0-or-later

import type { KingOfShojoComicChapter, KingOfShojoComicInfos, KingOfShojoSearchParams, KingOfShojoSearchResponse } from '../types'
import type { ApiResponse } from '~/utils/api'
import { KING_OF_SHOJO_SOURCE_NAME } from '~/constants'
import { ApiError, initApi } from '~/utils/api'

export interface KingOfShojoApi {
  search: (params: KingOfShojoSearchParams) => Promise<ApiResponse<KingOfShojoSearchResponse>>
  getInfosBySlug: (slug: string) => Promise<ApiResponse<KingOfShojoComicInfos>>
  getSeriesChapters: (slug: string) => Promise<ApiResponse<KingOfShojoComicChapter[]>>
}

export function createKingOfShojoApi(): KingOfShojoApi {
  const api = initApi(`/sources/${KING_OF_SHOJO_SOURCE_NAME}`)

  async function search(params: KingOfShojoSearchParams): Promise<ApiResponse<KingOfShojoSearchResponse>> {
    try {
      const response = await api<KingOfShojoSearchResponse>('/search', { method: 'GET', params: {
        ...(params.search && { search: params.search }),
        ...(params.sort && { sort: params.sort }),
        ...(params.status && { status: params.status }),
        ...(params.type && { type: params.type }),
        ...(params.artist && { artist: params.artist }),
        ...(params.page >= 1 && { page: params.page }),
      } })

      return {
        success: true,
        data: {
          hasNextPage: response.hasNextPage,
          items: response.items.map(({ lastChapterAt, updatedAt, createdAt, ...rest }) => ({
            ...rest,
            ...(lastChapterAt && { lastChapterAt: new Date(lastChapterAt) }),
            updatedAt: new Date(updatedAt),
            createdAt: new Date(createdAt),
          })),
        },
      }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function getInfosBySlug(slug: string): Promise<ApiResponse<KingOfShojoComicInfos>> {
    try {
      const {
        lastChapterAt,
        updatedAt,
        createdAt,
        ...rest
      } = await api<KingOfShojoComicInfos>(`/series/${slug}`, { method: 'GET' })

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

  async function getSeriesChapters(slug: string): Promise<ApiResponse<KingOfShojoComicChapter[]>> {
    try {
      const response = await api<KingOfShojoComicChapter[]>(`/series/${slug}/chapters`, { method: 'GET' })

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
