// SPDX-License-Identifier: AGPL-3.0-or-later

import { ScryptPasswordHasher } from '../../password/infrastructure/security/scrypt/scrypt-password.hasher'
import { sessionRepository } from '../../sessions/application'
import { GetCurrentUserUseCase } from '../../sessions/application/usecases'
import { CreateUserUseCase, SetUserStatusUseCase, UpdateUserCapabilitiesUseCase, UpdateUserNameUseCase, UpdateUserNsfwPreferenceUseCase } from '../../users/application/usecases'
import { PrismaUserRepository } from '../../users/infrastructure/persistence/prisma/prisma-user.repository'
import { MemoryLoginRateLimiter } from '../infrastructure/persistence/memory/memory-login-rate-limiter'
import { ChangePasswordUseCase, LoginUseCase, LogoutUseCase, SetupFirstAdminUseCase } from './usecases'

const { auth } = useRuntimeConfig()

const passwordHasher = new ScryptPasswordHasher()

export const userRepository = new PrismaUserRepository(prisma)
export const loginRateLimiter = new MemoryLoginRateLimiter({
  maxAttempts: auth.loginRateLimitMaxAttempts,
  windowMs: auth.loginRateLimitWindowMs,
})

export const setupFirstAdmin = new SetupFirstAdminUseCase(userRepository, passwordHasher)
export const login = new LoginUseCase(userRepository, sessionRepository, passwordHasher, auth.sessionTtlMs)
export const logout = new LogoutUseCase(sessionRepository)
export const getCurrentUser = new GetCurrentUserUseCase(userRepository, sessionRepository, auth.sessionTtlMs, auth.sessionRefreshThresholdMs)
export const createUser = new CreateUserUseCase(userRepository, passwordHasher)
export const setUserStatus = new SetUserStatusUseCase(userRepository, sessionRepository)
export const updateUserName = new UpdateUserNameUseCase(userRepository)
export const updateNsfwPreference = new UpdateUserNsfwPreferenceUseCase(userRepository)
export const changePassword = new ChangePasswordUseCase(userRepository, sessionRepository, passwordHasher)
export const updateUserCapabilities = new UpdateUserCapabilitiesUseCase(userRepository)
