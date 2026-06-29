// SPDX-License-Identifier: AGPL-3.0-or-later

import type { IUseCase } from '../../../../../shared'
import type { UserModel, UsersRepository } from '../../user.domain'

export interface UpdateUserCapabilitiesUseCaseOpts {
  id: string
  canManageExtensions?: boolean
  canDownload?: boolean
  allowNsfw?: boolean
}

export class UpdateUserCapabilitiesUseCase implements IUseCase<UpdateUserCapabilitiesUseCaseOpts, Omit<UserModel, 'passwordHash'>> {
  constructor(private readonly userRepository: UsersRepository) {}

  execute(opts: UpdateUserCapabilitiesUseCaseOpts): Promise<Omit<UserModel, 'passwordHash'>> {
    return this.userRepository.updateCapabilities(opts)
  }
}
