// SPDX-License-Identifier: AGPL-3.0-or-later

import type { UserModel } from '../../users/user.domain'
import type { GetCurrentUserUseCaseOpts } from './usecases'
import { sessionRepository } from '.'
import { userRepository } from '../../users/application'
import { GetCurrentUserUseCase } from './usecases'

export interface SessionsService {
  getCurrentUser: (opts: GetCurrentUserUseCaseOpts) => Promise<UserModel>
}

const { auth } = useRuntimeConfig()

function getCurrentUser(opts: GetCurrentUserUseCaseOpts): Promise<UserModel> {
  return new GetCurrentUserUseCase(
    userRepository,
    sessionRepository,
    auth.sessionTtlMs,
    auth.sessionRefreshThresholdMs,
  ).execute(opts)
}

export function sessionsService(): SessionsService {
  return {
    getCurrentUser,
  }
}
