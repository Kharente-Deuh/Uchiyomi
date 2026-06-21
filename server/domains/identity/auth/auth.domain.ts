// SPDX-License-Identifier: AGPL-3.0-or-later
// Flat ESM module, consumed via `import * as Auth`.
export type ErrorCode = 'invalid_credentials' | 'setup_closed' | 'unauthenticated'

export class AuthError extends Error {
  readonly code: ErrorCode

  constructor(code: ErrorCode, message: string) {
    super(message)
    this.name = 'AuthError'
    this.code = code
  }
}

export interface RateLimiterCheckParams {
  key: string
}

export interface RateLimiterResetParams {
  key: string
}

// The PORT — the in-memory adapter implements it.
export interface RateLimiter {
  check: (p: RateLimiterCheckParams) => boolean
  reset: (p: RateLimiterResetParams) => void
}
