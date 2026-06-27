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

  it('pOST /api/extensions/foo { action: update } → 403 for non-admin user', async () => {
    const res = await fetch('/api/extensions/foo', {
      method: 'POST',
      headers: { 'content-type': 'application/json', 'cookie': nonAdminCookie },
      body: JSON.stringify({ action: 'update' }),
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
    const res = await $fetch<{ items: unknown[], total: number }>('/api/extensions', {
      headers: { cookie: adminCookie },
    })
    expect(Array.isArray(res.items)).toBe(true)
    expect(typeof res.total).toBe('number')
  })

  itWithSuwayomi('GET /api/extensions?isInstalled=false → admin-only filtered items (requires SUWAYOMI_URL)', async () => {
    const res = await $fetch<{ items: { isInstalled: boolean }[] }>('/api/extensions?isInstalled=false', { headers: { cookie: extManagerCookie } })
    expect(res.items.every(e => e.isInstalled === false)).toBe(true)
  })

  // Overlay-only route (no Suwayomi load): an unknown pkgName yields an empty
  // source list, not a 404 — so this runs without a live engine.
  it('gET /api/extensions/nonexistent/sources → 200 empty array', async () => {
    const res = await $fetch<{ sources: unknown[] }>('/api/extensions/com.nonexistent.pkg/sources', {
      headers: { cookie: adminCookie },
    })
    expect(res.sources).toEqual([])
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

  // Icon visibility — managers may preview the icon of a not-yet-installed
  // extension (adminBypassesInstallCheck), regular users may not. Discovers a
  // real not-installed, non-NSFW extension from the catalogue so it stays robust
  // across environments; skips its assertions if the engine has none.
  itWithSuwayomi('GET /api/extensions/:pkgName/icon → manager bypasses the install gate that blocks a regular user', async () => {
    const list = await $fetch<{ items: { pkgName: string, isInstalled: boolean, isNsfw: boolean }[] }>(
      '/api/extensions?isInstalled=false&pageSize=50',
      { headers: { cookie: adminCookie } },
    )
    const candidate = list.items.find(e => !e.isInstalled && !e.isNsfw)
    if (!candidate) {
      return // No not-installed extension available in this engine — nothing to assert.
    }

    // Regular user: not installed → 403.
    const asRegular = await fetch(`/api/extensions/${candidate.pkgName}/icon`, { headers: { cookie: nonAdminCookie } })
    expect(asRegular.status).toBe(403)

    // Manager: install gate bypassed → reaches the Suwayomi proxy (200, or 502 if
    // the upstream icon fetch fails), never 403.
    const asManager = await fetch(`/api/extensions/${candidate.pkgName}/icon`, { headers: { cookie: extManagerCookie } })
    expect(asManager.status).not.toBe(403)
  })

  // ----------------------------------------------------------------
  // Source enable/disable route — 403 gate (always-run, no Suwayomi needed).
  // POST /api/extensions/:pkgName/sources/:id/enable
  // ----------------------------------------------------------------

  it('pOST /api/extensions/:pkgName/sources/:id/enable → 403 for non-admin user', async () => {
    const res = await fetch('/api/extensions/any.pkg/sources/1/enable', {
      method: 'POST',
      headers: { 'content-type': 'application/json', 'cookie': nonAdminCookie },
      body: JSON.stringify({ isEnabled: false }),
    })
    expect(res.status).toBe(403)
  })

  it('pOST /api/extensions/:pkgName/sources/:id/enable → 400 for ext-manager with invalid body (missing isEnabled)', async () => {
    // Guards validate the body before any Suwayomi/DB call — safe to run always.
    const res = await fetch('/api/extensions/any.pkg/sources/1/enable', {
      method: 'POST',
      headers: { 'content-type': 'application/json', 'cookie': extManagerCookie },
      body: JSON.stringify({}),
    })
    expect(res.status).toBe(400)
  })

  // SUWAYOMI-gated: deeper enable.post guards require a Suwayomi-backed instance:
  //   - 404 when extension pkgName is unknown
  //   - 403 when extension is not installed or hasUpdate
  //   - 404 when sourceId is unknown for that extension
  // Add under itWithSuwayomi + itWithSourceId once those env vars are available.

  // ----------------------------------------------------------------
  // Extension sources/settings routes — 403/400 gate (always-run, no Suwayomi needed).
  // GET/PUT /api/extensions/:pkgName/sources/settings
  // ----------------------------------------------------------------

  it('gET /api/extensions/:pkgName/sources/settings → 403 for non-admin user', async () => {
    const res = await fetch('/api/extensions/any.pkg/sources/settings', { headers: { cookie: nonAdminCookie } })
    expect(res.status).toBe(403)
  })

  describe('pUT /api/extensions/:pkgName/sources/settings', () => {
    it('rejects a non-admin with 403', async () => {
      await expect($fetch('/api/extensions/any.pkg/sources/settings', {
        method: 'PUT',
        headers: { cookie: nonAdminCookie },
        body: { not: 'valid' },
      }))
        .rejects
        .toMatchObject({ response: { status: 403 } })
    })

    it('rejects an invalid PUT body with 400', async () => {
      await expect($fetch('/api/extensions/any.pkg/sources/settings', {
        method: 'PUT',
        headers: { cookie: extManagerCookie },
        body: { not: 'valid' },
      }))
        .rejects
        .toMatchObject({ response: { status: 400 } })
    })

    it('rejects a settings PUT where a list pref is missing entries with 400', async () => {
      // list/multiSelect prefs in the echo body must include entries and entryValues.
      await expect($fetch('/api/extensions/any.pkg/sources/settings', {
        method: 'PUT',
        headers: { cookie: extManagerCookie },
        body: {
          common: [{ position: 0, type: 'list', visible: true }], // missing entries/entryValues
          sources: [],
        },
      }))
        .rejects
        .toMatchObject({ response: { status: 400 } })
    })

    it('rejects a settings PUT where a switch pref is missing booleanDefault with 400', async () => {
      // switch/checkbox prefs in the echo body must include booleanDefault.
      await expect($fetch('/api/extensions/any.pkg/sources/settings', {
        method: 'PUT',
        headers: { cookie: extManagerCookie },
        body: {
          common: [{ position: 0, type: 'switch', visible: true }], // missing booleanDefault
          sources: [],
        },
      }))
        .rejects
        .toMatchObject({ response: { status: 400 } })
    })
  })
})
