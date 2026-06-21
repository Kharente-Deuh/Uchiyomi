// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it } from 'vitest'
import { ApiError } from '~/utils/api/api-error'

describe('toApiError', () => {
  it('maps a FetchError statusCode + data.statusMessage to ApiError', () => {
    const fetchErr = Object.assign(new Error('boom'), {
      statusCode: 401,
      data: { statusCode: 401, statusMessage: 'Unauthenticated' },
    })
    const err = ApiError.fromFetchError(fetchErr)
    expect(err).toBeInstanceOf(ApiError)
    expect(err.status).toBe(401)
    expect(err.message).toBe('Unauthenticated')
  })

  it('falls back to status 0 and the error message when no HTTP info is present', () => {
    const err = ApiError.fromFetchError(new Error('network down'))
    expect(err.status).toBe(0)
    expect(err.message).toBe('network down')
  })

  it('prefers data.message when statusMessage is absent', () => {
    const fetchErr = Object.assign(new Error('boom'), {
      statusCode: 400,
      data: { message: 'Invalid body' },
    })
    expect(ApiError.fromFetchError(fetchErr).message).toBe('Invalid body')
  })
})
