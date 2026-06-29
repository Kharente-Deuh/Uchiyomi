// SPDX-License-Identifier: AGPL-3.0-or-later

// @vitest-environment node
import { describe, expect, it, vi } from 'vitest'
import { SuwayomiError } from '../../server/utils/suwayomi/errors'
import { withResilience } from '../../server/utils/suwayomi/resilience'

const RETRY_ALL = { timeoutMs: 1000, retries: 2, backoffMs: 0, isRetryable: () => true }

describe('withResilience', () => {
  it('returns the result when the function succeeds', async () => {
    const fn = vi.fn(async () => 'ok')
    await expect(withResilience(fn, RETRY_ALL)).resolves.toBe('ok')
    expect(fn).toHaveBeenCalledTimes(1)
  })

  it('retries a retryable failure up to `retries` times, then succeeds', async () => {
    const fn = vi.fn()
      .mockRejectedValueOnce(new Error('net'))
      .mockRejectedValueOnce(new Error('net'))
      .mockResolvedValueOnce('ok')
    await expect(withResilience(fn, RETRY_ALL)).resolves.toBe('ok')
    expect(fn).toHaveBeenCalledTimes(3)
  })

  it('throws the last error after exhausting retries', async () => {
    const fn = vi.fn(async () => {
      throw new Error('net')
    })
    await expect(withResilience(fn, RETRY_ALL)).rejects.toThrow('net')
    expect(fn).toHaveBeenCalledTimes(3) // initial + 2 retries
  })

  it('does not retry a non-retryable failure', async () => {
    const fn = vi.fn(async () => {
      throw new Error('graphql')
    })
    const opts = { ...RETRY_ALL, isRetryable: () => false }
    await expect(withResilience(fn, opts)).rejects.toThrow('graphql')
    expect(fn).toHaveBeenCalledTimes(1)
  })

  it('rejects with a timeout SuwayomiError when the function exceeds timeoutMs', async () => {
    const fn = vi.fn(() => new Promise<string>(() => {})) // never resolves
    const opts = { timeoutMs: 10, retries: 0, backoffMs: 0, isRetryable: () => true }
    await expect(withResilience(fn, opts)).rejects.toMatchObject({ kind: 'timeout' })
    await expect(withResilience(fn, opts)).rejects.toBeInstanceOf(SuwayomiError)
  })
})
