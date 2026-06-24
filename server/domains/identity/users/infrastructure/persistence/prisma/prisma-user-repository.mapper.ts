// SPDX-License-Identifier: AGPL-3.0-or-later
import type { Role, UserStatus } from '~~/prisma/generated/enums'
import { UserModel } from '../../../user.domain'

export interface UserRow {
  id: string
  accountName: string
  displayName: string
  role: Role
  status: UserStatus
  canManageExtensions: boolean
  canDownload: boolean
  allowNsfw: boolean
  showNsfw: boolean
}

export function toDomain(row: UserRow): UserModel {
  return new UserModel({
    id: row.id,
    accountName: row.accountName,
    displayName: row.displayName,
    role: row.role,
    status: row.status,
    canManageExtensions: row.canManageExtensions,
    canDownload: row.canDownload,
    allowNsfw: row.allowNsfw,
    showNsfw: row.showNsfw,
  })
}
