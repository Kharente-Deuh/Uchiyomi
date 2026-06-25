// SPDX-License-Identifier: AGPL-3.0-or-later
import type { PrismaClient } from '../../../../../../prisma/generated/client'
import type { ExtensionSource, ExtensionSourceRepository, StoredExtensionSource } from '../../../extension.domain'
import { sourceToDomain } from './prisma-source-repository.mapper'

export class PrismaSourceRepository implements ExtensionSourceRepository {
  constructor(private readonly prisma: PrismaClient) {}

  async syncForExtension(pkgName: string, sources: ExtensionSource[]): Promise<void> {
    // Upsert preserves isEnabled: it is only set on create, never on update.
    await this.prisma.$transaction(
      sources.map(s => this.prisma.source.upsert({
        where: { id: s.id },
        create: { id: s.id, pkgName, name: s.name, lang: s.lang, isNsfw: s.isNsfw, isConfigurable: s.isConfigurable },
        update: { pkgName, name: s.name, lang: s.lang, isNsfw: s.isNsfw, isConfigurable: s.isConfigurable },
      })),
    )
  }

  async listByPkg(pkgName: string): Promise<StoredExtensionSource[]> {
    const rows = await this.prisma.source.findMany({ where: { pkgName }, orderBy: { id: 'asc' } })

    return rows.map(r => sourceToDomain(r))
  }

  async findById(id: string): Promise<StoredExtensionSource | undefined> {
    const row = await this.prisma.source.findUnique({ where: { id } })

    return row ? sourceToDomain(row) : undefined
  }

  async setEnabled(id: string, isEnabled: boolean): Promise<StoredExtensionSource> {
    const row = await this.prisma.source.update({ where: { id }, data: { isEnabled } })

    return sourceToDomain(row)
  }
}
