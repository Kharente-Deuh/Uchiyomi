// SPDX-License-Identifier: AGPL-3.0-or-later
import type { PrismaClient } from '../../../../../../../prisma/generated/client'
import type * as Session from '../../../session.domain'
import { randomUUID } from 'node:crypto'
import { toDomain } from './prisma-session-repository.mapper'

export class PrismaSessionRepository implements Session.Repository {
  constructor(private readonly prisma: PrismaClient) {}

  async create(p: Session.CreateParams): Promise<Session.Model> {
    const row = await this.prisma.session.create({
      data: {
        id: randomUUID(),
        userId: p.userId,
        expiresAt: p.expiresAt,
        userAgent: p.userAgent ?? null,
        ip: p.ip ?? null,
      },
    })

    return toDomain(row)
  }

  async findValid(p: Session.FindValidParams): Promise<Session.Model | undefined> {
    const row = await this.prisma.session.findUnique({ where: { id: p.sessionId } })
    if (!row || row.expiresAt <= p.now) {
      return undefined
    }

    return toDomain(row)
  }

  async touch(p: Session.TouchParams): Promise<void> {
    await this.prisma.session.update({ where: { id: p.sessionId }, data: { expiresAt: p.expiresAt } })
  }

  async delete(p: Session.DeleteParams): Promise<void> {
    await this.prisma.session.deleteMany({ where: { id: p.sessionId } })
  }

  async deleteAllForUser(p: Session.DeleteAllForUserParams): Promise<void> {
    await this.prisma.session.deleteMany({ where: { userId: p.userId } })
  }

  async deleteAllForUserExcept(p: Session.DeleteAllForUserExceptParams): Promise<void> {
    await this.prisma.session.deleteMany({ where: { userId: p.userId, NOT: { id: p.exceptSessionId } } })
  }
}
