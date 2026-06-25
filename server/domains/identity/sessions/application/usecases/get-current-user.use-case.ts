// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../../shared/use-case'
import type { UserModel, UsersRepository } from '../../../users/user.domain'
import type { SessionsRepository } from '../../session.domain'
import { AuthError } from '../../../auth/auth.domain'
import { newSessionExpiry } from '../../session.domain'

export interface GetCurrentUserUseCaseOpts {
  sessionId: string
}

export class GetCurrentUserUseCase implements IUseCase<GetCurrentUserUseCaseOpts, UserModel> {
  constructor(
    private readonly userRepository: UsersRepository,
    private readonly sessionRepository: SessionsRepository,
    private readonly ttlMs: number,
    private readonly refreshThresholdMs: number,
    private readonly now: () => Date = () => new Date(),
  ) {}

  async execute(opts: GetCurrentUserUseCaseOpts): Promise<UserModel> {
    const now = this.now()
    const session = await this.sessionRepository.findValid({ sessionId: opts.sessionId, now })
    if (!session) {
      throw new AuthError('unauthenticated', 'No valid session')
    }

    const user = await this.userRepository.findById({ id: session.userId })
    if (!user || !user.isActive()) {
      throw new AuthError('unauthenticated', 'No valid session')
    }

    if (session.shouldRefresh(now, this.refreshThresholdMs)) {
      await this.sessionRepository.touch({
        sessionId: session.id,
        expiresAt: newSessionExpiry(now, this.ttlMs),
      })
    }

    return user
  }
}
