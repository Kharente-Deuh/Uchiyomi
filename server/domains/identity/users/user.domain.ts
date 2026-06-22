// SPDX-License-Identifier: AGPL-3.0-or-later
// Flat ESM module, consumed via `import * as User`.
export type Role = 'ADMIN' | 'USER'
export type Status = 'ACTIVE' | 'DISABLED'

export type ModelProps = Omit<Model, 'canManageUsers' | 'isActive'>

export class Model {
  declare id: string
  declare accountName: string
  declare displayName: string
  declare role: Role
  declare status: Status
  declare canManageExtensions: boolean
  declare canDownload: boolean
  declare allowNsfw: boolean
  // Stays on the Model (findByAccountName carries it for credential checks); NEVER
  // crosses HTTP — routes return DTOs via the presenter, never the raw Model.
  declare passwordHash?: string

  constructor(data: ModelProps) {
    Object.assign<ModelProps, ModelProps>(this, data)
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
  role?: Role
  canManageExtensions?: boolean
  canDownload?: boolean
  allowNsfw?: boolean
}

export interface FindByAccountNameParams {
  accountName: string
}

export interface FindByIdParams {
  id: string
}

export type CreateWithLocalIdentityParams = CreateParams & { passwordHash: string }

export interface CreateWithLocalIdentityOpts {
  onlyIfEmpty: boolean
}

export interface SetStatusParams {
  id: string
  status: Status
}

export interface UpdateDisplayNameParams {
  id: string
  displayName: string
}

export interface UpdateLocalPasswordHashParams {
  userId: string
  passwordHash: string
}

export interface FindLocalPasswordHashParams {
  userId: string
}

// The PORT — adapters in infrastructure implement it.
export interface Repository {
  countUsers: () => Promise<number>
  findByAccountName: (params: FindByAccountNameParams) => Promise<Model | undefined>
  findById: (params: FindByIdParams) => Promise<Model | undefined>
  createWithLocalIdentity: (
    params: CreateWithLocalIdentityParams,
    opts?: CreateWithLocalIdentityOpts,
  ) => Promise<Omit<Model, 'passwordHash'>>
  setStatus: (params: SetStatusParams) => Promise<void>
  updateDisplayName: (params: UpdateDisplayNameParams) => Promise<Omit<Model, 'passwordHash'>>
  updateLocalPasswordHash: (params: UpdateLocalPasswordHashParams) => Promise<void>
  findLocalPasswordHash: (params: FindLocalPasswordHashParams) => Promise<string | undefined>
}
