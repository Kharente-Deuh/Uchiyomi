// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ExtensionHealth as PrismaExtensionHealth } from '~~/prisma/generated/enums'
import type { ExtensionHealthRow } from '../../../extension.domain'

export interface RawExtension {
  pkgName: string
  health: PrismaExtensionHealth
  consecutiveFailures: number
  lastErrorAt: Date | null
  lastErrorMessage: string | null
  installedByUserId: string | null
}

export function healthToDomain(row: RawExtension): ExtensionHealthRow {
  return {
    pkgName: row.pkgName,
    health: row.health,
    consecutiveFailures: row.consecutiveFailures,
    lastErrorAt: row.lastErrorAt ?? undefined,
    lastErrorMessage: row.lastErrorMessage ?? undefined,
    installedByUserId: row.installedByUserId ?? undefined,
  }
}
