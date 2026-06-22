// SPDX-License-Identifier: AGPL-3.0-or-later
import * as User from '../../../user.domain'

export interface UserRow {
  id: string
  accountName: string
  displayName: string
  role: 'ADMIN' | 'USER'
  status: 'ACTIVE' | 'DISABLED'
  canManageExtensions: boolean
  canDownload: boolean
  allowNsfw: boolean
}

export function toDomain(row: UserRow): User.Model {
  return new User.Model({
    id: row.id,
    accountName: row.accountName,
    displayName: row.displayName,
    role: row.role,
    status: row.status,
    canManageExtensions: row.canManageExtensions,
    canDownload: row.canDownload,
    allowNsfw: row.allowNsfw,
  })
}
