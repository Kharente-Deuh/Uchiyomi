// SPDX-License-Identifier: AGPL-3.0-or-later

import type { LightComic, SearchComicParams, SearchComicResponse } from '../types'
import type { ApiResponse } from '~/utils/api'
import { ApiError, initApi } from '~/utils/api'

export interface ComicsApi {
  create: (params: CreateComicParams) => Promise<ApiResponse<LightComic>>
  deleteById: (id: string) => Promise<ApiResponse<void>>
  search: (params: SearchComicParams) => Promise<ApiResponse<SearchComicResponse>>
}

export interface CreateComicParams {
  source: string
  slug: string
}

export function createComicsApi(): ComicsApi {
  const api = initApi('/comics')

  async function create(params: CreateComicParams): Promise<ApiResponse<LightComic>> {
    try {
      const response = await api<LightComic>('/', { method: 'POST', body: params })

      return { success: true, data: response }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function deleteById(id: string): Promise<ApiResponse<void>> {
    try {
      await api<void>(`/${id}`, { method: 'DELETE' })

      return { success: true, data: undefined }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function search(params: SearchComicParams): Promise<ApiResponse<SearchComicResponse>> {
    try {
      const { search, source, type, status, ...definedParams } = params

      const response = await api<SearchComicResponse>('/', {
        method: 'GET',
        params: {
          ...definedParams,
          ...(search && { search }),
          ...(source && { source }),
          ...(type && { type }),
          ...(status && { status }),
        },
      })

      return { success: true, data: response }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  return {
    create,
    deleteById,
    search,
  }
}
