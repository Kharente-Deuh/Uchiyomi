import type { FeedParams, FeedResponse } from '../types'
import type { ApiResponse } from '~/utils/api'
import { ApiError, initApi } from '~/utils/api'

export interface FeedApi {
  getFeed: (params: FeedParams) => Promise<ApiResponse<FeedResponse>>
}

export function createFeedApi(): FeedApi {
  const api = initApi('/feed')

  async function getFeed(params: FeedParams): Promise<ApiResponse<FeedResponse>> {
    try {
      const { type, source, ...definedParams } = params
      const response = await api<FeedResponse>('/', { params: {
        ...definedParams,
        ...(type && { type }),
        ...(source && { source }),
      } })

      return {
        success: true,
        data: response,
      }
    } catch (error) {
      return {
        success: false,
        error: ApiError.fromFetchError(error),
      }
    }
  }

  return {
    getFeed,
  }
}
