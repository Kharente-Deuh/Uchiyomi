// SPDX-License-Identifier: AGPL-3.0-or-later

import type { UpdateUserRequestDto } from '#shared/dto/identity/admin.request'
import { z } from 'zod'
import { usersService } from '~~/server/domains/identity/users/application/users.service'
import { parseBody } from '~~/server/utils/request.util'
import { Prisma } from '../../../../prisma/generated/client'
import { toUserDto } from '../../../domains/identity/users/infrastructure/transport/http/user-http.presenter'
import { displayNameSchema } from '../../../utils/display-name'

const BodySchema = z.object({
  displayName: displayNameSchema.optional(),
  canManageExtensions: z.boolean().optional(),
  canDownload: z.boolean().optional(),
  allowNsfw: z.boolean().optional(),
}).refine(b => b.displayName !== undefined
  || b.canManageExtensions !== undefined
  || b.canDownload !== undefined
  || b.allowNsfw !== undefined, { message: 'At least one field is required' }) satisfies z.ZodType<UpdateUserRequestDto>

export default defineEventHandler(async (event) => {
  const { authUser } = event.context
  if (!authUser) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthenticated' })
  }

  const userId = getRouterParam(event, 'id')
  if (!userId) {
    throw createError({ statusCode: 400, statusMessage: 'Missing id' })
  }

  // Admin-only: this route grants capabilities (canManageExtensions, canDownload,
  // allowNsfw) that a user must never set on themselves — allowNsfw in particular
  // is admin-granted (ADR-0012). Self-service edits go through /api/auth/me.
  if (!authUser.canManageUsers()) {
    throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
  }

  const { displayName, ...caps } = await parseBody(event, BodySchema)

  try {
    let user
    if (displayName !== undefined) {
      user = await usersService().updateUserName({ id: userId, displayName })
    }

    if (caps.canManageExtensions !== undefined || caps.canDownload !== undefined || caps.allowNsfw !== undefined) {
      user = await usersService().updateUserCapabilities({ id: userId, ...caps })
    }

    return { user: toUserDto(user!) }
  } catch (err) {
    if (err instanceof Prisma.PrismaClientKnownRequestError && err.code === 'P2025') {
      throw createError({ statusCode: 404, statusMessage: 'User not found' })
    }

    throw err
  }
})
