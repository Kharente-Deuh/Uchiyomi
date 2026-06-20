// SPDX-License-Identifier: AGPL-3.0-or-later
import process from 'node:process'
import { PrismaPg } from '@prisma/adapter-pg'
import { PrismaClient } from '../../prisma/generated/client'

const globalForPrisma = globalThis as unknown as { prisma?: PrismaClient }

function createClient(): PrismaClient {
  // Prisma 7 requires a driver adapter for a direct connection.
  const adapter = new PrismaPg({ connectionString: process.env.DATABASE_URL })
  return new PrismaClient({ adapter })
}

export const prisma: PrismaClient = globalForPrisma.prisma ?? createClient()

if (import.meta.dev)
  globalForPrisma.prisma = prisma
