import type { Chapter } from '../types'
import type { ApiResponse } from '~/utils/api'
import { ApiError, initApi } from '~/utils/api'

export interface ChaptersApi {
  retryDownload: (chapterId: string) => Promise<ApiResponse<void>>
  getByIds: (ids: string[]) => Promise<ApiResponse<Chapter[]>>
  getByComicId: (comicId: string) => Promise<ApiResponse<Chapter[]>>
}

export function createChaptersApi(): ChaptersApi {
  const api = initApi('/chapters')

  async function retryDownload(chapterId: string): Promise<ApiResponse<void>> {
    try {
      await api(`/${chapterId}/retry`, { method: 'POST' })

      return { success: true, data: undefined }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function getByIds(ids: string[]): Promise<ApiResponse<Chapter[]>> {
    try {
      const response = await api<Chapter[]>(`/list`, { method: 'POST', body: { ids } })

      return { success: true, data: response }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  async function getByComicId(comicId: string): Promise<ApiResponse<Chapter[]>> {
    try {
      const response = await api<Chapter[]>(`/`, { params: { comicId } })

      return { success: true, data: response }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  return {
    retryDownload,
    getByIds,
    getByComicId,
  }
}
