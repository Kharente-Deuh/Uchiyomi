import type { UserModel } from '~~/server/domains/identity/users/user.domain'
// SPDX-License-Identifier: AGPL-3.0-or-later
import type { UpdateMeRequestDto } from '#shared/dto/identity/me.request'
import { z } from 'zod'
import { usersService } from '~~/server/domains/identity/users/application/users.service'
import { toUserDto } from '../../domains/identity/users/infrastructure/transport/http/user-http.presenter'
import { displayNameSchema } from '../../utils/display-name'

const Body = z
  .object({ displayName: displayNameSchema.optional(), showNsfw: z.boolean().optional() })
  .refine(b => b.displayName !== undefined || b.showNsfw !== undefined, {
    message: 'At least one field is required',
  }) satisfies z.ZodType<UpdateMeRequestDto>

export default defineEventHandler(async (event) => {
  const actor = event.context.authUser
  if (!actor) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthenticated' })
  }

  const parsed = await readValidatedBody(event, Body.safeParse)
  if (!parsed.success) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid body' })
  }

  // actor is UserModel (see server/types/auth.d.ts); it is assignable to
  // Omit<UserModel, 'passwordHash'> — used as the initial value so the type is
  // always satisfied. The .refine above guarantees at least one branch runs and
  // reassigns `user` with a fresh DB-backed result.
  let user: Omit<UserModel, 'passwordHash'> = actor
  if (parsed.data.displayName !== undefined) {
    user = await usersService().updateUserName({ id: actor.id, displayName: parsed.data.displayName })
  }

  if (parsed.data.showNsfw !== undefined) {
    user = await usersService().updateNsfwPreference({ id: actor.id, showNsfw: parsed.data.showNsfw })
  }

  return { user: toUserDto(user) }
})
