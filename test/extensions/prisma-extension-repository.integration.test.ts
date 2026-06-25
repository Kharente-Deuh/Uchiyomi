// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment node
import process from 'node:process'
import { PrismaPg } from '@prisma/adapter-pg'
import { afterAll, beforeEach, describe, expect, it } from 'vitest'
import { PrismaClient } from '../../prisma/generated/client'
import { PrismaExtensionRepository } from '../../server/domains/extensions/infrastructure/persistence/prisma/prisma-extension.repository'

const connectionString = process.env.TEST_DATABASE_URL
const describeIf = connectionString ? describe : describe.skip

describeIf('PrismaExtensionRepository', () => {
  const prisma = new PrismaClient({ adapter: new PrismaPg({ connectionString }) })
  const repo = new PrismaExtensionRepository(prisma)

  async function makeUser(): Promise<string> {
    const u = await prisma.appUser.create({ data: { accountName: `u-${Date.now()}`, displayName: 'U' } })

    return u.id
  }

  beforeEach(async () => {
    await prisma.extensionErrorLog.deleteMany()
    await prisma.subscription.deleteMany()
    await prisma.series.deleteMany()
    await prisma.source.deleteMany()
    await prisma.extension.deleteMany()
    await prisma.appUser.deleteMany()
  })
  afterAll(async () => {
    await prisma.extensionErrorLog.deleteMany()
    await prisma.subscription.deleteMany()
    await prisma.series.deleteMany()
    await prisma.source.deleteMany()
    await prisma.extension.deleteMany()
    await prisma.appUser.deleteMany()
    await prisma.$disconnect()
  })

  it('upserts an installed extension and lists health', async () => {
    const userId = await makeUser()
    await repo.upsertInstalled({ pkgName: 'p1', name: 'P1', lang: 'en', isNsfw: false, installedByUserId: userId })
    const health = await repo.listHealthByPkgNames(['p1'])
    expect(health).toHaveLength(1)
    expect(health[0]).toMatchObject({ pkgName: 'p1', health: 'OK', consecutiveFailures: 0 })
  })

  it('records a failure then a success', async () => {
    const userId = await makeUser()
    await repo.upsertInstalled({ pkgName: 'p1', name: 'P1', lang: 'en', isNsfw: false, installedByUserId: userId })
    await repo.recordFailure({ pkgName: 'p1', message: 'boom', context: 'install' })
    let h = await repo.findHealth('p1')
    expect(h).toMatchObject({ health: 'ERROR', consecutiveFailures: 1, lastErrorMessage: 'boom' })
    expect(await repo.listErrorLog('p1')).toHaveLength(1)
    await repo.recordSuccess('p1')
    h = await repo.findHealth('p1')
    expect(h).toMatchObject({ health: 'OK', consecutiveFailures: 0 })
  })

  it('deletes by pkgName (cascades error log)', async () => {
    const userId = await makeUser()
    await repo.upsertInstalled({ pkgName: 'p1', name: 'P1', lang: 'en', isNsfw: false, installedByUserId: userId })
    await repo.recordFailure({ pkgName: 'p1', message: 'x' })
    await repo.deleteByPkgName('p1')
    expect(await repo.findHealth('p1')).toBeUndefined()
    expect(await repo.listErrorLog('p1')).toHaveLength(0)
  })

  it('deleteByPkgName drops source rows but preserves followed series and subscriptions', async () => {
    const userId = await makeUser()
    await prisma.extension.create({ data: { pkgName: 'pkg-dormancy', name: 'Dormancy Pkg', lang: 'en', isNsfw: false } })
    await prisma.source.create({ data: { id: 'src-dormancy-1', pkgName: 'pkg-dormancy', name: 'Dormancy Source', lang: 'en', isNsfw: false, isConfigurable: false } })
    const series = await prisma.series.create({
      data: { mangaId: 99_001, sourceId: 'src-dormancy-1', mangaUrl: '/m/99001', title: 'Dormancy Test Series', type: 'MANGA' },
    })
    await prisma.subscription.create({ data: { userId, seriesId: series.id } })

    await repo.deleteByPkgName('pkg-dormancy')

    expect(await prisma.source.findMany({ where: { pkgName: 'pkg-dormancy' } })).toEqual([])
    expect(await prisma.series.findUnique({ where: { id: series.id } })).not.toBeNull()
    expect(await prisma.subscription.count({ where: { seriesId: series.id } })).toBe(1)
  })
})
