import type { FeedParams, FeedResponse } from '../types'
import type { ApiResponse } from '~/utils/api'
import { ApiError, initApi } from '~/utils/api'
import { parseOptionalDate } from '~/utils/date.utils'

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
        data: {
          total: response.total,
          items: response.items.map(({ latestChapters, ...item }) => ({
            ...item,
            latestChapters: latestChapters.map(({ earlyAccessUntil, publishedAt, ...chapter }) => {
              const until = parseOptionalDate(earlyAccessUntil)

              return {
                ...chapter,
                publishedAt: new Date(publishedAt),
                ...(until && { earlyAccessUntil: until }),
              }
            }),
          })),
        },
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
