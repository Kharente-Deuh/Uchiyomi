// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment node
import { describe, expect, it, vi } from 'vitest'
import * as ChangePassword from '../../server/domains/identity/auth/application/usecases/change-password.use-case'
import * as UpdateDisplayName from '../../server/domains/identity/users/application/usecases/update-display-name.use-case'
import * as User from '../../server/domains/identity/users/user.domain'

function userMock(over: Record<string, unknown> = {}): User.Repository {
  return {
    countUsers: vi.fn(),
    findByAccountName: vi.fn(),
    findById: vi.fn(),
    createWithLocalIdentity: vi.fn(),
    setStatus: vi.fn(),
    updateDisplayName: vi.fn(),
    updateLocalPasswordHash: vi.fn(),
    findLocalPasswordHash: vi.fn(),
    ...over,
  } as unknown as User.Repository
}

function sessionMock(over: Record<string, unknown> = {}): any {
  return { create: vi.fn(), findValid: vi.fn(), touch: vi.fn(), delete: vi.fn(), deleteAllForUser: vi.fn(), deleteAllForUserExcept: vi.fn(), ...over }
}

function hasher(): { hash: ReturnType<typeof vi.fn>, verify: ReturnType<typeof vi.fn> } {
  return {
    hash: vi.fn(async (p: { password: string }) => `h:${p.password}`),
    verify: vi.fn(async (p: { hash: string, password: string }) => p.hash === `h:${p.password}`),
  }
}

describe('changePassword', () => {
  const base = { userId: 'u1', currentSessionId: 's1' }

  it('throws invalid_password when the current password is wrong', async () => {
    const users = userMock({ findLocalPasswordHash: vi.fn(async () => 'h:right') })
    const sessions = sessionMock()
    await expect(
      new ChangePassword.UseCase(users, sessions, hasher()).execute({ ...base, currentPassword: 'wrong', newPassword: 'newpass1234', logoutOtherDevices: false }),
    ).rejects.toMatchObject({ code: 'invalid_password' })
    expect(users.updateLocalPasswordHash).not.toHaveBeenCalled()
    expect(sessions.deleteAllForUserExcept).not.toHaveBeenCalled()
  })

  it('throws invalid_password when there is no local hash', async () => {
    const users = userMock({ findLocalPasswordHash: vi.fn(async () => {}) })
    await expect(
      new ChangePassword.UseCase(users, sessionMock(), hasher()).execute({ ...base, currentPassword: 'whatever', newPassword: 'newpass1234', logoutOtherDevices: false }),
    ).rejects.toMatchObject({ code: 'invalid_password' })
  })

  it('hashes and stores the new password, no revocation when logoutOtherDevices is false', async () => {
    const users = userMock({ findLocalPasswordHash: vi.fn(async () => 'h:current123') })
    const sessions = sessionMock()
    const h = hasher()
    await new ChangePassword.UseCase(users, sessions, h).execute({ ...base, currentPassword: 'current123', newPassword: 'newpass1234', logoutOtherDevices: false })
    expect(h.hash).toHaveBeenCalledWith({ password: 'newpass1234' })
    expect(users.updateLocalPasswordHash).toHaveBeenCalledWith({ userId: 'u1', passwordHash: 'h:newpass1234' })
    expect(sessions.deleteAllForUserExcept).not.toHaveBeenCalled()
  })

  it('revokes other sessions (keeping the current) when logoutOtherDevices is true', async () => {
    const users = userMock({ findLocalPasswordHash: vi.fn(async () => 'h:current123') })
    const sessions = sessionMock()
    await new ChangePassword.UseCase(users, sessions, hasher()).execute({ ...base, currentPassword: 'current123', newPassword: 'newpass1234', logoutOtherDevices: true })
    expect(sessions.deleteAllForUserExcept).toHaveBeenCalledWith({ userId: 'u1', exceptSessionId: 's1' })
  })
})

describe('updateDisplayName', () => {
  it('delegates to the repository and returns the updated user', async () => {
    const updated = new User.Model({
      id: 'u1',
      accountName: 'alice',
      displayName: 'New',
      role: 'USER',
      status: 'ACTIVE',
      canManageExtensions: false,
      canDownload: false,
      allowNsfw: false,
    })
    const users = userMock({ updateDisplayName: vi.fn(async () => updated) })
    const result = await new UpdateDisplayName.UseCase(users).execute({ id: 'u1', displayName: 'New' })
    expect(users.updateDisplayName).toHaveBeenCalledWith({ id: 'u1', displayName: 'New' })
    expect(result.displayName).toBe('New')
  })
})
