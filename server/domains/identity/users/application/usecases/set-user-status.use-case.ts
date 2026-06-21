// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../../shared/use-case'
import type * as Session from '../../../sessions/session.domain'
import type * as User from '../../user.domain'

export interface Opts {
  userId: string
  status: User.Status
}

export class UseCase implements IUseCase<Opts, void> {
  constructor(
    private readonly userRepository: User.Repository,
    private readonly sessionRepository: Session.Repository,
  ) {}

  async execute(opts: Opts): Promise<void> {
    await this.userRepository.setStatus({ id: opts.userId, status: opts.status })
    // Disabling revokes immediately — kill every session the user holds.
    if (opts.status === 'DISABLED') {
      await this.sessionRepository.deleteAllForUser({ userId: opts.userId })
    }
  }
}
