// SPDX-License-Identifier: AGPL-3.0-or-later
import process from 'node:process'
import { $fetch, fetch, setup } from '@nuxt/test-utils/e2e'
import { PrismaPg } from '@prisma/adapter-pg'
import { afterAll, beforeAll, describe, expect, it } from 'vitest'
import { PrismaClient } from '../../prisma/generated/client'

const connectionString = process.env.TEST_DATABASE_URL
const describeIf = connectionString ? describe : describe.skip

describeIf('extensions routes e2e', async () => {
  const prisma = new PrismaClient({ adapter: new PrismaPg({ connectionString }) })
  await setup({
    server: true,
    env: {
      DATABASE_URL: connectionString,
      NUXT_SESSION_PASSWORD: 'e2e-test-session-password-32-characters',
    },
    nuxtConfig: {
      pwa: false as never,
    },
  })

  // adminCookie: the setup admin (canManageUsers AND canManageExtensions — bootstrap grants full caps).
  // extManagerCookie: a user explicitly created with canManageExtensions: true.
  // nonAdminCookie: a plain user with no special capabilities.
  let adminCookie: string
  let extManagerCookie: string
  let nonAdminCookie: string

  beforeAll(async () => {
    await prisma.appUser.deleteMany()
    // Create the first admin via setup endpoint.
    const setupRes = await fetch('/api/auth/setup', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ accountName: 'admin', displayName: 'Admin', password: 'longenough1' }),
    })
    adminCookie = setupRes.headers.getSetCookie?.().join('; ') ?? setupRes.headers.get('set-cookie') ?? ''

    // Create a user with canManageExtensions: true (for admin extension routes).
    await fetch('/api/admin/users', {
      method: 'POST',
      headers: { 'content-type': 'application/json', 'cookie': adminCookie },
      body: JSON.stringify({ accountName: 'extmanager', displayName: 'Ext Manager', password: 'longenough1', canManageExtensions: true }),
    })
    const extLoginRes = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ accountName: 'extmanager', password: 'longenough1' }),
    })
    extManagerCookie = extLoginRes.headers.getSetCookie?.().join('; ') ?? extLoginRes.headers.get('set-cookie') ?? ''

    // Create a regular user (no canManageExtensions) via the admin endpoint.
    await fetch('/api/admin/users', {
      method: 'POST',
      headers: { 'content-type': 'application/json', 'cookie': adminCookie },
      body: JSON.stringify({ accountName: 'regular', displayName: 'Regular', password: 'longenough1' }),
    })
    // Log in as the regular user to obtain a session cookie.
    const loginRes = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ accountName: 'regular', password: 'longenough1' }),
    })
    nonAdminCookie = loginRes.headers.getSetCookie?.().join('; ') ?? loginRes.headers.get('set-cookie') ?? ''
  })
  afterAll(async () => {
    await prisma.appUser.deleteMany()
    await prisma.$disconnect()
  })

  // ----------------------------------------------------------------
  // Auth-gating assertions — safe regardless of Suwayomi availability.
  // These return 401 before any Suwayomi call is made.
  // ----------------------------------------------------------------

  it('gET /api/extensions → 401 when unauthenticated', async () => {
    const res = await fetch('/api/extensions')
    expect(res.status).toBe(401)
  })

  it('gET /api/extensions?pageSize=abc → 400 (invalid query)', async () => {
    const res = await fetch('/api/extensions?pageSize=abc', { headers: { cookie: adminCookie } })
    expect(res.status).toBe(400)
  })

  it('gET /api/extensions/whatever/sources → 401 when unauthenticated', async () => {
    const res = await fetch('/api/extensions/whatever/sources')
    expect(res.status).toBe(401)
  })

  it('gET /api/extensions/whatever → 401 when unauthenticated', async () => {
    const res = await fetch('/api/extensions/whatever')
    expect(res.status).toBe(401)
  })

  it('gET /api/extensions/whatever/icon → 401 when unauthenticated', async () => {
    const res = await fetch('/api/extensions/whatever/icon')
    expect(res.status).toBe(401)
  })

  // ----------------------------------------------------------------
  // Extension management — 403 gate (always-run, no Suwayomi needed).
  // Non-admin user → 403 before any use-case / Suwayomi call.
  // ----------------------------------------------------------------

  it('pOST /api/extensions/foo → 403 for non-admin user', async () => {
    const res = await fetch('/api/extensions/foo', {
      method: 'POST',
      headers: { 'content-type': 'application/json', 'cookie': nonAdminCookie },
      body: JSON.stringify({ action: 'install' }),
    })
    expect(res.status).toBe(403)
  })

  // ----------------------------------------------------------------
  // Authenticated happy-path assertions — gated on SUWAYOMI_URL.
  // When SUWAYOMI_URL is unset, these tests are individually skipped
  // via it.skipIf so the suite still reports as passed.
  // ----------------------------------------------------------------

  const suwayomiBase = process.env.SUWAYOMI_URL
  const itWithSuwayomi = suwayomiBase ? it : it.skip

  itWithSuwayomi('GET /api/extensions → 200 with paginated shape (requires SUWAYOMI_URL)', async () => {
    const res = await $fetch<{ items: unknown[], page: number, pageSize: number, totalCount: number }>('/api/extensions', {
      headers: { cookie: adminCookie },
    })
    expect(Array.isArray(res.items)).toBe(true)
    expect(typeof res.totalCount).toBe('number')
    expect(res.page).toBe(1)
  })

  itWithSuwayomi('GET /api/extensions?isInstalled=false → admin-only filtered items (requires SUWAYOMI_URL)', async () => {
    const res = await $fetch<{ items: { isInstalled: boolean }[] }>('/api/extensions?isInstalled=false', { headers: { cookie: extManagerCookie } })
    expect(res.items.every(e => e.isInstalled === false)).toBe(true)
  })

  itWithSuwayomi('GET /api/extensions/nonexistent/sources → 200 empty array (requires SUWAYOMI_URL)', async () => {
    const res = await $fetch<{ sources: unknown[] }>('/api/extensions/com.nonexistent.pkg/sources', {
      headers: { cookie: adminCookie },
    })
    expect(Array.isArray(res.sources)).toBe(true)
  })

  // Install happy-path — mutates Suwayomi state, gated on SUWAYOMI_URL.
  // Uses a non-existent pkgName to confirm auth passes (non-403) without
  // actually installing anything. A real install would require a valid pkgName
  // and would mutate Suwayomi's extension list.
  itWithSuwayomi('POST /api/extensions/nonexistent-pkg → non-403 for ext-manager (requires SUWAYOMI_URL)', async () => {
    const res = await fetch('/api/extensions/nonexistent-pkg', {
      method: 'POST',
      headers: { 'content-type': 'application/json', 'cookie': extManagerCookie },
      body: JSON.stringify({ action: 'install' }),
    })
    // Auth gate passed (would be 403 for non-canManageExtensions user).
    // Route may 4xx/5xx from Suwayomi returning an error for an unknown pkgName — that is expected.
    expect(res.status).not.toBe(403)
  })

  // ----------------------------------------------------------------
  // Source preferences routes — 403 gate (always-run, no Suwayomi needed).
  // ----------------------------------------------------------------

  it('gET /api/sources/1/preferences → 403 for non-admin user', async () => {
    const res = await fetch('/api/sources/1/preferences', { headers: { cookie: nonAdminCookie } })
    expect(res.status).toBe(403)
  })

  it('pATCH /api/sources/1 → 403 for non-admin user', async () => {
    const res = await fetch('/api/sources/1', {
      method: 'PATCH',
      headers: { 'content-type': 'application/json', 'cookie': nonAdminCookie },
      body: JSON.stringify({ isEnabled: false }),
    })
    expect(res.status).toBe(403)
  })

  it('pUT /api/sources/1/preferences → 403 for non-admin user', async () => {
    const res = await fetch('/api/sources/1/preferences', {
      method: 'PUT',
      headers: { 'content-type': 'application/json', 'cookie': nonAdminCookie },
      body: JSON.stringify({ position: 0 }),
    })
    expect(res.status).toBe(403)
  })

  // Source preferences happy-path — gated on SUWAYOMI_URL.
  // Uses a configurable source id from env (TEST_SOURCE_ID) or skips if not available.
  const testSourceId = process.env.TEST_SOURCE_ID
  const itWithSourceId = suwayomiBase && testSourceId ? it : it.skip

  itWithSourceId('GET /api/sources/:id/preferences → 200 with preferences array (requires SUWAYOMI_URL + TEST_SOURCE_ID)', async () => {
    const res = await $fetch<{ preferences: unknown[] }>(`/api/sources/${testSourceId}/preferences`, { headers: { cookie: extManagerCookie } })
    expect(Array.isArray(res.preferences)).toBe(true)
  })
})
