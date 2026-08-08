// SPDX-License-Identifier: AGPL-3.0-or-later

import type { FetchError } from 'ofetch'

interface H3ErrorData {
  statusCode?: number
  statusMessage?: string
  message?: string
}

export class ApiError extends Error {
  static fromFetchError(error: unknown): ApiError {
    const fetchError = error as Partial<FetchError<H3ErrorData>>
    const status = fetchError.statusCode ?? fetchError.response?.status ?? 0
    const data = fetchError.data
    const message = data?.statusMessage
      ?? data?.message
      ?? (error instanceof Error ? error.message : 'Request failed')

    return new ApiError(message, status)
  }

  readonly status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export type ApiResponse<T> = { success: true, data: T } | { success: false, error: ApiError }
