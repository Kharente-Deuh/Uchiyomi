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

  it('setup → login (case-insensitive) → me → duplicate 409 → disable revokes', async () => {
    const setupStatus = await $fetch('/api/auth/setup')
    expect(setupStatus.required).toBe(true)
    expect(setupStatus.minPasswordLength).toBe(10)

    const setupRes = await fetch('/api/auth/setup', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ accountName: 'admin', displayName: 'Admin', password: 'longenough1' }),
    })
    const adminCookie = setupRes.headers.getSetCookie?.().join('; ') ?? setupRes.headers.get('set-cookie') ?? ''
    expect(adminCookie).toBeTruthy()
    const setupStatusAfter = await $fetch('/api/auth/setup')
    expect(setupStatusAfter.required).toBe(false)

    // Admin creates a USER.
    await $fetch('/api/admin/users', {
      method: 'POST',
      headers: { cookie: adminCookie },
      body: { accountName: 'user', displayName: 'User', password: 'longenough2' },
    })

    // Duplicate account name (case-insensitive) → 409.
    await expect($fetch('/api/admin/users', {
      method: 'POST',
      headers: { cookie: adminCookie },
      body: { accountName: 'USER', displayName: 'User2', password: 'longenough9' },
    })).rejects.toMatchObject({ statusCode: 409 })

    // Login is case-insensitive: `USER` logs into the `user` account.
    const loginRes = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ accountName: 'USER', password: 'longenough2' }),
    })
    const userCookie = loginRes.headers.getSetCookie?.().join('; ') ?? loginRes.headers.get('set-cookie') ?? ''
    expect(userCookie).toBeTruthy()
    const me = await $fetch('/api/auth/me', { headers: { cookie: userCookie } })
    expect(me.user.accountName).toBe('user')

    // A non-admin cannot create users (authorization 403).
    await expect($fetch('/api/admin/users', {
      method: 'POST',
      headers: { cookie: userCookie },
      body: { accountName: 'nope', displayName: 'No', password: 'longenough3' },
    })).rejects.toMatchObject({ statusCode: 403 })

    // Admin disables the user → the user's session is revoked immediately.
    const userRow = await prisma.appUser.findUnique({ where: { accountName: 'user' } })
    await $fetch(`/api/admin/users/${userRow!.id}/status`, {
      method: 'POST',
      headers: { cookie: adminCookie },
      body: { status: 'DISABLED' },
    })
    await expect($fetch('/api/auth/me', { headers: { cookie: userCookie } })).rejects.toMatchObject({ statusCode: 401 })
  })

  it('rejects bad credentials with a generic 401', async () => {
    await expect(
      $fetch('/api/auth/login', { method: 'POST', body: { accountName: 'admin', password: 'wrongpass11' } }),
    ).rejects.toMatchObject({ statusCode: 401 })
  })

  it('rejects a too-short password on setup with 400', async () => {
    await prisma.appUser.deleteMany()
    await expect(
      $fetch('/api/auth/setup', {
        method: 'POST',
        body: { accountName: 'admin', displayName: 'Admin', password: 'short' },
      }),
    ).rejects.toMatchObject({ statusCode: 400 })
  })

  it('self-service: PATCH /me changes displayName and /me reflects it', async () => {
    await prisma.appUser.deleteMany()
    const setupRes = await fetch('/api/auth/setup', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ accountName: 'admin', displayName: 'Admin', password: 'longenough1' }),
    })
    const cookie = setupRes.headers.getSetCookie?.().join('; ') ?? setupRes.headers.get('set-cookie') ?? ''
    const patched = await $fetch('/api/auth/me', { method: 'PATCH', headers: { cookie }, body: { displayName: 'Renamed' } })
    expect(patched.user.displayName).toBe('Renamed')
    const me = await $fetch('/api/auth/me', { headers: { cookie } })
    expect(me.user.displayName).toBe('Renamed')
  })

  it('self-service: change password — new works, old fails; logoutOtherDevices keeps current and revokes others', async () => {
    await prisma.appUser.deleteMany()
    await fetch('/api/auth/setup', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ accountName: 'admin', displayName: 'Admin', password: 'oldpass1234' }),
    })
    // Two independent sessions for the same account (two logins).
    const login = async (password: string): Promise<string> => {
      const r = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ accountName: 'admin', password }),
      })

      return r.headers.getSetCookie?.().join('; ') ?? r.headers.get('set-cookie') ?? ''
    }

    const sessionA = await login('oldpass1234')
    const sessionB = await login('oldpass1234')

    // Change the password from session A, revoking other devices.
    await $fetch('/api/auth/me/password', {
      method: 'POST',
      headers: { cookie: sessionA },
      body: { currentPassword: 'oldpass1234', newPassword: 'newpass1234', logoutOtherDevices: true },
    })

    // Session A (current) still valid; session B revoked.
    const meA = await $fetch('/api/auth/me', { headers: { cookie: sessionA } })
    expect(meA.user.accountName).toBe('admin')
    await expect($fetch('/api/auth/me', { headers: { cookie: sessionB } })).rejects.toMatchObject({ statusCode: 401 })

    // New password logs in; old password is rejected.
    await expect(login('oldpass1234').then(c => $fetch('/api/auth/me', { headers: { cookie: c } })))
      .rejects
      .toMatchObject({ statusCode: 401 })
    const fresh = await login('newpass1234')
    const meFresh = await $fetch('/api/auth/me', { headers: { cookie: fresh } })
    expect(meFresh.user.accountName).toBe('admin')
  })

  it('self-service: wrong current password → 400', async () => {
    await prisma.appUser.deleteMany()
    const setupRes = await fetch('/api/auth/setup', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ accountName: 'admin', displayName: 'Admin', password: 'longenough1' }),
    })
    const cookie = setupRes.headers.getSetCookie?.().join('; ') ?? setupRes.headers.get('set-cookie') ?? ''
    await expect($fetch('/api/auth/me/password', {
      method: 'POST',
      headers: { cookie },
      body: { currentPassword: 'wrongpass99', newPassword: 'newpass1234' },
    })).rejects.toMatchObject({ statusCode: 400 })
  })

  it('self-service: PATCH /me with { showNsfw: true } updates the preference; {} → 400', async () => {
    await prisma.appUser.deleteMany()
    const setupRes = await fetch('/api/auth/setup', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ accountName: 'admin', displayName: 'Admin', password: 'longenough1' }),
    })
    const cookie = setupRes.headers.getSetCookie?.().join('; ') ?? setupRes.headers.get('set-cookie') ?? ''

    // { showNsfw: true } → 200 and user.showNsfw === true
    const patched = await $fetch('/api/auth/me', { method: 'PATCH', headers: { cookie }, body: { showNsfw: true } })
    expect(patched.user.showNsfw).toBe(true)

    // {} → 400 (at least one field required)
    await expect(
      $fetch('/api/auth/me', { method: 'PATCH', headers: { cookie }, body: {} }),
    ).rejects.toMatchObject({ statusCode: 400 })
  })

  it('admin: PATCH /users/[id] by admin works; by non-admin 403; unknown id 404', async () => {
    await prisma.appUser.deleteMany()
    const setupRes = await fetch('/api/auth/setup', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ accountName: 'admin', displayName: 'Admin', password: 'longenough1' }),
    })
    const adminCookie = setupRes.headers.getSetCookie?.().join('; ') ?? setupRes.headers.get('set-cookie') ?? ''
    const created = await $fetch('/api/admin/users', {
      method: 'POST',
      headers: { cookie: adminCookie },
      body: { accountName: 'user', displayName: 'User', password: 'longenough2' },
    })
    const userId = created.user.id

    // Admin renames the user.
    const patched = await $fetch(`/api/admin/users/${userId}`, { method: 'PATCH', headers: { cookie: adminCookie }, body: { displayName: 'Renamed' } })
    expect(patched.user.displayName).toBe('Renamed')

    // Non-admin is forbidden.
    const loginRes = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ accountName: 'user', password: 'longenough2' }),
    })
    const userCookie = loginRes.headers.getSetCookie?.().join('; ') ?? loginRes.headers.get('set-cookie') ?? ''
    await expect($fetch(`/api/admin/users/${userId}`, { method: 'PATCH', headers: { cookie: userCookie }, body: { displayName: 'Nope' } }))
      .rejects
      .toMatchObject({ statusCode: 403 })

    // Unknown id → 404.
    await expect($fetch('/api/admin/users/does-not-exist', { method: 'PATCH', headers: { cookie: adminCookie }, body: { displayName: 'X' } }))
      .rejects
      .toMatchObject({ statusCode: 404 })
  })

  it('admin: PATCH /users/[id] with { allowNsfw: true } → 200 & user.allowNsfw true; non-admin → 403', async () => {
    await prisma.appUser.deleteMany()
    const setupRes = await fetch('/api/auth/setup', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ accountName: 'admin', displayName: 'Admin', password: 'longenough1' }),
    })
    const adminCookie = setupRes.headers.getSetCookie?.().join('; ') ?? setupRes.headers.get('set-cookie') ?? ''
    const created = await $fetch('/api/admin/users', {
      method: 'POST',
      headers: { cookie: adminCookie },
      body: { accountName: 'user', displayName: 'User', password: 'longenough2' },
    })
    const userId = created.user.id

    // Admin grants allowNsfw.
    const patched = await $fetch(`/api/admin/users/${userId}`, {
      method: 'PATCH',
      headers: { cookie: adminCookie },
      body: { allowNsfw: true },
    })
    expect(patched.user.allowNsfw).toBe(true)

    // Non-admin is forbidden.
    const loginRes = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ accountName: 'user', password: 'longenough2' }),
    })
    const userCookie = loginRes.headers.getSetCookie?.().join('; ') ?? loginRes.headers.get('set-cookie') ?? ''
    await expect($fetch(`/api/admin/users/${userId}`, {
      method: 'PATCH',
      headers: { cookie: userCookie },
      body: { allowNsfw: false },
    })).rejects.toMatchObject({ statusCode: 403 })
  })
})
