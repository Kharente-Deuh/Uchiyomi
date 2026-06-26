// SPDX-License-Identifier: AGPL-3.0-or-later
import type { PrismaClient } from '../../../../../../prisma/generated/client'
import type { ExtensionSource, ExtensionSourceRepository, FindManySourcesParams, StoredExtensionSource } from '../../../extension.domain'
import type { RawSource } from './prisma-source-repository.mapper'
import { sourceToDomain } from './prisma-source-repository.mapper'

export class PrismaSourceRepository implements ExtensionSourceRepository {
  constructor(private readonly prisma: PrismaClient) {}

  async syncForExtension(pkgName: string, sources: ExtensionSource[]): Promise<void> {
    // Upsert preserves isEnabled: it is only set on create, never on update.
    await this.prisma.$transaction(sources.map((s) => {
      const { id, ...source } = s

      return this.prisma.source.upsert({
        where: { id },
        create: { pkgName, ...source, id },
        update: { pkgName, ...source },
      })
    }))
  }

  async findMany(params: FindManySourcesParams): Promise<StoredExtensionSource[]> {
    const sources = await this.prisma.source.findMany({ where: params, orderBy: { lang: 'asc' } })

    const allIndex = sources.findIndex(s => s.lang === 'all')
    if (allIndex !== -1) {
      const allItem = sources[allIndex] as RawSource
      sources.splice(allIndex, 1)
      sources.unshift(allItem)
    }

    return sources.map(s => sourceToDomain(s))
  }

  async findById(id: string): Promise<StoredExtensionSource | undefined> {
    const source = await this.prisma.source.findFirst({ where: { id } })
    if (!source) {
      return
    }

    return sourceToDomain(source)
  }

  async update(id: string, data: Partial<Omit<StoredExtensionSource, 'id'>>): Promise<StoredExtensionSource> {
    const source = await this.prisma.source.update({ where: { id }, data })

    return sourceToDomain(source)
  }
}
