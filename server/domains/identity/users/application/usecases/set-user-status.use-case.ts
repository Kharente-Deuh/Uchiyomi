// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../../shared/use-case'
import type { SessionsRepository } from '../../../sessions/session.domain'
import type { UsersRepository, UserStatus } from '../../user.domain'

export interface SetUserStatusUseCaseOpts {
  userId: string
  status: UserStatus
}

export class SetUserStatusUseCase implements IUseCase<SetUserStatusUseCaseOpts, void> {
  constructor(
    private readonly userRepository: UsersRepository,
    private readonly sessionRepository: SessionsRepository,
  ) {}

  async execute(opts: SetUserStatusUseCaseOpts): Promise<void> {
    await this.userRepository.setStatus({ id: opts.userId, status: opts.status })
    // Disabling revokes immediately — kill every session the user holds.
    if (opts.status === 'DISABLED') {
      await this.sessionRepository.deleteAllForUser({ userId: opts.userId })
    }
  }
}
