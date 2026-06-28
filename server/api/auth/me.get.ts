// SPDX-License-Identifier: AGPL-3.0-or-later

import { authGuard } from '~~/server/domains/identity/auth/infrastructure/http/guards/auth.guard'
import { toUserDto } from '../../domains/identity/users/infrastructure/transport/http/user-http.presenter'

export default defineEventHandler((event) => {
  const authUser = authGuard(event)

  return { user: toUserDto(authUser) }
})
