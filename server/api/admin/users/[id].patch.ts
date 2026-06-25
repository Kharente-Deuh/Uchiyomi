// SPDX-License-Identifier: AGPL-3.0-or-later
import type { UpdateUserRequestDto } from '#shared/dto/identity/admin.request'
import { z } from 'zod'
import { usersService } from '~~/server/domains/identity/users/application/users.service'
import { Prisma } from '../../../../prisma/generated/client'
import { toUserDto } from '../../../domains/identity/users/infrastructure/transport/http/user-http.presenter'
import { displayNameSchema } from '../../../utils/display-name'

const Body = z
  .object({
    displayName: displayNameSchema.optional(),
    canManageExtensions: z.boolean().optional(),
    canDownload: z.boolean().optional(),
    allowNsfw: z.boolean().optional(),
  })
  .refine(b => b.displayName !== undefined
    || b.canManageExtensions !== undefined
    || b.canDownload !== undefined
    || b.allowNsfw !== undefined, { message: 'At least one field is required' }) satisfies z.ZodType<UpdateUserRequestDto>

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
    let user
    const { displayName, ...caps } = parsed.data
    if (displayName !== undefined) {
      user = await usersService().updateUserName({ id, displayName })
    }

    if (caps.canManageExtensions !== undefined || caps.canDownload !== undefined || caps.allowNsfw !== undefined) {
      user = await usersService().updateUserCapabilities({ id, ...caps })
    }

    return { user: toUserDto(user!) }
  } catch (err) {
    if (err instanceof Prisma.PrismaClientKnownRequestError && err.code === 'P2025') {
      throw createError({ statusCode: 404, statusMessage: 'User not found' })
    }

    throw err
  }
})
