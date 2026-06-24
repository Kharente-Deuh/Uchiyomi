// SPDX-License-Identifier: AGPL-3.0-or-later
import type { PrismaClient } from '../../../../../../../prisma/generated/client'
import type { CreateUserWithLocalIdentityOpts, CreateUserWithLocalIdentityParams, FindUserByAccountNameParams, FindUserByIdParams, FindUserLocalPasswordHashParams, SetUserStatusParams, UpdateUserCapabilitiesParams, UpdateUserDisplayNameParams, UpdateUserLocalPasswordHashParams, UpdateUserNsfwPreferenceParams, UserModel, UsersRepository } from '../../../user.domain'
import { toDomain } from './prisma-user-repository.mapper'

export class PrismaUserRepository implements UsersRepository {
  constructor(private readonly prisma: PrismaClient) {}

  countUsers(): Promise<number> {
    return this.prisma.appUser.count()
  }

  async findByAccountName(params: FindUserByAccountNameParams): Promise<UserModel | undefined> {
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

  async findById(params: FindUserByIdParams): Promise<UserModel | undefined> {
    const row = await this.prisma.appUser.findUnique({ where: { id: params.id } })

    return row ? toDomain(row) : undefined
  }

  createWithLocalIdentity(
    params: CreateUserWithLocalIdentityParams,
    opts?: CreateUserWithLocalIdentityOpts,
  ): Promise<Omit<UserModel, 'passwordHash'>> {
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
          showNsfw: params.showNsfw ?? false,
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

  async setStatus(params: SetUserStatusParams): Promise<void> {
    await this.prisma.appUser.update({ where: { id: params.id }, data: { status: params.status } })
  }

  async updateDisplayName(params: UpdateUserDisplayNameParams): Promise<Omit<UserModel, 'passwordHash'>> {
    const row = await this.prisma.appUser.update({
      where: { id: params.id },
      data: { displayName: params.displayName },
    })

    return toDomain(row)
  }

  async updateNsfwPreference(params: UpdateUserNsfwPreferenceParams): Promise<Omit<UserModel, 'passwordHash'>> {
    const row = await this.prisma.appUser.update({
      where: { id: params.id },
      data: { showNsfw: params.showNsfw },
    })

    return toDomain(row)
  }

  async updateLocalPasswordHash(params: UpdateUserLocalPasswordHashParams): Promise<void> {
    await this.prisma.authIdentity.updateMany({
      where: { userId: params.userId, provider: 'LOCAL' },
      data: { passwordHash: params.passwordHash },
    })
  }

  async findLocalPasswordHash(params: FindUserLocalPasswordHashParams): Promise<string | undefined> {
    const row = await this.prisma.authIdentity.findFirst({
      where: { userId: params.userId, provider: 'LOCAL' },
      select: { passwordHash: true },
    })

    return row?.passwordHash ?? undefined
  }

  async updateCapabilities(params: UpdateUserCapabilitiesParams): Promise<Omit<UserModel, 'passwordHash'>> {
    const row = await this.prisma.appUser.update({
      where: { id: params.id },
      data: {
        ...(params.canManageExtensions !== undefined && { canManageExtensions: params.canManageExtensions }),
        ...(params.canDownload !== undefined && { canDownload: params.canDownload }),
        ...(params.allowNsfw !== undefined && { allowNsfw: params.allowNsfw }),
      },
    })

    return toDomain(row)
  }
}
