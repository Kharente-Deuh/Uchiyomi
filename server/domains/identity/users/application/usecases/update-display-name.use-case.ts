// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../../shared/use-case'
import type * as User from '../../user.domain'

export interface Opts {
  id: string
  displayName: string
}

export class UseCase implements IUseCase<Opts, Omit<User.Model, 'passwordHash'>> {
  constructor(private readonly userRepository: User.Repository) {}

  execute(opts: Opts): Promise<Omit<User.Model, 'passwordHash'>> {
    return this.userRepository.updateDisplayName({ id: opts.id, displayName: opts.displayName })
  }
}
