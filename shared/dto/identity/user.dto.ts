// SPDX-License-Identifier: AGPL-3.0-or-later
// Wire contract for a user. Lives in the Nuxt `shared/` layer so BOTH server and
// client can import it. MUST NOT import from `server/` — hence inline literal
// unions instead of the server domain types. Never carries `passwordHash`.
export interface UserDto {
  id: string
  accountName: string
  displayName: string
  role: 'ADMIN' | 'USER'
  status: 'ACTIVE' | 'DISABLED'
  canManageExtensions: boolean
  canDownload: boolean
  allowNsfw: boolean
  showNsfw: boolean
}
