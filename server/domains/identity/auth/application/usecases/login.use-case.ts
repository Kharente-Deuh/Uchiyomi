// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../../shared/use-case'
import type { PasswordHasher } from '../../../password/password.domain'
import type { SessionModel, SessionsRepository } from '../../../sessions/session.domain'
import type { UsersRepository } from '../../../users/user.domain'
import { normalizeAccountName } from '../../../../../../shared/dto/identity/account-name'
import { newSessionExpiry } from '../../../sessions/session.domain'
import { AuthError } from '../../auth.domain'

export interface LoginUseCaseOpts {
  accountName: string
  password: string
  userAgent?: string
  ip?: string
}

export class LoginUseCase implements IUseCase<LoginUseCaseOpts, SessionModel> {
  constructor(
    private readonly userRepository: UsersRepository,
    private readonly sessionRepository: SessionsRepository,
    private readonly passwordHasher: PasswordHasher,
    private readonly ttlMs: number,
    private readonly now: () => Date = () => new Date(),
  ) {}

  async execute(opts: LoginUseCaseOpts): Promise<SessionModel> {
    const now = this.now()
    const accountName = normalizeAccountName(opts.accountName)
    const user = await this.userRepository.findByAccountName({ accountName })
    // Uniform failure: never reveal whether the account exists, the password is
    // wrong, or the account is disabled.
    const ok = user !== undefined
      && user.isActive()
      && user.passwordHash !== undefined
      && await this.passwordHasher.verify({ hash: user.passwordHash, password: opts.password })

    if (!ok || !user) {
      throw new AuthError('invalid_credentials', 'Invalid credentials')
    }

    return this.sessionRepository.create({
      userId: user.id,
      expiresAt: newSessionExpiry(now, this.ttlMs),
      userAgent: opts.userAgent,
      ip: opts.ip,
    })
  }
}
