// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment node
import { describe, expect, it, vi } from 'vitest'
import * as Login from '../../server/domains/identity/auth/application/usecases/login.use-case'
import * as SetupFirstAdmin from '../../server/domains/identity/auth/application/usecases/setup-first-admin.use-case'
import * as GetCurrentUser from '../../server/domains/identity/sessions/application/usecases/get-current-user.use-case'
import * as Session from '../../server/domains/identity/sessions/session.domain'
import * as User from '../../server/domains/identity/users/user.domain'

const TTL_MS = 30 * 24 * 60 * 60 * 1000
const REFRESH_MS = 7 * 24 * 60 * 60 * 1000

const adminProps: User.UserModelProps = {
  id: 'u1',
  accountName: 'alice',
  displayName: 'Admin',
  role: 'ADMIN',
  status: 'ACTIVE',
  canManageExtensions: true,
  canDownload: true,
  allowNsfw: true,
}

function admin(overrides: Partial<User.UserModelProps> = {}): User.UserModel {
  return new User.UserModel({ ...adminProps, ...overrides })
}

function withHash(overrides: Partial<User.UserModelProps> = {}): User.UserModel {
  return admin({ passwordHash: 'h:longenough1', ...overrides })
}

function fakeHasher(): { hash: ReturnType<typeof vi.fn>, verify: ReturnType<typeof vi.fn> } {
  return {
    hash: vi.fn(async (p: { password: string }) => `h:${p.password}`),
    verify: vi.fn(async (p: { hash: string, password: string }) => p.hash === `h:${p.password}`),
  }
}

describe('setupFirstAdmin', () => {
  it('hashes the password and creates the admin only-if-empty', async () => {
    const hasher = fakeHasher()
    const users = {
      createWithLocalIdentity: vi.fn(async () => admin()),
      countUsers: vi.fn(),
      findByAccountName: vi.fn(),
      findById: vi.fn(),
      setStatus: vi.fn(),
      updateDisplayName: vi.fn(),
      updateLocalPasswordHash: vi.fn(),
      findLocalPasswordHash: vi.fn(),
    }
    const result = await new SetupFirstAdmin.SetupFirstAdminUseCase(users, hasher).execute({ accountName: 'alice', displayName: 'Admin', password: 'longenough1' })
    expect(hasher.hash).toHaveBeenCalledWith({ password: 'longenough1' })
    expect(users.createWithLocalIdentity).toHaveBeenCalledWith(
      expect.objectContaining({ accountName: 'alice', role: 'ADMIN', passwordHash: 'h:longenough1' }),
      { onlyIfEmpty: true },
    )
    expect(result.role).toBe('ADMIN')
  })
})

describe('login', () => {
  it('throws invalid_credentials when the user is missing', async () => {
    const users = { findByAccountName: vi.fn(async () => {}), countUsers: vi.fn(), findById: vi.fn(), createWithLocalIdentity: vi.fn(), setStatus: vi.fn(), updateDisplayName: vi.fn(), updateLocalPasswordHash: vi.fn(), findLocalPasswordHash: vi.fn() }
    const sessions = { create: vi.fn(), findValid: vi.fn(), touch: vi.fn(), delete: vi.fn(), deleteAllForUser: vi.fn(), deleteAllForUserExcept: vi.fn() }
    await expect(new Login.LoginUseCase(users, sessions, fakeHasher(), TTL_MS).execute({ accountName: 'ghost', password: 'longenough1' }))
      .rejects
      .toMatchObject({ code: 'invalid_credentials' })
  })

  it('throws invalid_credentials for a disabled user (no session created)', async () => {
    const users = { findByAccountName: vi.fn(async () => withHash({ status: 'DISABLED' })), countUsers: vi.fn(), findById: vi.fn(), createWithLocalIdentity: vi.fn(), setStatus: vi.fn(), updateDisplayName: vi.fn(), updateLocalPasswordHash: vi.fn(), findLocalPasswordHash: vi.fn() }
    const sessions = { create: vi.fn(), findValid: vi.fn(), touch: vi.fn(), delete: vi.fn(), deleteAllForUser: vi.fn(), deleteAllForUserExcept: vi.fn() }
    await expect(new Login.LoginUseCase(users, sessions, fakeHasher(), TTL_MS).execute({ accountName: 'alice', password: 'longenough1' }))
      .rejects
      .toMatchObject({ code: 'invalid_credentials' })
    expect(sessions.create).not.toHaveBeenCalled()
  })

  it('creates a session with a TTL in the future on success', async () => {
    const now = new Date('2026-06-21T00:00:00Z')
    const created = new Session.SessionModel({ id: 's1', userId: 'u1', expiresAt: new Date(now.getTime() + TTL_MS) })
    const users = { findByAccountName: vi.fn(async () => withHash()), countUsers: vi.fn(), findById: vi.fn(), createWithLocalIdentity: vi.fn(), setStatus: vi.fn(), updateDisplayName: vi.fn(), updateLocalPasswordHash: vi.fn(), findLocalPasswordHash: vi.fn() }
    const sessions = { create: vi.fn(async () => created), findValid: vi.fn(), touch: vi.fn(), delete: vi.fn(), deleteAllForUser: vi.fn(), deleteAllForUserExcept: vi.fn() }
    const result = await new Login.LoginUseCase(users, sessions, fakeHasher(), TTL_MS, () => now).execute({ accountName: 'alice', password: 'longenough1' })
    expect(sessions.create).toHaveBeenCalledWith(expect.objectContaining({ userId: 'u1', expiresAt: new Date(now.getTime() + TTL_MS) }))
    expect(result.id).toBe('s1')
  })
})

describe('getCurrentUser', () => {
  it('throws unauthenticated when the session is missing/expired', async () => {
    const users = { findById: vi.fn(), countUsers: vi.fn(), findByAccountName: vi.fn(), createWithLocalIdentity: vi.fn(), setStatus: vi.fn(), updateDisplayName: vi.fn(), updateLocalPasswordHash: vi.fn(), findLocalPasswordHash: vi.fn() }
    const sessions = { findValid: vi.fn(async () => {}), create: vi.fn(), touch: vi.fn(), delete: vi.fn(), deleteAllForUser: vi.fn(), deleteAllForUserExcept: vi.fn() }
    await expect(new GetCurrentUser.GetCurrentUserUseCase(users, sessions, TTL_MS, REFRESH_MS).execute({ sessionId: 's-missing' })).rejects.toMatchObject({ code: 'unauthenticated' })
  })

  it('returns the user and slides expiry when within the refresh threshold', async () => {
    const now = new Date('2026-06-21T00:00:00Z')
    // expires in 1 day → within the 7-day refresh threshold → should be touched
    const session = new Session.SessionModel({ id: 's1', userId: 'u1', expiresAt: new Date(now.getTime() + 24 * 60 * 60 * 1000) })
    const sessions = { findValid: vi.fn(async () => session), touch: vi.fn(), create: vi.fn(), delete: vi.fn(), deleteAllForUser: vi.fn(), deleteAllForUserExcept: vi.fn() }
    const users = { findById: vi.fn(async () => admin()), countUsers: vi.fn(), findByAccountName: vi.fn(), createWithLocalIdentity: vi.fn(), setStatus: vi.fn(), updateDisplayName: vi.fn(), updateLocalPasswordHash: vi.fn(), findLocalPasswordHash: vi.fn() }
    const result = await new GetCurrentUser.GetCurrentUserUseCase(users, sessions, TTL_MS, REFRESH_MS, () => now).execute({ sessionId: 's1' })
    expect(result.id).toBe('u1')
    expect(sessions.touch).toHaveBeenCalledWith({ sessionId: 's1', expiresAt: new Date(now.getTime() + TTL_MS) })
  })

  it('throws unauthenticated when the user is disabled', async () => {
    const now = new Date('2026-06-21T00:00:00Z')
    const session = new Session.SessionModel({ id: 's1', userId: 'u1', expiresAt: new Date(now.getTime() + TTL_MS) })
    const sessions = { findValid: vi.fn(async () => session), touch: vi.fn(), create: vi.fn(), delete: vi.fn(), deleteAllForUser: vi.fn(), deleteAllForUserExcept: vi.fn() }
    const users = { findById: vi.fn(async () => admin({ status: 'DISABLED' })), countUsers: vi.fn(), findByAccountName: vi.fn(), createWithLocalIdentity: vi.fn(), setStatus: vi.fn(), updateDisplayName: vi.fn(), updateLocalPasswordHash: vi.fn(), findLocalPasswordHash: vi.fn() }
    await expect(new GetCurrentUser.GetCurrentUserUseCase(users, sessions, TTL_MS, REFRESH_MS, () => now).execute({ sessionId: 's1' })).rejects.toMatchObject({ code: 'unauthenticated' })
  })
})
