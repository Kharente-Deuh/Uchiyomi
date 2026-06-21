// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../../shared/use-case'
import type * as User from '../../../users/user.domain'
import * as Auth from '../../../auth/auth.domain'
import * as Session from '../../session.domain'

export interface Opts {
  sessionId: string
}

export class UseCase implements IUseCase<Opts, User.Model> {
  constructor(
    private readonly userRepository: User.Repository,
    private readonly sessionRepository: Session.Repository,
    private readonly ttlMs: number,
    private readonly refreshThresholdMs: number,
    private readonly now: () => Date = () => new Date(),
  ) {}

  async execute(opts: Opts): Promise<User.Model> {
    const now = this.now()
    const session = await this.sessionRepository.findValid({ sessionId: opts.sessionId, now })
    if (!session) {
      throw new Auth.AuthError('unauthenticated', 'No valid session')
    }

    const user = await this.userRepository.findById({ id: session.userId })
    if (!user || !user.isActive()) {
      throw new Auth.AuthError('unauthenticated', 'No valid session')
    }

    if (session.shouldRefresh(now, this.refreshThresholdMs)) {
      await this.sessionRepository.touch({ sessionId: session.id, expiresAt: Session.newExpiry(now, this.ttlMs) })
    }

    return user
  }
}
