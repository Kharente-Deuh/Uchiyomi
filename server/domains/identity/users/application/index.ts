// SPDX-License-Identifier: AGPL-3.0-or-later

import { PrismaUserRepository } from '../infrastructure/persistence/prisma/prisma-user.repository'

// Shared infrastructure singleton: the user repository is consumed by several
// domain services (auth, users, sessions) and a couple of routes directly. It
// lives here — dependency-free, importing only its Prisma adapter — so the
// service factories can share it without importing one another (which would
// create a users.service ↔ sessions.service cycle).
export const userRepository = new PrismaUserRepository(prisma)
