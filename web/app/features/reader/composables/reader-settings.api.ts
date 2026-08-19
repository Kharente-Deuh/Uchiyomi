import type { GetReaderSettingsResponse, ReaderSettings } from '../types'
import type { ApiResponse } from '~/utils/api'
import { ApiError, initApi } from '~/utils/api'

interface ReaderSettingsApi {
  getReaderSettings: () => Promise<ApiResponse<ReaderSettings[]>>
  updateReaderSettings: (settings: ReaderSettings) => Promise<ApiResponse<ReaderSettings>>
}

export function createReaderSettingsApi(): ReaderSettingsApi {
  const api = initApi('/me/reader-settings')

  async function getReaderSettings(): Promise<ApiResponse<ReaderSettings[]>> {
    try {
      const res = await api<GetReaderSettingsResponse>('/')

      return {
        success: true,
        data: res.items,
      }
    } catch (error) {
      return {
        success: false,
        error: ApiError.fromFetchError(error),
      }
    }
  }

  async function updateReaderSettings({ type, ...body }: ReaderSettings): Promise<ApiResponse<ReaderSettings>> {
    try {
      const res = await api<ReaderSettings>(`/${type}`, { method: 'PUT', body })

      return {
        success: true,
        data: res,
      }
    } catch (error) {
      return {
        success: false,
        error: ApiError.fromFetchError(error),
      }
    }
  }

  return {
    getReaderSettings,
    updateReaderSettings,
  }
}
