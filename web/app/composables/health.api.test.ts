// SPDX-License-Identifier: AGPL-3.0-or-later

import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createHealthApi } from './health.api'

const raw = vi.fn()

vi.mock('~/utils/api', async importOriginal => ({
  ...(await importOriginal<typeof import('~/utils/api')>()),
  initApi: () => ({ raw }),
}))

describe('createHealthApi().getServerStatus', () => {
  beforeEach(() => {
    raw.mockReset()
  })

  it('returns the payload when it carries a known status', async () => {
    const payload = { status: 'ok', components: { db: { status: 'ok' } } }
    raw.mockResolvedValue({ _data: payload })

    await expect(createHealthApi().getServerStatus()).resolves.toEqual(payload)
  })

  it('keeps a failed status and its component reasons', async () => {
    const payload = { status: 'failed', components: { db: { status: 'failed', reason: 'connection refused' } } }
    raw.mockResolvedValue({ _data: payload })

    await expect(createHealthApi().getServerStatus()).resolves.toEqual(payload)
  })

  it('does not let a non-2xx response throw', async () => {
    raw.mockResolvedValue({ _data: { status: 'starting', components: {} } })
    await createHealthApi().getServerStatus()

    expect(raw).toHaveBeenCalledWith('/readyz', expect.objectContaining({ ignoreResponseError: true }))
  })

  it('falls back to unreachable on an unknown status', async () => {
    raw.mockResolvedValue({ _data: { status: 'degraded' } })

    await expect(createHealthApi().getServerStatus()).resolves.toEqual({ status: 'unreachable' })
  })

  it('falls back to unreachable on a body that is not an object', async () => {
    raw.mockResolvedValue({ _data: 'ok' })

    await expect(createHealthApi().getServerStatus()).resolves.toEqual({ status: 'unreachable' })
  })

  it('falls back to unreachable on an empty body', async () => {
    raw.mockResolvedValue({ _data: null })

    await expect(createHealthApi().getServerStatus()).resolves.toEqual({ status: 'unreachable' })
  })

  it('falls back to unreachable when the request throws', async () => {
    raw.mockRejectedValue(new TypeError('Failed to fetch'))

    await expect(createHealthApi().getServerStatus()).resolves.toEqual({ status: 'unreachable' })
  })

  it('bounds the probe with a timeout', async () => {
    raw.mockResolvedValue({ _data: { status: 'ok', components: {} } })
    await createHealthApi().getServerStatus()

    expect(raw).toHaveBeenCalledWith('/readyz', expect.objectContaining({ timeout: 5000 }))
  })
})
