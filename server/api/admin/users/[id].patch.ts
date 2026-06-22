// SPDX-License-Identifier: AGPL-3.0-or-later
import type { UpdateUserRequestDto } from '#shared/dto/identity/admin.request'
import { z } from 'zod'
import { Prisma } from '../../../../prisma/generated/client'
import { toUserDto } from '../../../domains/identity/users/infrastructure/transport/http/user-http.presenter'
import { updateDisplayName } from '../../../utils/identity'

const Body = z.object({ displayName: z.string().trim().min(1).max(64) }) satisfies z.ZodType<UpdateUserRequestDto>

export default defineEventHandler(async (event) => {
  const actor = event.context.authUser
  if (!actor || !actor.canManageUsers()) {
    throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
  }

  const id = getRouterParam(event, 'id')
  if (!id) {
    throw createError({ statusCode: 400, statusMessage: 'Missing id' })
  }

  const parsed = await readValidatedBody(event, Body.safeParse)
  if (!parsed.success) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid body' })
  }

  try {
    const user = await updateDisplayName.execute({ id, displayName: parsed.data.displayName })

    return { user: toUserDto(user) }
  } catch (err) {
    if (err instanceof Prisma.PrismaClientKnownRequestError && err.code === 'P2025') {
      throw createError({ statusCode: 404, statusMessage: 'User not found' })
    }

    throw err
  }
})
