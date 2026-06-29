// SPDX-License-Identifier: AGPL-3.0-or-later

import type { H3Event } from 'h3'
import type { UserModel } from '~~/server/domains/identity/users/user.domain'

export interface AuthGuardOpts {
  // Route requires the extension-management capability (403 otherwise).
  mustBeAbleToManage?: boolean
}

export function authGuard(event: H3Event, opts?: AuthGuardOpts): UserModel {
  const { authUser } = event.context
  if (!authUser) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthenticated' })
  }

  if (opts?.mustBeAbleToManage && !authUser.canManageExtensions) {
    throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
  }

  return authUser
}
