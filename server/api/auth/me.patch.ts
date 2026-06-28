// SPDX-License-Identifier: AGPL-3.0-or-later

import type { UpdateMeRequestDto } from '#shared/dto/identity/me.request'
import { z } from 'zod'
import { authGuard } from '~~/server/domains/identity/auth/infrastructure/http/guards/auth.guard'
import { usersService } from '~~/server/domains/identity/users/application/users.service'
import { parseBody } from '~~/server/utils/request.util'
import { toUserDto } from '../../domains/identity/users/infrastructure/transport/http/user-http.presenter'
import { displayNameSchema } from '../../utils/display-name'

const Body = z
  .object({ displayName: displayNameSchema.optional(), showNsfw: z.boolean().optional() })
  .refine(b => b.displayName !== undefined || b.showNsfw !== undefined, {
    message: 'At least one field is required',
  }) satisfies z.ZodType<UpdateMeRequestDto>

export default defineEventHandler(async (event) => {
  let user = authGuard(event)
  const { displayName, showNsfw } = await parseBody(event, Body)

  // actor is UserModel (see server/types/auth.d.ts); it is assignable to
  // Omit<UserModel, 'passwordHash'> — used as the initial value so the type is
  // always satisfied. The .refine above guarantees at least one branch runs and
  // reassigns `user` with a fresh DB-backed result.
  if (displayName !== undefined) {
    user = await usersService().updateUserName({ id: user.id, displayName })
  }

  if (showNsfw !== undefined) {
    user = await usersService().updateNsfwPreference({ id: user.id, showNsfw })
  }

  return { user: toUserDto(user) }
})
