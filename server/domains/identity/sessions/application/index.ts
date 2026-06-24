// SPDX-License-Identifier: AGPL-3.0-or-later
import { PrismaSessionRepository } from '../infrastructure/persistence/prisma/prisma-session.repository'

export const sessionRepository = new PrismaSessionRepository(prisma)
