// SPDX-License-Identifier: AGPL-3.0-or-later
import process from 'node:process'
import { PrismaPg } from '@prisma/adapter-pg'
import { afterAll, beforeEach, describe, expect, it } from 'vitest'
import { PrismaClient } from '../../prisma/generated/client'
import { PrismaSourceRepository } from '../../server/domains/extensions/infrastructure/persistence/prisma/prisma-source.repository'

const connectionString = process.env.TEST_DATABASE_URL
const describeIf = connectionString ? describe : describe.skip

function src(id: string): { id: string, name: string, lang: string, isNsfw: boolean, isConfigurable: boolean } {
  return { id, name: `S${id}`, lang: 'en', isNsfw: false, isConfigurable: true }
}

describeIf('PrismaSourceRepository', () => {
  const prisma = new PrismaClient({ adapter: new PrismaPg({ connectionString }) })
  const repo = new PrismaSourceRepository(prisma)

  beforeEach(async () => {
    await prisma.source.deleteMany()
    await prisma.extension.deleteMany()
    await prisma.extension.create({ data: { pkgName: 'pkg', name: 'P', lang: 'en', isNsfw: false } })
  })
  afterAll(async () => {
    await prisma.source.deleteMany()
    await prisma.extension.deleteMany()
    await prisma.$disconnect()
  })

  it('syncForExtension upserts and listByPkg returns rows (default enabled)', async () => {
    await repo.syncForExtension('pkg', [src('1'), src('2')])
    const rows = await repo.listByPkg('pkg')
    expect(rows.map(r => [r.id, r.isEnabled])).toEqual([['1', true], ['2', true]])
  })

  it('setEnabled toggles and is preserved across re-sync', async () => {
    await repo.syncForExtension('pkg', [src('1')])
    await repo.setEnabled('1', false)
    await repo.syncForExtension('pkg', [src('1')]) // metadata refresh must not reset isEnabled
    const found = await repo.findById('1')
    expect(found?.isEnabled).toBe(false)
  })

  it('cascades on extension delete', async () => {
    await repo.syncForExtension('pkg', [src('1')])
    await prisma.extension.deleteMany({ where: { pkgName: 'pkg' } })
    expect(await repo.listByPkg('pkg')).toEqual([])
  })
})
