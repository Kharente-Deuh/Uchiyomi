// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../../shared/use-case'
import type { SessionsRepository } from '../../../sessions/session.domain'

export interface LogoutUseCaseOpts {
  sessionId: string
}

export class LogoutUseCase implements IUseCase<LogoutUseCaseOpts, void> {
  constructor(
    private readonly sessionRepository: SessionsRepository,
  ) {}

  execute(opts: LogoutUseCaseOpts): Promise<void> {
    return this.sessionRepository.delete({ sessionId: opts.sessionId })
  }
}
