// SPDX-License-Identifier: AGPL-3.0-or-later
import process from 'node:process'
import { $fetch, fetch, setup } from '@nuxt/test-utils/e2e'
import { PrismaPg } from '@prisma/adapter-pg'
import { afterAll, beforeAll, describe, expect, it } from 'vitest'
import { PrismaClient } from '../../prisma/generated/client'

const connectionString = process.env.TEST_DATABASE_URL
const describeIf = connectionString ? describe : describe.skip

describeIf('auth spine e2e', async () => {
  const prisma = new PrismaClient({ adapter: new PrismaPg({ connectionString }) })
  await setup({
    server: true,
    env: {
      DATABASE_URL: connectionString,
      NUXT_SESSION_PASSWORD: 'e2e-test-session-password-32-characters',
    },
    nuxtConfig: {
      // Disable the PWA module during e2e tests — it tries to read its own
      // package.json via import.meta.url which resolves incorrectly under the
      // vitest node environment, crashing the Nuxt build step.
      pwa: false as never,
    },
  })

  beforeAll(async () => {
    await prisma.appUser.deleteMany()
  })
  afterAll(async () => {
    await prisma.appUser.deleteMany()
    await prisma.$disconnect()
  })

  it('setup → login → me → disable revokes', async () => {
    const setupStatus = await $fetch('/api/auth/setup')
    expect(setupStatus.required).toBe(true)
    expect(setupStatus.minPasswordLength).toBe(10)

    // First-run setup creates the admin and returns a session cookie.
    const setupRes = await fetch('/api/auth/setup', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ email: 'admin@x.io', displayName: 'Admin', password: 'longenough1' }),
    })
    const adminCookie = setupRes.headers.getSetCookie?.().join('; ') ?? setupRes.headers.get('set-cookie') ?? ''
    expect(adminCookie).toBeTruthy() // fail here, not on a downstream 401, if no cookie was set
    const setupStatusAfter = await $fetch('/api/auth/setup')
    expect(setupStatusAfter.required).toBe(false)

    // Admin creates a USER.
    await $fetch('/api/admin/users', {
      method: 'POST',
      headers: { cookie: adminCookie },
      body: { email: 'user@x.io', displayName: 'User', password: 'longenough2' },
    })

    // The new user logs in and reads /me.
    const loginRes = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ email: 'user@x.io', password: 'longenough2' }),
    })
    const userCookie = loginRes.headers.getSetCookie?.().join('; ') ?? loginRes.headers.get('set-cookie') ?? ''
    expect(userCookie).toBeTruthy() // fail here, not on a downstream 401, if no cookie was set
    const me = await $fetch('/api/auth/me', { headers: { cookie: userCookie } })
    expect(me.user.email).toBe('user@x.io')

    // A non-admin cannot create users (authorization 403).
    await expect($fetch('/api/admin/users', {
      method: 'POST',
      headers: { cookie: userCookie },
      body: { email: 'nope@x.io', displayName: 'No', password: 'longenough3' },
    })).rejects.toMatchObject({ statusCode: 403 })

    // Admin disables the user → the user's session is revoked immediately.
    const userRow = await prisma.appUser.findUnique({ where: { email: 'user@x.io' } })
    await $fetch(`/api/admin/users/${userRow!.id}/status`, {
      method: 'POST',
      headers: { cookie: adminCookie },
      body: { status: 'DISABLED' },
    })
    await expect($fetch('/api/auth/me', { headers: { cookie: userCookie } })).rejects.toMatchObject({ statusCode: 401 })
  })

  it('rejects bad credentials with a generic 401', async () => {
    await expect(
      $fetch('/api/auth/login', { method: 'POST', body: { email: 'admin@x.io', password: 'wrongpass11' } }),
    ).rejects.toMatchObject({ statusCode: 401 })
  })

  it('rejects a too-short password on setup with 400', async () => {
    // The password-length rule lives in the zod schema on the route — zod validation → 400, not 422.
    // We need to clear users first so setup is available (re-use a fresh state).
    await prisma.appUser.deleteMany()
    await expect(
      $fetch('/api/auth/setup', {
        method: 'POST',
        body: { email: 'admin@x.io', displayName: 'Admin', password: 'short' },
      }),
    ).rejects.toMatchObject({ statusCode: 400 })
  })
})
