// SPDX-License-Identifier: AGPL-3.0-or-later
import type { FetchError } from 'ofetch'

interface H3ErrorData {
  statusCode?: number
  statusMessage?: string
  message?: string
}

// Front-only normalised error. NOT a wire contract — the wire stays H3's
// { statusCode, statusMessage, ... }. UI (M3.3) maps `status` to i18n copy.
export class ApiError extends Error {
  readonly status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }

  static fromFetchError(err: unknown): ApiError {
    const fetchErr = err as Partial<FetchError<H3ErrorData>>
    const status = fetchErr.statusCode ?? fetchErr.response?.status ?? 0
    const data = fetchErr.data
    const message = data?.statusMessage
      ?? data?.message
      ?? (err instanceof Error ? err.message : 'Request failed')

    return new ApiError(message, status)
  }
}

export type ApiResponse<T> = { success: true, data: T } | { success: false, error: ApiError }
