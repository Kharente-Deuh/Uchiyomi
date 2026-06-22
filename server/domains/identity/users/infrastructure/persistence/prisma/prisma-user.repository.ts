// SPDX-License-Identifier: AGPL-3.0-or-later
import type { PrismaClient } from '../../../../../../../prisma/generated/client'
import type * as User from '../../../user.domain'
import { toDomain } from './prisma-user-repository.mapper'

export class PrismaUserRepository implements User.Repository {
  constructor(private readonly prisma: PrismaClient) {}

  countUsers(): Promise<number> {
    return this.prisma.appUser.count()
  }

  async findByAccountName(params: User.FindByAccountNameParams): Promise<User.Model | undefined> {
    const row = await this.prisma.appUser.findUnique({
      where: { accountName: params.accountName },
      include: { identities: { where: { provider: 'LOCAL' }, take: 1 } },
    })
    if (!row) {
      return undefined
    }

    const user = toDomain(row)
    user.passwordHash = row.identities[0]?.passwordHash ?? undefined

    return user
  }

  async findById(params: User.FindByIdParams): Promise<User.Model | undefined> {
    const row = await this.prisma.appUser.findUnique({ where: { id: params.id } })

    return row ? toDomain(row) : undefined
  }

  createWithLocalIdentity(
    params: User.CreateWithLocalIdentityParams,
    opts?: User.CreateWithLocalIdentityOpts,
  ): Promise<Omit<User.Model, 'passwordHash'>> {
    return this.prisma.$transaction(async (tx) => {
      // Race-safe first-run guard: re-check emptiness inside the transaction.
      if (opts?.onlyIfEmpty && (await tx.appUser.count()) > 0) {
        throw new Error('setup_closed: a user already exists')
      }

      // SECURITY: only safe fields are written — plaintext `params.password` is
      // intentionally excluded. The hashed credential lives exclusively in
      // `identities.create.passwordHash`.
      const created = await tx.appUser.create({
        data: {
          accountName: params.accountName,
          displayName: params.displayName,
          role: params.role ?? 'USER',
          canManageExtensions: params.canManageExtensions ?? false,
          canDownload: params.canDownload ?? false,
          allowNsfw: params.allowNsfw ?? false,
          identities: {
            create: {
              provider: 'LOCAL',
              subject: params.accountName,
              passwordHash: params.passwordHash,
            },
          },
        },
      })

      return toDomain(created)
    })
  }

  async setStatus(params: User.SetStatusParams): Promise<void> {
    await this.prisma.appUser.update({ where: { id: params.id }, data: { status: params.status } })
  }

  async updateDisplayName(params: User.UpdateDisplayNameParams): Promise<Omit<User.Model, 'passwordHash'>> {
    const row = await this.prisma.appUser.update({
      where: { id: params.id },
      data: { displayName: params.displayName },
    })

    return toDomain(row)
  }

  async updateLocalPasswordHash(params: User.UpdateLocalPasswordHashParams): Promise<void> {
    await this.prisma.authIdentity.updateMany({
      where: { userId: params.userId, provider: 'LOCAL' },
      data: { passwordHash: params.passwordHash },
    })
  }

  async findLocalPasswordHash(params: User.FindLocalPasswordHashParams): Promise<string | undefined> {
    const row = await this.prisma.authIdentity.findFirst({
      where: { userId: params.userId, provider: 'LOCAL' },
      select: { passwordHash: true },
    })

    return row?.passwordHash ?? undefined
  }
}
