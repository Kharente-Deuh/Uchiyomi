// SPDX-License-Identifier: AGPL-3.0-or-later

import type { PrismaClient } from '../../../../../../prisma/generated/client'
import type { ExtensionsOverlayRepository, UpsertInstalledExtensionParams } from '../../../extension.domain'

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
}
