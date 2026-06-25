// SPDX-License-Identifier: AGPL-3.0-or-later
import type { SetUserStatusRequestDto } from '#shared/dto/identity/admin.request'
import { z } from 'zod'
import { usersService } from '~~/server/domains/identity/users/application/users.service'

const Body = z.object({ status: z.enum(['ACTIVE', 'DISABLED']) }) satisfies z.ZodType<SetUserStatusRequestDto>

export default defineEventHandler(async (event): Promise<void> => {
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

  const body = parsed.data
  await usersService().setUserStatus({ userId: id, status: body.status })
})
