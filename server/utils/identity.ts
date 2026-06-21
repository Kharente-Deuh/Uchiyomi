import * as Login from '../domains/identity/auth/application/usecases/login.use-case'
import * as Logout from '../domains/identity/auth/application/usecases/logout.use-case'
import * as SetupFirstAdmin from '../domains/identity/auth/application/usecases/setup-first-admin.use-case'
// SPDX-License-Identifier: AGPL-3.0-or-later
import { MemoryLoginRateLimiter } from '../domains/identity/auth/infrastructure/persistence/memory/memory-login-rate-limiter'
import { ScryptPasswordHasher } from '../domains/identity/password/infrastructure/security/scrypt/scrypt-password.hasher'
import * as GetCurrentUser from '../domains/identity/sessions/application/usecases/get-current-user.use-case'
import { PrismaSessionRepository } from '../domains/identity/sessions/infrastructure/persistence/prisma/prisma-session.repository'
import * as CreateUser from '../domains/identity/users/application/usecases/create-user.use-case'
import * as SetUserStatus from '../domains/identity/users/application/usecases/set-user-status.use-case'
import { PrismaUserRepository } from '../domains/identity/users/infrastructure/persistence/prisma/prisma-user.repository'
import { prisma } from './prisma'

const { auth } = useRuntimeConfig()

const passwordHasher = new ScryptPasswordHasher()

export const userRepository = new PrismaUserRepository(prisma)
export const sessionRepository = new PrismaSessionRepository(prisma)

export const loginRateLimiter = new MemoryLoginRateLimiter({
  maxAttempts: auth.loginRateLimitMaxAttempts,
  windowMs: auth.loginRateLimitWindowMs,
})

export const setupFirstAdmin = new SetupFirstAdmin.UseCase(userRepository, passwordHasher)
export const login = new Login.UseCase(userRepository, sessionRepository, passwordHasher, auth.sessionTtlMs)
export const logout = new Logout.UseCase(sessionRepository)
export const getCurrentUser = new GetCurrentUser.UseCase(
  userRepository,
  sessionRepository,
  auth.sessionTtlMs,
  auth.sessionRefreshThresholdMs,
)
export const createUser = new CreateUser.UseCase(userRepository, passwordHasher)
export const setUserStatus = new SetUserStatus.UseCase(userRepository, sessionRepository)
