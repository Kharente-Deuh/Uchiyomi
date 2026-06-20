// SPDX-License-Identifier: AGPL-3.0-or-later
import { SuwayomiError } from './errors'

export interface ResilienceOptions {
  timeoutMs: number
  retries: number
  backoffMs: number
  isRetryable: (err: unknown) => boolean
}

function delay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}

async function runWithTimeout<T>(fn: () => Promise<T>, timeoutMs: number): Promise<T> {
  let timer: ReturnType<typeof setTimeout>
  const timeout = new Promise<never>((_, reject) => {
    timer = setTimeout(
      () => reject(new SuwayomiError('timeout', `Suwayomi request timed out after ${timeoutMs}ms`)),
      timeoutMs,
    )
  })
  try {
    return await Promise.race([fn(), timeout])
  } finally {
    clearTimeout(timer!)
  }
}

/**
 * Run `fn` with a per-attempt timeout and bounded retry. A timeout is always
 *  retryable (until `retries` is exhausted); other failures retry only when
 *  `isRetryable` says so. The last error propagates; the caller maps it.
 */
export async function withResilience<T>(fn: () => Promise<T>, opts: ResilienceOptions): Promise<T> {
  let lastError: unknown
  for (let attempt = 0; attempt <= opts.retries; attempt++) {
    try {
      return await runWithTimeout(fn, opts.timeoutMs)
    } catch (err) {
      lastError = err
      const isTimeout = err instanceof SuwayomiError && err.kind === 'timeout'
      const retryable = isTimeout || opts.isRetryable(err)
      if (!retryable || attempt === opts.retries) {
        throw err
      }

      await delay(opts.backoffMs * 2 ** attempt)
    }
  }

  throw lastError // unreachable, but satisfies the type checker
}
