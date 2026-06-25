// SPDX-License-Identifier: AGPL-3.0-or-later

import { SessionModel } from '../../../session.domain'

export interface SessionRow {
  id: string
  userId: string
  expiresAt: Date
}

export function toDomain(row: SessionRow): SessionModel {
  return new SessionModel({
    id: row.id,
    userId: row.userId,
    expiresAt: row.expiresAt,
  })
}
