// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { ApiError } from '~/utils/api/api-error'

describe('toApiError', () => {
  it('maps a FetchError statusCode + data.statusMessage to ApiError', () => {
    const fetchError = Object.assign(new Error('boom'), {
      statusCode: 401,
      data: { statusCode: 401, statusMessage: 'Unauthenticated' },
    })
    const error = ApiError.fromFetchError(fetchError)
    expect(error).toBeInstanceOf(ApiError)
    expect(error.status).toBe(401)
    expect(error.message).toBe('Unauthenticated')
  })

  it('falls back to status 0 and the error message when no HTTP info is present', () => {
    const error = ApiError.fromFetchError(new Error('network down'))
    expect(error.status).toBe(0)
    expect(error.message).toBe('network down')
  })

  it('prefers data.message when statusMessage is absent', () => {
    const fetchError = Object.assign(new Error('boom'), {
      statusCode: 400,
      data: { message: 'Invalid body' },
    })
    expect(ApiError.fromFetchError(fetchError).message).toBe('Invalid body')
  })
})
