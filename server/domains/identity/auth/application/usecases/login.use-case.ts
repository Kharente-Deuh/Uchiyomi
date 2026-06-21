// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../../shared/use-case'
import type * as Password from '../../../password/password.domain'
import type * as User from '../../../users/user.domain'
import * as Session from '../../../sessions/session.domain'
import * as Auth from '../../auth.domain'

export interface Opts {
  email: string
  password: string
  userAgent?: string
  ip?: string
}

export class UseCase implements IUseCase<Opts, Session.Model> {
  constructor(
    private readonly userRepository: User.Repository,
    private readonly sessionRepository: Session.Repository,
    private readonly passwordHasher: Password.Hasher,
    private readonly ttlMs: number,
    private readonly now: () => Date = () => new Date(),
  ) {}

  async execute(opts: Opts): Promise<Session.Model> {
    const now = this.now()
    const user = await this.userRepository.findByEmail({ email: opts.email })
    // Uniform failure: never reveal whether the email exists, the password is
    // wrong, or the account is disabled.
    const ok = user !== undefined
      && user.isActive()
      && user.passwordHash !== undefined
      && await this.passwordHasher.verify({ hash: user.passwordHash, password: opts.password })
    if (!ok || !user) {
      throw new Auth.AuthError('invalid_credentials', 'Invalid email or password')
    }

    return this.sessionRepository.create({
      userId: user.id,
      expiresAt: Session.newExpiry(now, this.ttlMs),
      userAgent: opts.userAgent,
      ip: opts.ip,
    })
  }
}
