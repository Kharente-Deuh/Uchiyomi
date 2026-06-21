// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../../shared/use-case'
import type * as Password from '../../../password/password.domain'
import type * as User from '../../../users/user.domain'

export interface Opts {
  email: string
  displayName: string
  password: string
}

export class UseCase implements IUseCase<Opts, Omit<User.Model, 'passwordHash'>> {
  constructor(
    private readonly userRepository: User.Repository,
    private readonly passwordHasher: Password.Hasher,
  ) {}

  async execute(opts: Opts): Promise<Omit<User.Model, 'passwordHash'>> {
    const passwordHash = await this.passwordHasher.hash({ password: opts.password })

    return this.userRepository.createWithLocalIdentity(
      {
        email: opts.email,
        displayName: opts.displayName,
        password: opts.password,
        role: 'ADMIN',
        passwordHash,
      },
      { onlyIfEmpty: true },
    )
  }
}
