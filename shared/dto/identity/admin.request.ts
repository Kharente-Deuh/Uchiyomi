// SPDX-License-Identifier: AGPL-3.0-or-later
// Request wire contracts for the admin user-management routes.
export interface CreateUserRequestDto {
  accountName: string
  displayName: string
  password: string
  canManageExtensions?: boolean
  canDownload?: boolean
  allowNsfw?: boolean
}

export interface SetUserStatusRequestDto {
  status: 'ACTIVE' | 'DISABLED'
}
