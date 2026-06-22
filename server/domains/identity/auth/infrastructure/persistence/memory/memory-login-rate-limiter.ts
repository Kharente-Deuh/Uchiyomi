// SPDX-License-Identifier: AGPL-3.0-or-later
import type * as Auth from '../../../auth.domain'

interface Bucket {
  count: number
  resetAt: number
}

/**
 * In-memory fixed-window limiter (single-instance app). Key is typically
 * `${ip}:${accountName}`. Implements the `Auth.RateLimiter` port.
 */
export class MemoryLoginRateLimiter implements Auth.RateLimiter {
  private readonly maxAttempts: number
  private readonly windowMs: number
  private readonly now: () => number
  private readonly buckets = new Map<string, Bucket>()

  constructor(opts: { maxAttempts: number, windowMs: number, now?: () => number }) {
    this.maxAttempts = opts.maxAttempts
    this.windowMs = opts.windowMs
    this.now = opts.now ?? Date.now
  }

  /** Returns true if the attempt is allowed (and records it), false if rate-limited. */
  check(p: Auth.RateLimiterCheckParams): boolean {
    const t = this.now()
    const bucket = this.buckets.get(p.key)

    if (!bucket || t >= bucket.resetAt) {
      this.buckets.set(p.key, { count: 1, resetAt: t + this.windowMs })

      return true
    }

    if (bucket.count >= this.maxAttempts) {
      return false
    }

    bucket.count += 1

    return true
  }

  /** Clear a key — call on successful login so a good login doesn't count against the user. */
  reset(p: Auth.RateLimiterResetParams): void {
    this.buckets.delete(p.key)
  }
}
