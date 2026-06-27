// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../../shared'
import type { UserModel, UsersRepository } from '../../user.domain'

export interface UpdateUserNameUseCaseOpts {
  id: string
  displayName: string
}

export class UpdateUserNameUseCase implements IUseCase<UpdateUserNameUseCaseOpts, Omit<UserModel, 'passwordHash'>> {
  constructor(private readonly userRepository: UsersRepository) {}

  execute(opts: UpdateUserNameUseCaseOpts): Promise<Omit<UserModel, 'passwordHash'>> {
    return this.userRepository.updateDisplayName({ id: opts.id, displayName: opts.displayName })
  }
}
