// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment node
import process from 'node:process'
import { PrismaPg } from '@prisma/adapter-pg'
import { afterAll, beforeEach, describe, expect, it } from 'vitest'
import { PrismaClient } from '../../prisma/generated/client'
import { PrismaSessionRepository } from '../../server/domains/identity/sessions/infrastructure/persistence/prisma/prisma-session.repository'
import { PrismaUserRepository } from '../../server/domains/identity/users/infrastructure/persistence/prisma/prisma-user.repository'

const connectionString = process.env.TEST_DATABASE_URL
const describeIf = connectionString ? describe : describe.skip

describeIf('prisma identity repositories (live DB)', () => {
  const prisma = new PrismaClient({ adapter: new PrismaPg({ connectionString }) })
  const users = new PrismaUserRepository(prisma)
  const sessions = new PrismaSessionRepository(prisma)

  beforeEach(async () => {
    await prisma.appUser.deleteMany()
  })
  afterAll(async () => {
    await prisma.appUser.deleteMany()
    await prisma.$disconnect()
  })

  it('createWithLocalIdentity persists a user + LOCAL identity and findByAccountName returns the hash', async () => {
    const user = await users.createWithLocalIdentity({ accountName: 'alice', displayName: 'A', password: 'x', role: 'ADMIN', passwordHash: 'HASH' })
    const found = await users.findByAccountName({ accountName: 'alice' })
    expect(found?.passwordHash).toBe('HASH')
    expect(found?.role).toBe('ADMIN')
    expect(user.showNsfw).toBe(false)
  })

  it('onlyIfEmpty rejects a second create', async () => {
    await users.createWithLocalIdentity({ accountName: 'alice', displayName: 'A', password: 'x', role: 'ADMIN', passwordHash: 'H' }, { onlyIfEmpty: true })
    await expect(
      users.createWithLocalIdentity({ accountName: 'bob', displayName: 'B', password: 'x', role: 'ADMIN', passwordHash: 'H' }, { onlyIfEmpty: true }),
    ).rejects.toBeTruthy()
    expect(await users.countUsers()).toBe(1)
  })

  it('session create/findValid/touch/delete and deleteAllForUser', async () => {
    const u = await users.createWithLocalIdentity({ accountName: 'alice', displayName: 'A', password: 'x', role: 'USER', passwordHash: 'H' })
    const now = new Date()
    const s = await sessions.create({ userId: u.id, expiresAt: new Date(now.getTime() + 60_000), ip: '127.0.0.1' })
    expect(await sessions.findValid({ sessionId: s.id, now })).not.toBeUndefined()
    expect(await sessions.findValid({ sessionId: s.id, now: new Date(now.getTime() + 120_000) })).toBeUndefined()
    await sessions.deleteAllForUser({ userId: u.id })
    expect(await sessions.findValid({ sessionId: s.id, now })).toBeUndefined()
  })

  it('updateDisplayName changes the name and findById reflects it', async () => {
    const u = await users.createWithLocalIdentity({ accountName: 'alice', displayName: 'Old', password: 'x', role: 'USER', passwordHash: 'H' })
    const updated = await users.updateDisplayName({ id: u.id, displayName: 'New Name' })
    expect(updated.displayName).toBe('New Name')
    const found = await users.findById({ id: u.id })
    expect(found?.displayName).toBe('New Name')
  })

  it('updateDisplayName throws for a missing id', async () => {
    await expect(users.updateDisplayName({ id: 'does-not-exist', displayName: 'X' })).rejects.toBeTruthy()
  })

  it('updateLocalPasswordHash + findLocalPasswordHash round-trip', async () => {
    const u = await users.createWithLocalIdentity({ accountName: 'bob', displayName: 'B', password: 'x', role: 'USER', passwordHash: 'H1' })
    expect(await users.findLocalPasswordHash({ userId: u.id })).toBe('H1')
    await users.updateLocalPasswordHash({ userId: u.id, passwordHash: 'H2' })
    expect(await users.findLocalPasswordHash({ userId: u.id })).toBe('H2')
  })

  it('deleteAllForUserExcept keeps the named session and deletes the rest', async () => {
    const u = await users.createWithLocalIdentity({ accountName: 'carol', displayName: 'C', password: 'x', role: 'USER', passwordHash: 'H' })
    const now = new Date()
    const keep = await sessions.create({ userId: u.id, expiresAt: new Date(now.getTime() + 60_000) })
    const drop = await sessions.create({ userId: u.id, expiresAt: new Date(now.getTime() + 60_000) })
    await sessions.deleteAllForUserExcept({ userId: u.id, exceptSessionId: keep.id })
    expect(await sessions.findValid({ sessionId: keep.id, now })).not.toBeUndefined()
    expect(await sessions.findValid({ sessionId: drop.id, now })).toBeUndefined()
  })
})
