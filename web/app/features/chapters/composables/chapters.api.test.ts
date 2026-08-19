// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createChaptersApi } from './chapters.api'

const call = vi.fn()

vi.mock('~/utils/api', async importOriginal => ({
  ...(await importOriginal<typeof import('~/utils/api')>()),
  initApi: () => call,
}))

beforeEach(() => {
  call.mockReset()
})

describe('createChaptersApi().retryDownload', () => {
  it('posts retry for the chapter id', async () => {
    call.mockResolvedValue(undefined)

    const res = await createChaptersApi().retryDownload('ch-1')

    expect(call).toHaveBeenCalledWith('/ch-1/retry', { method: 'POST' })
    expect(res).toEqual({ success: true, data: undefined })
  })

  it('surfaces a 409 with its status', async () => {
    call.mockRejectedValue({ statusCode: 409, data: {} })

    const res = await createChaptersApi().retryDownload('ch-1')

    expect(res.success === false && res.error.status).toBe(409)
  })
})
