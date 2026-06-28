// SPDX-License-Identifier: AGPL-3.0-or-later

// @vitest-environment node
import { describe, expect, it } from 'vitest'
import { MemoryLoginRateLimiter } from '../../server/domains/identity/auth/infrastructure/persistence/memory/memory-login-rate-limiter'

describe('loginRateLimiter', () => {
  it('allows up to maxAttempts then blocks within the window', () => {
    const t = 0
    const rl = new MemoryLoginRateLimiter({ maxAttempts: 3, windowMs: 1000, now: () => t })
    expect(rl.check({ key: 'ip:a@b.c' })).toBe(true)
    expect(rl.check({ key: 'ip:a@b.c' })).toBe(true)
    expect(rl.check({ key: 'ip:a@b.c' })).toBe(true)
    expect(rl.check({ key: 'ip:a@b.c' })).toBe(false) // 4th within window
  })

  it('resets after the window elapses', () => {
    let t = 0
    const rl = new MemoryLoginRateLimiter({ maxAttempts: 1, windowMs: 1000, now: () => t })
    expect(rl.check({ key: 'k' })).toBe(true)
    expect(rl.check({ key: 'k' })).toBe(false)
    t = 1001
    expect(rl.check({ key: 'k' })).toBe(true)
  })

  it('reset() clears a key (used on successful login)', () => {
    const rl = new MemoryLoginRateLimiter({ maxAttempts: 1, windowMs: 1000, now: () => 0 })
    expect(rl.check({ key: 'k' })).toBe(true)
    rl.reset({ key: 'k' })
    expect(rl.check({ key: 'k' })).toBe(true)
  })
})
