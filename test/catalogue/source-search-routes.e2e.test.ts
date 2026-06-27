// SPDX-License-Identifier: AGPL-3.0-or-later
import process from 'node:process'
import { fetch, setup } from '@nuxt/test-utils/e2e'
import { PrismaPg } from '@prisma/adapter-pg'
import { afterAll, beforeAll, describe, expect, it } from 'vitest'
import { PrismaClient } from '../../prisma/generated/client'

const connectionString = process.env.TEST_DATABASE_URL
const describeIf = connectionString ? describe : describe.skip

describeIf('source search routes e2e', async () => {
  const prisma = new PrismaClient({ adapter: new PrismaPg({ connectionString }) })
  await setup({
    server: true,
    env: {
      DATABASE_URL: connectionString,
      NUXT_SESSION_PASSWORD: 'e2e-test-session-password-32-characters',
    },
    nuxtConfig: { pwa: false as never },
  })

  let adminCookie: string

  beforeAll(async () => {
    await prisma.appUser.deleteMany()
    const setupRes = await fetch('/api/auth/setup', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ accountName: 'admin', displayName: 'Admin', password: 'longenough1' }),
    })
    adminCookie = setupRes.headers.getSetCookie?.().join('; ') ?? setupRes.headers.get('set-cookie') ?? ''
  })
  afterAll(async () => {
    await prisma.appUser.deleteMany()
    await prisma.$disconnect()
  })

  it('gET search → 401 when unauthenticated', async () => {
    const res = await fetch('/api/extensions/pkg/sources/123/search?q=a')
    expect(res.status).toBe(401)
  })

  it('gET search → 404 for an unknown source', async () => {
    const res = await fetch('/api/extensions/pkg/sources/does-not-exist/search?q=a', { headers: { cookie: adminCookie } })
    expect(res.status).toBe(404)
  })

  it('gET search → 400 for an invalid page', async () => {
    const res = await fetch('/api/extensions/pkg/sources/123/search?page=abc', { headers: { cookie: adminCookie } })
    expect(res.status).toBe(400)
  })

  it('gET search → 400 for an invalid type', async () => {
    const res = await fetch('/api/extensions/pkg/sources/123/search?type=bogus', { headers: { cookie: adminCookie } })
    expect(res.status).toBe(400)
  })

  it('gET search → 400 for type=latest on a source that does not support it', async () => {
    await prisma.source.deleteMany({ where: { pkgName: 'pkg.latest' } })
    await prisma.extension.deleteMany({ where: { pkgName: 'pkg.latest' } })
    await prisma.extension.create({ data: { pkgName: 'pkg.latest', name: 'L', lang: 'en', isNsfw: false } })
    await prisma.source.create({ data: { id: '9001', pkgName: 'pkg.latest', name: 'S', lang: 'en', isNsfw: false, isConfigurable: false, isEnabled: true, supportsLatest: false } })
    try {
      const res = await fetch('/api/extensions/pkg.latest/sources/9001/search?type=latest', { headers: { cookie: adminCookie } })
      expect(res.status).toBe(400)
    } finally {
      await prisma.source.deleteMany({ where: { pkgName: 'pkg.latest' } })
      await prisma.extension.deleteMany({ where: { pkgName: 'pkg.latest' } })
    }
  })

  it('gET thumbnail → 401 when unauthenticated', async () => {
    const res = await fetch('/api/manga/123/thumbnail')
    expect(res.status).toBe(401)
  })

  it('gET thumbnail → 400 for a non-numeric id', async () => {
    const res = await fetch('/api/manga/abc/thumbnail', { headers: { cookie: adminCookie } })
    expect(res.status).toBe(400)
  })
})
