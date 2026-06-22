// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment node
import { describe, expect, it } from 'vitest'
import * as User from '../../server/domains/identity/users/user.domain'

function user(overrides: Partial<User.ModelProps> = {}): User.Model {
  return new User.Model({
    id: 'u1',
    accountName: 'alice',
    displayName: 'A',
    role: 'USER',
    status: 'ACTIVE',
    canManageExtensions: false,
    canDownload: false,
    allowNsfw: false,
    ...overrides,
  })
}

describe('canManageUsers', () => {
  it('allows an ADMIN', () => {
    expect(user({ role: 'ADMIN' }).canManageUsers()).toBe(true)
  })
  it('denies a USER', () => {
    expect(user({ role: 'USER' }).canManageUsers()).toBe(false)
  })
})
