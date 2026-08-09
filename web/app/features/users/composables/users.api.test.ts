// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createUsersApi } from './users.api'

const call = vi.fn()

vi.mock('~/utils/api', async importOriginal => ({
  ...(await importOriginal<typeof import('~/utils/api')>()),
  initApi: () => call,
}))

describe('createUsersApi().getCurrentUser', () => {
  beforeEach(() => {
    call.mockReset()
  })

  it('returns the user carried by /me', async () => {
    const user = { id: 'e2d1', username: 'alice', isAdmin: true }
    call.mockResolvedValue(user)

    await expect(createUsersApi().getCurrentUser()).resolves.toEqual({ success: true, data: user })
  })

  it('calls /me', async () => {
    call.mockResolvedValue({})
    await createUsersApi().getCurrentUser()

    expect(call).toHaveBeenCalledWith('/me')
  })

  it('surfaces a 401 as a failed response carrying the status', async () => {
    call.mockRejectedValue({ statusCode: 401, data: {} })

    const res = await createUsersApi().getCurrentUser()

    expect(res.success).toBe(false)
    expect(res.success === false && res.error.status).toBe(401)
  })

  it('reports a status of 0 when the API cannot be reached', async () => {
    call.mockRejectedValue(new TypeError('Failed to fetch'))

    const res = await createUsersApi().getCurrentUser()

    expect(res.success).toBe(false)
    expect(res.success === false && res.error.status).toBe(0)
  })
})
