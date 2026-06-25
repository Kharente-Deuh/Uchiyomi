// SPDX-License-Identifier: AGPL-3.0-or-later
// Flat ESM module, consumed via `import * as Session`.
export type SessionModelProps = Omit<SessionModel, 'shouldRefresh'>

export class SessionModel {
  declare id: string
  declare userId: string
  declare expiresAt: Date

  constructor(data: SessionModelProps) {
    Object.assign<SessionModelProps, SessionModelProps>(this, data)
  }

  /** True when the session is close enough to expiry that we should slide it forward. */
  shouldRefresh(now: Date, thresholdMs: number): boolean {
    return this.expiresAt.getTime() - now.getTime() <= thresholdMs
  }
}

export function newSessionExpiry(now: Date, ttlMs: number): Date {
  return new Date(now.getTime() + ttlMs)
}

export interface CreateSessionParams {
  userId: string
  expiresAt: Date
  userAgent?: string
  ip?: string
}

export interface FindValidSessionParams {
  sessionId: string
  now: Date
}

export interface TouchSessionParams {
  sessionId: string
  expiresAt: Date
}

export interface DeleteSessionParams {
  sessionId: string
}

export interface DeleteAllSessionsForUserParams {
  userId: string
}

export interface DeleteAllSessionsForUserExceptParams {
  userId: string
  exceptSessionId: string
}

export interface SessionsRepository {
  create: (p: CreateSessionParams) => Promise<SessionModel>
  findValid: (p: FindValidSessionParams) => Promise<SessionModel | undefined>
  touch: (p: TouchSessionParams) => Promise<void>
  delete: (p: DeleteSessionParams) => Promise<void>
  deleteAllForUser: (p: DeleteAllSessionsForUserParams) => Promise<void>
  deleteAllForUserExcept: (p: DeleteAllSessionsForUserExceptParams) => Promise<void>
}
