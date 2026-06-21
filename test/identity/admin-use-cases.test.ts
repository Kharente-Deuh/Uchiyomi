// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment node
import { describe, expect, it, vi } from 'vitest'
import * as CreateUser from '../../server/domains/identity/users/application/usecases/create-user.use-case'
import * as SetUserStatus from '../../server/domains/identity/users/application/usecases/set-user-status.use-case'
import * as User from '../../server/domains/identity/users/user.domain'

const newUser = new User.Model({
  id: 'u2',
  email: 'u@b.c',
  displayName: 'U',
  role: 'USER',
  status: 'ACTIVE',
  canManageExtensions: false,
  canDownload: false,
  allowNsfw: false,
})

function fakeHasher(): { hash: ReturnType<typeof vi.fn>, verify: ReturnType<typeof vi.fn> } {
  return { hash: vi.fn(async (p: { password: string }) => `h:${p.password}`), verify: vi.fn(async () => true) }
}

describe('createUser', () => {
  it('hashes and creates a USER (default role), no onlyIfEmpty guard', async () => {
    const hasher = fakeHasher()
    const users = { createWithLocalIdentity: vi.fn(async () => newUser), countUsers: vi.fn(), findByEmail: vi.fn(), findById: vi.fn(), setStatus: vi.fn() }
    const result = await new CreateUser.UseCase(users, hasher).execute({ email: 'u@b.c', displayName: 'U', password: 'longenough1' })
    expect(hasher.hash).toHaveBeenCalledWith({ password: 'longenough1' })
    expect(users.createWithLocalIdentity).toHaveBeenCalledWith(
      expect.objectContaining({ email: 'u@b.c', role: 'USER', passwordHash: 'h:longenough1' }),
    )
    expect(result.id).toBe('u2')
  })
})

describe('setUserStatus', () => {
  it('revokes all sessions when disabling', async () => {
    const users = { setStatus: vi.fn(), countUsers: vi.fn(), findByEmail: vi.fn(), findById: vi.fn(), createWithLocalIdentity: vi.fn() }
    const sessions = { deleteAllForUser: vi.fn(), create: vi.fn(), findValid: vi.fn(), touch: vi.fn(), delete: vi.fn() }
    await new SetUserStatus.UseCase(users, sessions).execute({ userId: 'u2', status: 'DISABLED' })
    expect(users.setStatus).toHaveBeenCalledWith({ id: 'u2', status: 'DISABLED' })
    expect(sessions.deleteAllForUser).toHaveBeenCalledWith({ userId: 'u2' })
  })

  it('does not touch sessions when enabling', async () => {
    const users = { setStatus: vi.fn(), countUsers: vi.fn(), findByEmail: vi.fn(), findById: vi.fn(), createWithLocalIdentity: vi.fn() }
    const sessions = { deleteAllForUser: vi.fn(), create: vi.fn(), findValid: vi.fn(), touch: vi.fn(), delete: vi.fn() }
    await new SetUserStatus.UseCase(users, sessions).execute({ userId: 'u2', status: 'ACTIVE' })
    expect(users.setStatus).toHaveBeenCalledWith({ id: 'u2', status: 'ACTIVE' })
    expect(sessions.deleteAllForUser).not.toHaveBeenCalled()
  })
})
