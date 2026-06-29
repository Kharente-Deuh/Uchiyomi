// SPDX-License-Identifier: AGPL-3.0-or-later

// @vitest-environment node
import { describe, expect, it, vi } from 'vitest'
import * as SetupFirstAdmin from '../../server/domains/identity/auth/application/usecases/setup-first-admin.use-case'
import * as CreateUser from '../../server/domains/identity/users/application/usecases/create-user.use-case'
import * as SetUserStatus from '../../server/domains/identity/users/application/usecases/set-user-status.use-case'
import * as UpdateUserCapabilities from '../../server/domains/identity/users/application/usecases/update-user-capabilities.use-case'
import * as User from '../../server/domains/identity/users/user.domain'

const newUser = new User.UserModel({
  id: 'u2',
  accountName: 'newbie',
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
    const users = { createWithLocalIdentity: vi.fn(async () => newUser), countUsers: vi.fn(), findByAccountName: vi.fn(), findById: vi.fn(), setStatus: vi.fn(), updateDisplayName: vi.fn(), updateLocalPasswordHash: vi.fn(), findLocalPasswordHash: vi.fn(), updateCapabilities: vi.fn(), updateNsfwPreference: vi.fn() }
    const result = await new CreateUser.CreateUserUseCase(users, hasher).execute({ accountName: 'newbie', displayName: 'U', password: 'longenough1' })
    expect(hasher.hash).toHaveBeenCalledWith({ password: 'longenough1' })
    expect(users.createWithLocalIdentity).toHaveBeenCalledWith(
      expect.objectContaining({ accountName: 'newbie', role: 'USER', passwordHash: 'h:longenough1' }),
    )
    expect(result.id).toBe('u2')
  })
})

describe('setUserStatus', () => {
  it('revokes all sessions when disabling', async () => {
    const users = { setStatus: vi.fn(), countUsers: vi.fn(), findByAccountName: vi.fn(), findById: vi.fn(), createWithLocalIdentity: vi.fn(), updateDisplayName: vi.fn(), updateLocalPasswordHash: vi.fn(), findLocalPasswordHash: vi.fn(), updateCapabilities: vi.fn(), updateNsfwPreference: vi.fn() }
    const sessions = { deleteAllForUser: vi.fn(), create: vi.fn(), findValid: vi.fn(), touch: vi.fn(), delete: vi.fn(), deleteAllForUserExcept: vi.fn() }
    await new SetUserStatus.SetUserStatusUseCase(users, sessions).execute({ userId: 'u2', status: 'DISABLED' })
    expect(users.setStatus).toHaveBeenCalledWith({ id: 'u2', status: 'DISABLED' })
    expect(sessions.deleteAllForUser).toHaveBeenCalledWith({ userId: 'u2' })
  })

  it('does not touch sessions when enabling', async () => {
    const users = { setStatus: vi.fn(), countUsers: vi.fn(), findByAccountName: vi.fn(), findById: vi.fn(), createWithLocalIdentity: vi.fn(), updateDisplayName: vi.fn(), updateLocalPasswordHash: vi.fn(), findLocalPasswordHash: vi.fn(), updateCapabilities: vi.fn(), updateNsfwPreference: vi.fn() }
    const sessions = { deleteAllForUser: vi.fn(), create: vi.fn(), findValid: vi.fn(), touch: vi.fn(), delete: vi.fn(), deleteAllForUserExcept: vi.fn() }
    await new SetUserStatus.SetUserStatusUseCase(users, sessions).execute({ userId: 'u2', status: 'ACTIVE' })
    expect(users.setStatus).toHaveBeenCalledWith({ id: 'u2', status: 'ACTIVE' })
    expect(sessions.deleteAllForUser).not.toHaveBeenCalled()
  })
})

describe('setupFirstAdmin', () => {
  it('creates the first admin with ADMIN role AND full capabilities (canManageExtensions, canDownload, allowNsfw)', async () => {
    const adminUser = new User.UserModel({
      id: 'u0',
      accountName: 'admin',
      displayName: 'Admin',
      role: 'ADMIN',
      status: 'ACTIVE',
      canManageExtensions: true,
      canDownload: true,
      allowNsfw: true,
      showNsfw: false,
    })
    const hasher = fakeHasher()
    const users = {
      createWithLocalIdentity: vi.fn(async () => adminUser),
      countUsers: vi.fn(),
      findByAccountName: vi.fn(),
      findById: vi.fn(),
      setStatus: vi.fn(),
      updateDisplayName: vi.fn(),
      updateLocalPasswordHash: vi.fn(),
      findLocalPasswordHash: vi.fn(),
      updateCapabilities: vi.fn(),
      updateNsfwPreference: vi.fn(),
    }
    const result = await new SetupFirstAdmin.SetupFirstAdminUseCase(users, hasher).execute({
      accountName: 'Admin',
      displayName: 'Admin',
      password: 'longenough1',
    })
    expect(users.createWithLocalIdentity).toHaveBeenCalledWith(
      expect.objectContaining({
        role: 'ADMIN',
        canManageExtensions: true,
        canDownload: true,
        allowNsfw: true,
      }),
      { onlyIfEmpty: true },
    )
    expect(result.canManageExtensions).toBe(true)
    expect(result.canDownload).toBe(true)
    expect(result.allowNsfw).toBe(true)
  })
})

it('update-user-capabilities delegates to the repository', async () => {
  const updated = { id: 'u1', allowNsfw: true } as never
  const repo = { updateCapabilities: vi.fn().mockResolvedValue(updated) } as never
  const useCase = new UpdateUserCapabilities.UpdateUserCapabilitiesUseCase(repo)

  const result = await useCase.execute({ id: 'u1', allowNsfw: true })

  expect(repo.updateCapabilities).toHaveBeenCalledWith({ id: 'u1', allowNsfw: true })
  expect(result).toBe(updated)
})
