import type { ApiResponse } from '~/utils/api'
import { ApiError, initApi } from '~/utils/api'

export interface ChaptersApi {
  retryDownload: (chapterId: string) => Promise<ApiResponse<void>>
}

export function useChaptersApi(): ChaptersApi {
  const api = initApi('/chapters')

  async function retryDownload(chapterId: string): Promise<ApiResponse<void>> {
    try {
      await api(`/${chapterId}/retry`, { method: 'POST' })

      return { success: true, data: undefined }
    } catch (error) {
      return { success: false, error: ApiError.fromFetchError(error) }
    }
  }

  return {
    retryDownload,
  }
}
