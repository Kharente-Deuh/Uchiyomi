// SPDX-License-Identifier: AGPL-3.0-or-later
// Flat ESM module, consumed via `import * as Session`.
export type ModelProps = Omit<Model, 'shouldRefresh'>

export class Model {
  declare id: string
  declare userId: string
  declare expiresAt: Date

  constructor(data: ModelProps) {
    Object.assign<ModelProps, ModelProps>(this, data)
  }

  /** True when the session is close enough to expiry that we should slide it forward. */
  shouldRefresh(now: Date, thresholdMs: number): boolean {
    return this.expiresAt.getTime() - now.getTime() <= thresholdMs
  }
}

export function newExpiry(now: Date, ttlMs: number): Date {
  return new Date(now.getTime() + ttlMs)
}

export interface CreateParams {
  userId: string
  expiresAt: Date
  userAgent?: string
  ip?: string
}

export interface FindValidParams {
  sessionId: string
  now: Date
}

export interface TouchParams {
  sessionId: string
  expiresAt: Date
}

export interface DeleteParams {
  sessionId: string
}

export interface DeleteAllForUserParams {
  userId: string
}

export interface DeleteAllForUserExceptParams {
  userId: string
  exceptSessionId: string
}

export interface Repository {
  create: (p: CreateParams) => Promise<Model>
  findValid: (p: FindValidParams) => Promise<Model | undefined>
  touch: (p: TouchParams) => Promise<void>
  delete: (p: DeleteParams) => Promise<void>
  deleteAllForUser: (p: DeleteAllForUserParams) => Promise<void>
  deleteAllForUserExcept: (p: DeleteAllForUserExceptParams) => Promise<void>
}
