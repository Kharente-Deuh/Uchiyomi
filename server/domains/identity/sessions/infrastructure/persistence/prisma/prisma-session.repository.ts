// SPDX-License-Identifier: AGPL-3.0-or-later
import type { PrismaClient } from '../../../../../../../prisma/generated/client'
import type { CreateSessionParams, DeleteAllSessionsForUserExceptParams, DeleteAllSessionsForUserParams, DeleteSessionParams, FindValidSessionParams, SessionModel, TouchSessionParams } from '../../../session.domain'
import { randomUUID } from 'node:crypto'
import { toDomain } from './prisma-session-repository.mapper'

export class PrismaSessionRepository implements PrismaSessionRepository {
  constructor(private readonly prisma: PrismaClient) {}

  async create(p: CreateSessionParams): Promise<SessionModel> {
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

  async findValid(p: FindValidSessionParams): Promise<SessionModel | undefined> {
    const row = await this.prisma.session.findUnique({ where: { id: p.sessionId } })
    if (!row || row.expiresAt <= p.now) {
      return
    }

    return toDomain(row)
  }

  async touch(p: TouchSessionParams): Promise<void> {
    await this.prisma.session.update({
      where: { id: p.sessionId },
      data: { expiresAt: p.expiresAt },
    })
  }

  async delete(p: DeleteSessionParams): Promise<void> {
    await this.prisma.session.deleteMany({ where: { id: p.sessionId } })
  }

  async deleteAllForUser(p: DeleteAllSessionsForUserParams): Promise<void> {
    await this.prisma.session.deleteMany({ where: { userId: p.userId } })
  }

  async deleteAllForUserExcept(p: DeleteAllSessionsForUserExceptParams): Promise<void> {
    await this.prisma.session.deleteMany({
      where: {
        userId: p.userId,
        NOT: { id: p.exceptSessionId },
      },
    })
  }
}
