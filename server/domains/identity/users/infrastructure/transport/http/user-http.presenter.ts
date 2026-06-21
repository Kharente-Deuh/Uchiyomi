// SPDX-License-Identifier: AGPL-3.0-or-later
import type { UserDto } from '#shared/dto/identity/user.dto'
import type * as User from '../../../user.domain'

// Maps a (password-free) user Model to the wire DTO. Explicitly NO `passwordHash`
// — the Model never crosses HTTP.
export function toUserDto(user: Omit<User.Model, 'passwordHash'>): UserDto {
  return {
    id: user.id,
    email: user.email,
    displayName: user.displayName,
    role: user.role,
    status: user.status,
    canManageExtensions: user.canManageExtensions,
    canDownload: user.canDownload,
    allowNsfw: user.allowNsfw,
  }
}
