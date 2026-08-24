// SPDX-License-Identifier: AGPL-3.0-or-later

import type { Comic, ComicProgress, LightComic, SearchComicParams, SearchComicResponse, SetChaptersProgressParams } from '../types'
import type { ApiResponse } from '~/utils/api'
import { ApiError, initApi } from '~/utils/api'

export interface ComicsApi {
  create: (params: CreateComicParams) => Promise<ApiResponse<LightComic>>
  deleteById: (id: string) => Promise<ApiResponse<void>>
  search: (params: SearchComicParams) => Promise<ApiResponse<SearchComicResponse>>
  getById: (id: string) => Promise<ApiResponse<Comic>>
  refreshById: (id: string) => Promise<ApiResponse<Comic>>
  getProgress: (id: string) => Promise<ApiResponse<ComicProgress>>
  setChaptersProgress: (params: SetChaptersProgressParams) => Promise<ApiResponse<void>>
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

  async function getById(id: string): Promise<ApiResponse<Comic>> {
    try {
      const response = await api<Comic>(`/${id}`, { method: 'GET' })

      return { success: true, data: response }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function refreshById(id: string): Promise<ApiResponse<Comic>> {
    try {
      const response = await api<Comic>(`/${id}/refresh`, { method: 'POST' })

      return { success: true, data: response }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function getProgress(id: string): Promise<ApiResponse<ComicProgress>> {
    try {
      const response = await api<ComicProgress>(`/${id}/progress`, { method: 'GET' })

      return {
        success: true,
        data: {
          continue: response.continue,
        },
      }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function setChaptersProgress({ comicId, ...body }: SetChaptersProgressParams): Promise<ApiResponse<void>> {
    try {
      await api<void>(`/${comicId}/progress`, { method: 'POST', body })

      return { success: true, data: undefined }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  return {
    create,
    deleteById,
    search,
    getById,
    refreshById,
    getProgress,
    setChaptersProgress,
  }
}
