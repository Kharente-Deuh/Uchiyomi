// SPDX-License-Identifier: AGPL-3.0-or-later
import type { PrismaClient } from '../../../../../../prisma/generated/client'
import type { ExtensionErrorLogEntry, ExtensionHealthRow, ExtensionsOverlayRepository, RecordExtensionFailureParams, UpsertInstalledExtensionParams } from '../../../extension.domain'
import { healthToDomain } from './prisma-extension-repository.mapper'

export class PrismaExtensionRepository implements ExtensionsOverlayRepository {
  constructor(private readonly prisma: PrismaClient) {}

  async upsertInstalled(p: UpsertInstalledExtensionParams): Promise<void> {
    await this.prisma.extension.upsert({
      where: { pkgName: p.pkgName },
      create: {
        pkgName: p.pkgName,
        name: p.name,
        lang: p.lang,
        iconUrl: p.iconUrl,
        isNsfw: p.isNsfw,
        installedByUserId: p.installedByUserId,
      },
      update: { name: p.name, lang: p.lang, iconUrl: p.iconUrl, isNsfw: p.isNsfw },
    })
  }

  async deleteByPkgName(pkgName: string): Promise<void> {
    await this.prisma.extension.deleteMany({ where: { pkgName } })
  }

  async listHealthByPkgNames(pkgNames: string[]): Promise<ExtensionHealthRow[]> {
    if (pkgNames.length === 0) {
      return []
    }

    const rows = await this.prisma.extension.findMany({
      where: { pkgName: { in: pkgNames } },
      select: {
        pkgName: true,
        health: true,
        consecutiveFailures: true,
        lastErrorAt: true,
        lastErrorMessage: true,
        installedByUserId: true,
      },
    })

    return rows.map(r => healthToDomain(r))
  }

  async findHealth(pkgName: string): Promise<ExtensionHealthRow | undefined> {
    const row = await this.prisma.extension.findUnique({ where: { pkgName } })

    return row ? healthToDomain(row) : undefined
  }

  async recordFailure(p: RecordExtensionFailureParams): Promise<void> {
    await this.prisma.$transaction(async (tx) => {
      const ext = await tx.extension.update({
        where: { pkgName: p.pkgName },
        data: {
          health: 'ERROR',
          consecutiveFailures: { increment: 1 },
          lastErrorAt: new Date(),
          lastErrorMessage: p.message,
        },
      })
      await tx.extensionErrorLog.create({
        data: { extensionId: ext.id, message: p.message, context: p.context },
      })
    })
  }

  async recordSuccess(pkgName: string): Promise<void> {
    await this.prisma.extension.update({
      where: { pkgName },
      data: { health: 'OK', consecutiveFailures: 0 },
    })
  }

  async listErrorLog(pkgName: string): Promise<ExtensionErrorLogEntry[]> {
    const ext = await this.prisma.extension.findUnique({
      where: { pkgName },
      include: { errors: { orderBy: { occurredAt: 'desc' } } },
    })
    if (!ext) {
      return []
    }

    return ext.errors.map(e => ({ occurredAt: e.occurredAt, message: e.message, context: e.context ?? undefined }))
  }
}
