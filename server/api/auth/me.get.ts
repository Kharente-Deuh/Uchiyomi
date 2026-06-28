// SPDX-License-Identifier: AGPL-3.0-or-later

import { toUserDto } from '../../domains/identity/users/infrastructure/transport/http/user-http.presenter'

export default defineEventHandler((event) => {
  if (!event.context.authUser) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthenticated' })
  }

  return { user: toUserDto(event.context.authUser) }
})
