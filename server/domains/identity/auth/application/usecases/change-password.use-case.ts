// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../../shared/use-case'
import type * as Password from '../../../password/password.domain'
import type * as Session from '../../../sessions/session.domain'
import type * as User from '../../../users/user.domain'
import * as Auth from '../../auth.domain'

export interface Opts {
  userId: string
  currentPassword: string
  newPassword: string
  logoutOtherDevices: boolean
  currentSessionId: string
}

export class UseCase implements IUseCase<Opts, void> {
  constructor(
    private readonly userRepository: User.Repository,
    private readonly sessionRepository: Session.Repository,
    private readonly passwordHasher: Password.Hasher,
  ) {}

  async execute(opts: Opts): Promise<void> {
    const currentHash = await this.userRepository.findLocalPasswordHash({ userId: opts.userId })
    const ok = currentHash !== undefined
      && await this.passwordHasher.verify({ hash: currentHash, password: opts.currentPassword })
    if (!ok) {
      throw new Auth.AuthError('invalid_password', 'Current password is incorrect')
    }

    const passwordHash = await this.passwordHasher.hash({ password: opts.newPassword })
    await this.userRepository.updateLocalPasswordHash({ userId: opts.userId, passwordHash })

    if (opts.logoutOtherDevices) {
      await this.sessionRepository.deleteAllForUserExcept({ userId: opts.userId, exceptSessionId: opts.currentSessionId })
    }
  }
}
