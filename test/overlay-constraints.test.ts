// SPDX-License-Identifier: AGPL-3.0-or-later

// @vitest-environment node
import type { Series } from '../prisma/generated/client'
import process from 'node:process'
import { PrismaPg } from '@prisma/adapter-pg'
import { afterAll, beforeAll, describe, expect, it } from 'vitest'
import { PrismaClient } from '../prisma/generated/client'

const connectionString = process.env.TEST_DATABASE_URL
const prisma = new PrismaClient({ adapter: new PrismaPg({ connectionString }) })

async function makeSeries(suffix: string): Promise<Series> {
  return prisma.series.create({
    data: {
      mangaId: Number.parseInt(suffix, 10),
      sourceId: `src-${suffix}`,
      mangaUrl: `/m/${suffix}`,
      title: `S ${suffix}`,
      type: 'WEBTOON',
    },
  })
}

beforeAll(async () => {
  // Clean slate; order respects FKs (cascades handle the rest). `extension` is
  // standalone (no cascade from user) so it is purged explicitly to stay idempotent.
  await prisma.appUser.deleteMany()
  await prisma.series.deleteMany()
  await prisma.extension.deleteMany()
})

afterAll(async () => {
  await prisma.appUser.deleteMany()
  await prisma.series.deleteMany()
  await prisma.extension.deleteMany()
  await prisma.$disconnect()
})

// Skipped (not failed) when no test database is configured, so `pnpm test`
// stays green for contributors without a local Postgres.
describe.skipIf(!connectionString)('overlay constraints', () => {
  it('enforces one Series per (mangaId, sourceId)', async () => {
    const a = await makeSeries('1001')
    await expect(
      prisma.series.create({ data: { mangaId: a.mangaId, sourceId: a.sourceId, mangaUrl: '/dup', title: 'dup', type: 'MANGA' } }),
    ).rejects.toThrow()
  })

  it('enforces one DownloadJob per chapter', async () => {
    const series = await makeSeries('1002')
    const chapter = await prisma.chapter.create({
      data: { chapterId: 1, seriesId: series.id, number: 1, name: 'c1', uploadedAt: new Date() },
    })
    await prisma.downloadJob.create({ data: { chapterId: chapter.id, seriesId: series.id } })
    await expect(
      prisma.downloadJob.create({ data: { chapterId: chapter.id, seriesId: series.id } }),
    ).rejects.toThrow()
  })

  it('enforces one Subscription per (user, series) and cascades on user delete', async () => {
    const series = await makeSeries('1003')
    const user = await prisma.appUser.create({ data: { accountName: 'user1003', displayName: 'U' } })
    await prisma.subscription.create({ data: { userId: user.id, seriesId: series.id } })
    await expect(
      prisma.subscription.create({ data: { userId: user.id, seriesId: series.id } }),
    ).rejects.toThrow()

    await prisma.appUser.delete({ where: { id: user.id } })
    expect(await prisma.subscription.count({ where: { userId: user.id } })).toBe(0)
  })

  it('cascades reading progress on user delete', async () => {
    const series = await makeSeries('1004')
    const chapter = await prisma.chapter.create({
      data: { chapterId: 1, seriesId: series.id, number: 1, name: 'c1', uploadedAt: new Date() },
    })
    const user = await prisma.appUser.create({ data: { accountName: 'user1004', displayName: 'U' } })
    await prisma.readingProgress.create({ data: { userId: user.id, chapterId: chapter.id } })

    await prisma.appUser.delete({ where: { id: user.id } })

    expect(await prisma.readingProgress.count({ where: { userId: user.id } })).toBe(0)
  })
})
