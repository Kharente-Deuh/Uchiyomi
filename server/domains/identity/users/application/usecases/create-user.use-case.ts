// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../../shared'
import type { PasswordHasher } from '../../../password/password.domain'
import type { CreateUserWithLocalIdentityParams, UserModel, UsersRepository } from '../../user.domain'
import { normalizeAccountName } from '../../../../../../shared/dto/identity/account-name'

export type CreateUserUseCaseOpts = Omit<CreateUserWithLocalIdentityParams, 'passwordHash'>

export class CreateUserUseCase implements IUseCase<CreateUserUseCaseOpts, Omit<UserModel, 'passwordHash'>> {
  constructor(
    private readonly userRepository: UsersRepository,
    private readonly passwordHasher: PasswordHasher,
  ) {}

  async execute(opts: CreateUserUseCaseOpts): Promise<Omit<UserModel, 'passwordHash'>> {
    const passwordHash = await this.passwordHasher.hash({ password: opts.password })

    return this.userRepository.createWithLocalIdentity({
      ...opts,
      accountName: normalizeAccountName(opts.accountName),
      role: opts.role ?? 'USER',
      passwordHash,
    })
  }
}
