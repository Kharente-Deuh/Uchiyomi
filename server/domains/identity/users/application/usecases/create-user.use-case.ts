// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../../shared/use-case'
import type * as Password from '../../../password/password.domain'
import type * as User from '../../user.domain'

export type Opts = User.CreateParams

export class UseCase implements IUseCase<Opts, Omit<User.Model, 'passwordHash'>> {
  constructor(
    private readonly userRepository: User.Repository,
    private readonly passwordHasher: Password.Hasher,
  ) {}

  async execute(opts: Opts): Promise<Omit<User.Model, 'passwordHash'>> {
    const passwordHash = await this.passwordHasher.hash({ password: opts.password })

    return this.userRepository.createWithLocalIdentity({ ...opts, role: opts.role ?? 'USER', passwordHash })
  }
}
