// SPDX-License-Identifier: AGPL-3.0-or-later

import type { H3Event } from 'h3'
import type { UserModel } from '../../server/domains/identity/users/user.domain'
import { createError } from 'h3'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { authGuard } from '../../server/domains/identity/auth/infrastructure/http/guards/auth.guard'

// authGuard relies on Nitro's auto-imported `createError`; the plain node test
// project has no auto-imports, so provide the real h3 implementation as a global.
beforeEach(() => {
  vi.stubGlobal('createError', createError)
})

afterEach(() => {
  vi.unstubAllGlobals()
})

function eventWith(authUser?: Partial<UserModel>): H3Event {
  return { context: { authUser } } as unknown as H3Event
}

function thrown(fn: () => unknown): { statusCode?: number } | undefined {
  try {
    fn()

    return undefined
  } catch (error) {
    return error as { statusCode?: number }
  }
}

describe('authGuard', () => {
  it('throws 401 when the request is unauthenticated', () => {
    expect(thrown(() => authGuard(eventWith()))?.statusCode).toBe(401)
  })

  it('returns the authenticated user when no capability is required', () => {
    const authUser = { id: 'u1', canManageExtensions: false } as UserModel
    expect(authGuard(eventWith(authUser))).toBe(authUser)
  })

  it('throws 403 when management is required but the user lacks the capability', () => {
    const authUser = { id: 'u1', canManageExtensions: false } as UserModel
    expect(thrown(() => authGuard(eventWith(authUser), { mustBeAbleToManage: true }))?.statusCode).toBe(403)
  })

  it('returns the user when management is required and granted', () => {
    const authUser = { id: 'u1', canManageExtensions: true } as UserModel
    expect(authGuard(eventWith(authUser), { mustBeAbleToManage: true })).toBe(authUser)
  })
})
