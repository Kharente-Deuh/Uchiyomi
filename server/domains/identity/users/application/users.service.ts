// SPDX-License-Identifier: AGPL-3.0-or-later

import type { UserModel } from '../user.domain'
import type { CreateUserUseCaseOpts, SetUserStatusUseCaseOpts, UpdateUserCapabilitiesUseCaseOpts, UpdateUserNameUseCaseOpts, UpdateUserNsfwPreferenceUseCaseOpts } from './usecases'
import { userRepository } from '.'
import { ScryptPasswordHasher } from '../../password/infrastructure/security/scrypt/scrypt-password.hasher'
import { sessionRepository } from '../../sessions/application'
import { CreateUserUseCase, SetUserStatusUseCase, UpdateUserCapabilitiesUseCase, UpdateUserNameUseCase, UpdateUserNsfwPreferenceUseCase } from './usecases'

export interface UsersService {
  createUser: (opts: CreateUserUseCaseOpts) => Promise<Omit<UserModel, 'passwordHash'>>
  setUserStatus: (opts: SetUserStatusUseCaseOpts) => Promise<void>
  updateUserName: (opts: UpdateUserNameUseCaseOpts) => Promise<Omit<UserModel, 'passwordHash'>>
  updateNsfwPreference: (opts: UpdateUserNsfwPreferenceUseCaseOpts) => Promise<Omit<UserModel, 'passwordHash'>>
  updateUserCapabilities: (opts: UpdateUserCapabilitiesUseCaseOpts) => Promise<Omit<UserModel, 'passwordHash'>>
}

const passwordHasher = new ScryptPasswordHasher()

function createUser(opts: CreateUserUseCaseOpts): Promise<Omit<UserModel, 'passwordHash'>> {
  return new CreateUserUseCase(userRepository, passwordHasher).execute(opts)
}

function setUserStatus(opts: SetUserStatusUseCaseOpts): Promise<void> {
  return new SetUserStatusUseCase(userRepository, sessionRepository).execute(opts)
}

function updateUserName(opts: UpdateUserNameUseCaseOpts): Promise<Omit<UserModel, 'passwordHash'>> {
  return new UpdateUserNameUseCase(userRepository).execute(opts)
}

function updateNsfwPreference(opts: UpdateUserNsfwPreferenceUseCaseOpts): Promise<Omit<UserModel, 'passwordHash'>> {
  return new UpdateUserNsfwPreferenceUseCase(userRepository).execute(opts)
}

function updateUserCapabilities(opts: UpdateUserCapabilitiesUseCaseOpts): Promise<Omit<UserModel, 'passwordHash'>> {
  return new UpdateUserCapabilitiesUseCase(userRepository).execute(opts)
}

export function usersService(): UsersService {
  return {
    createUser,
    setUserStatus,
    updateUserName,
    updateNsfwPreference,
    updateUserCapabilities,
  }
}
