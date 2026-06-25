// SPDX-License-Identifier: AGPL-3.0-or-later
import type { SessionModel } from '../../sessions/session.domain'
import type { UserModel } from '../../users/user.domain'
import type { ChangePasswordUseCaseOpts, LoginUseCaseOpts, LogoutUseCaseOpts, SetupFirstAdminUseCaseOpts } from './usecases'
import { ScryptPasswordHasher } from '../../password/infrastructure/security/scrypt/scrypt-password.hasher'
import { sessionRepository } from '../../sessions/application'
import { userRepository } from '../../users/application'
import { MemoryLoginRateLimiter } from '../infrastructure/persistence/memory/memory-login-rate-limiter'
import { ChangePasswordUseCase, LoginUseCase, LogoutUseCase, SetupFirstAdminUseCase } from './usecases'

export interface AuthService {
  setupFirstAdmin: (opts: SetupFirstAdminUseCaseOpts) => Promise<Omit<UserModel, 'passwordHash'>>
  login: (opts: LoginUseCaseOpts) => Promise<SessionModel>
  logout: (opts: LogoutUseCaseOpts) => Promise<void>
  changePassword: (opts: ChangePasswordUseCaseOpts) => Promise<void>
}

const { auth } = useRuntimeConfig()

const passwordHasher = new ScryptPasswordHasher()

// Stateful in-memory rate limiter — must be a single instance per process, so it
// is wired here and consumed directly by the login route alongside the service.
export const loginRateLimiter = new MemoryLoginRateLimiter({
  maxAttempts: auth.loginRateLimitMaxAttempts,
  windowMs: auth.loginRateLimitWindowMs,
})

function setupFirstAdmin(opts: SetupFirstAdminUseCaseOpts): Promise<Omit<UserModel, 'passwordHash'>> {
  return new SetupFirstAdminUseCase(userRepository, passwordHasher).execute(opts)
}

function login(opts: LoginUseCaseOpts): Promise<SessionModel> {
  return new LoginUseCase(userRepository, sessionRepository, passwordHasher, auth.sessionTtlMs).execute(opts)
}

function logout(opts: LogoutUseCaseOpts): Promise<void> {
  return new LogoutUseCase(sessionRepository).execute(opts)
}

function changePassword(opts: ChangePasswordUseCaseOpts): Promise<void> {
  return new ChangePasswordUseCase(userRepository, sessionRepository, passwordHasher).execute(opts)
}

export function authService(): AuthService {
  return {
    setupFirstAdmin,
    login,
    logout,
    changePassword,
  }
}
