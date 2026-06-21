// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../../shared/use-case'
import type * as Session from '../../../sessions/session.domain'

export interface Opts {
  sessionId: string
}

export class UseCase implements IUseCase<Opts, void> {
  constructor(
    private readonly sessionRepository: Session.Repository,
  ) {}

  execute(opts: Opts): Promise<void> {
    return this.sessionRepository.delete({ sessionId: opts.sessionId })
  }
}
