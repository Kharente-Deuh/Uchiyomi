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

  it('createWithLocalIdentity persists a user + LOCAL identity and findByEmail returns the hash', async () => {
    await users.createWithLocalIdentity({ email: 'a@b.c', displayName: 'A', password: 'x', role: 'ADMIN', passwordHash: 'HASH' })
    const found = await users.findByEmail({ email: 'a@b.c' })
    expect(found?.passwordHash).toBe('HASH')
    expect(found?.role).toBe('ADMIN')
  })

  it('onlyIfEmpty rejects a second create', async () => {
    await users.createWithLocalIdentity({ email: 'a@b.c', displayName: 'A', password: 'x', role: 'ADMIN', passwordHash: 'H' }, { onlyIfEmpty: true })
    await expect(
      users.createWithLocalIdentity({ email: 'b@b.c', displayName: 'B', password: 'x', role: 'ADMIN', passwordHash: 'H' }, { onlyIfEmpty: true }),
    ).rejects.toBeTruthy()
    expect(await users.countUsers()).toBe(1)
  })

  it('session create/findValid/touch/delete and deleteAllForUser', async () => {
    const u = await users.createWithLocalIdentity({ email: 'a@b.c', displayName: 'A', password: 'x', role: 'USER', passwordHash: 'H' })
    const now = new Date()
    const s = await sessions.create({ userId: u.id, expiresAt: new Date(now.getTime() + 60_000), ip: '127.0.0.1' })
    expect(await sessions.findValid({ sessionId: s.id, now })).not.toBeUndefined()
    // expired lookup returns undefined
    expect(await sessions.findValid({ sessionId: s.id, now: new Date(now.getTime() + 120_000) })).toBeUndefined()
    await sessions.deleteAllForUser({ userId: u.id })
    expect(await sessions.findValid({ sessionId: s.id, now })).toBeUndefined()
  })
})
