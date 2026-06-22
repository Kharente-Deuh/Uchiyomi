// SPDX-License-Identifier: AGPL-3.0-or-later
import type { UpdateMeRequestDto } from '#shared/dto/identity/me.request'
import { z } from 'zod'
import { toUserDto } from '../../domains/identity/users/infrastructure/transport/http/user-http.presenter'
import { displayNameSchema } from '../../utils/display-name'
import { updateDisplayName } from '../../utils/identity'

const Body = z.object({ displayName: displayNameSchema }) satisfies z.ZodType<UpdateMeRequestDto>

export default defineEventHandler(async (event) => {
  const actor = event.context.authUser
  if (!actor) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthenticated' })
  }

  const parsed = await readValidatedBody(event, Body.safeParse)
  if (!parsed.success) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid body' })
  }

  const user = await updateDisplayName.execute({ id: actor.id, displayName: parsed.data.displayName })

  return { user: toUserDto(user) }
})
