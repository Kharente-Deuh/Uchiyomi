// SPDX-License-Identifier: AGPL-3.0-or-later

// Flat ESM module, consumed via `import * as Auth`.
export type ErrorCode = 'invalid_credentials' | 'setup_closed' | 'unauthenticated' | 'invalid_password'

export class AuthError extends Error {
  readonly code: ErrorCode

  constructor(code: ErrorCode, message: string) {
    super(message)
    this.name = 'AuthError'
    this.code = code
  }
}

export interface AuthRateLimiterCheckParams {
  key: string
}

export interface AuthRateLimiterResetParams {
  key: string
}

// The PORT — the in-memory adapter implements it.
export interface AuthRateLimiter {
  check: (p: AuthRateLimiterCheckParams) => boolean
  reset: (p: AuthRateLimiterResetParams) => void
}
