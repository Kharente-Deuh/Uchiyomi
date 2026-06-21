// SPDX-License-Identifier: AGPL-3.0-or-later
import * as Session from '../../../session.domain'

export interface SessionRow {
  id: string
  userId: string
  expiresAt: Date
}

export function toDomain(row: SessionRow): Session.Model {
  return new Session.Model({
    id: row.id,
    userId: row.userId,
    expiresAt: row.expiresAt,
  })
}
