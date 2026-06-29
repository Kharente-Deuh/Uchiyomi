// SPDX-License-Identifier: AGPL-3.0-or-later

import type { IUseCase } from '../../../../../shared'
import type { PasswordHasher } from '../../../password/password.domain'
import type { SessionsRepository } from '../../../sessions/session.domain'
import type { UsersRepository } from '../../../users/user.domain'
import { AuthError } from '../../auth.domain'

export interface ChangePasswordUseCaseOpts {
  userId: string
  currentPassword: string
  newPassword: string
  logoutOtherDevices: boolean
  currentSessionId: string
}

export class ChangePasswordUseCase implements IUseCase<ChangePasswordUseCaseOpts, void> {
  constructor(
    private readonly userRepository: UsersRepository,
    private readonly sessionRepository: SessionsRepository,
    private readonly passwordHasher: PasswordHasher,
  ) {}

  async execute(opts: ChangePasswordUseCaseOpts): Promise<void> {
    const currentHash = await this.userRepository.findLocalPasswordHash({ userId: opts.userId })
    const ok = currentHash !== undefined && await this.passwordHasher.verify({ hash: currentHash, password: opts.currentPassword })
    if (!ok) {
      throw new AuthError('invalid_password', 'Current password is incorrect')
    }

    const passwordHash = await this.passwordHasher.hash({ password: opts.newPassword })
    await this.userRepository.updateLocalPasswordHash({ userId: opts.userId, passwordHash })

    if (opts.logoutOtherDevices) {
      await this.sessionRepository.deleteAllForUserExcept({ userId: opts.userId, exceptSessionId: opts.currentSessionId })
    }
  }
}
