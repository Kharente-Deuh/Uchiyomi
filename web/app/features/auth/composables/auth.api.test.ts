// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createAuthApi } from './auth.api'

const call = vi.fn()

vi.mock('~/utils/api', async importOriginal => ({
  ...(await importOriginal<typeof import('~/utils/api')>()),
  initApi: () => call,
}))

describe('createAuthApi().loginWithPwd', () => {
  beforeEach(() => {
    call.mockReset()
  })

  const user = { id: 'e2d1', username: 'alice', isAdmin: true }

  it('returns the user carried by a successful login', async () => {
    call.mockResolvedValue(user)

    await expect(createAuthApi().loginWithPwd({ username: 'a', password: 'b' })).resolves.toEqual({
      status: 'ok',
      user,
    })
  })

  it('returns invalid-credentials on a 401', async () => {
    call.mockRejectedValue({ statusCode: 401, data: {} })

    await expect(createAuthApi().loginWithPwd({ username: 'a', password: 'b' })).resolves.toEqual({
      status: 'invalid-credentials',
    })
  })

  it('returns unknown-error on a 500', async () => {
    call.mockRejectedValue({ statusCode: 500, data: {} })

    await expect(createAuthApi().loginWithPwd({ username: 'a', password: 'b' })).resolves.toEqual({
      status: 'unknown-error',
    })
  })
})
