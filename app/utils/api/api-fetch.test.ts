// SPDX-License-Identifier: AGPL-3.0-or-later
import { afterEach, describe, expect, it, vi } from 'vitest'
import { __resetSessionLost, handleResponseError, onSessionLost } from './api-fetch'

function ctx(status: number, url: string): Parameters<typeof handleResponseError>[0] {
  return { request: url, response: { status } } as Parameters<typeof handleResponseError>[0]
}

describe('handleResponseError', () => {
  afterEach(() => __resetSessionLost())

  it('invokes the session-loss handler on a 401 from a non-auth route', () => {
    const handler = vi.fn()
    onSessionLost(handler)
    handleResponseError(ctx(401, '/api/library'))
    expect(handler).toHaveBeenCalledTimes(1)
  })

  it('ignores 401s from /api/auth/* routes (expected during boot/login)', () => {
    const handler = vi.fn()
    onSessionLost(handler)
    handleResponseError(ctx(401, '/api/auth/me'))
    expect(handler).not.toHaveBeenCalled()
  })

  it('ignores non-401 responses', () => {
    const handler = vi.fn()
    onSessionLost(handler)
    handleResponseError(ctx(500, '/api/library'))
    expect(handler).not.toHaveBeenCalled()
  })

  it('is a no-op when no handler is registered', () => {
    expect(() => handleResponseError(ctx(401, '/api/library'))).not.toThrow()
  })
})
