// SPDX-License-Identifier: AGPL-3.0-or-later

// Flat ESM module, consumed via `import * as User`.
export type UserRole = 'ADMIN' | 'USER'
export type UserStatus = 'ACTIVE' | 'DISABLED'

export type UserModelProps = Omit<UserModel, 'canManageUsers' | 'isActive'>

export class UserModel {
  declare id: string
  declare accountName: string
  declare displayName: string
  declare role: UserRole
  declare status: UserStatus
  declare canManageExtensions: boolean
  declare canDownload: boolean
  declare allowNsfw: boolean
  declare showNsfw: boolean
  // Stays on the Model (findByAccountName carries it for credential checks); NEVER
  // crosses HTTP — routes return DTOs via the presenter, never the raw Model.
  declare passwordHash?: string

  constructor(data: UserModelProps) {
    Object.assign<UserModelProps, UserModelProps>(this, data)
  }

  canManageUsers(): boolean {
    return this.role === 'ADMIN'
  }

  isActive(): boolean {
    return this.status === 'ACTIVE'
  }
}

export interface CreateParams {
  accountName: string
  displayName: string
  password: string
  role?: UserRole
  canManageExtensions?: boolean
  canDownload?: boolean
  allowNsfw?: boolean
  showNsfw?: boolean
}

export interface FindUserByAccountNameParams {
  accountName: string
}

export interface FindUserByIdParams {
  id: string
}

export type CreateUserWithLocalIdentityParams = CreateParams & { passwordHash: string }

export interface CreateUserWithLocalIdentityOpts {
  onlyIfEmpty: boolean
}

export interface SetUserStatusParams {
  id: string
  status: UserStatus
}

export interface UpdateUserDisplayNameParams {
  id: string
  displayName: string
}

export interface UpdateUserNsfwPreferenceParams {
  id: string
  showNsfw: boolean
}

export interface UpdateUserLocalPasswordHashParams {
  userId: string
  passwordHash: string
}

export interface FindUserLocalPasswordHashParams {
  userId: string
}

export interface UpdateUserCapabilitiesParams {
  id: string
  canManageExtensions?: boolean
  canDownload?: boolean
  allowNsfw?: boolean
}

// The PORT — adapters in infrastructure implement it.
export interface UsersRepository {
  countUsers: () => Promise<number>
  findByAccountName: (params: FindUserByAccountNameParams) => Promise<UserModel | undefined>
  findById: (params: FindUserByIdParams) => Promise<UserModel | undefined>
  createWithLocalIdentity: (
    params: CreateUserWithLocalIdentityParams,
    opts?: CreateUserWithLocalIdentityOpts,
  ) => Promise<Omit<UserModel, 'passwordHash'>>
  setStatus: (params: SetUserStatusParams) => Promise<void>
  updateDisplayName: (params: UpdateUserDisplayNameParams) => Promise<Omit<UserModel, 'passwordHash'>>
  updateNsfwPreference: (params: UpdateUserNsfwPreferenceParams) => Promise<Omit<UserModel, 'passwordHash'>>
  updateLocalPasswordHash: (params: UpdateUserLocalPasswordHashParams) => Promise<void>
  findLocalPasswordHash: (params: FindUserLocalPasswordHashParams) => Promise<string | undefined>
  updateCapabilities: (params: UpdateUserCapabilitiesParams) => Promise<Omit<UserModel, 'passwordHash'>>
}
