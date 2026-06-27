// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../../shared'
import type { UserModel, UsersRepository } from '../../user.domain'

export interface UpdateUserNsfwPreferenceUseCaseOpts {
  id: string
  showNsfw: boolean
}

export class UpdateUserNsfwPreferenceUseCase implements IUseCase<UpdateUserNsfwPreferenceUseCaseOpts, Omit<UserModel, 'passwordHash'>> {
  constructor(private readonly userRepository: UsersRepository) {}

  execute(opts: UpdateUserNsfwPreferenceUseCaseOpts): Promise<Omit<UserModel, 'passwordHash'>> {
    return this.userRepository.updateNsfwPreference({ id: opts.id, showNsfw: opts.showNsfw })
  }
}
