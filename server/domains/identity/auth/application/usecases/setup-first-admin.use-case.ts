// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../../shared/use-case'
import type { PasswordHasher } from '../../../password/password.domain'
import type { UserModel, UsersRepository } from '../../../users/user.domain'
import { normalizeAccountName } from '../../../../../../shared/dto/identity/account-name'

export interface SetupFirstAdminUseCaseOpts {
  accountName: string
  displayName: string
  password: string
}

export class SetupFirstAdminUseCase implements IUseCase<SetupFirstAdminUseCaseOpts, Omit<UserModel, 'passwordHash'>> {
  constructor(
    private readonly userRepository: UsersRepository,
    private readonly passwordHasher: PasswordHasher,
  ) {}

  async execute(opts: SetupFirstAdminUseCaseOpts): Promise<Omit<UserModel, 'passwordHash'>> {
    const passwordHash = await this.passwordHasher.hash({ password: opts.password })

    return this.userRepository.createWithLocalIdentity(
      {
        accountName: normalizeAccountName(opts.accountName),
        displayName: opts.displayName,
        password: opts.password,
        role: 'ADMIN',
        passwordHash,
        canManageExtensions: true,
        canDownload: true,
        allowNsfw: true,
      },
      { onlyIfEmpty: true },
    )
  }
}
