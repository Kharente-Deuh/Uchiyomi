// SPDX-License-Identifier: AGPL-3.0-or-later

import type { CreateUserRequestDto } from '#shared/dto/identity/admin.request'
import { z } from 'zod'
import { authGuard } from '~~/server/domains/identity/auth/infrastructure/http/guards/auth.guard'
import { usersService } from '~~/server/domains/identity/users/application/users.service'
import { parseBody } from '~~/server/utils/request.util'
import { Prisma } from '../../../../prisma/generated/client'
import { toUserDto } from '../../../domains/identity/users/infrastructure/transport/http/user-http.presenter'
import { accountNameSchema } from '../../../utils/account-name'

export default defineEventHandler(async (event) => {
  const cfg = useRuntimeConfig(event).auth
  const BodySchema = z.object({
    accountName: accountNameSchema,
    displayName: z.string().min(1),
    password: z.string().min(cfg.minPasswordLength),
    canManageExtensions: z.boolean().optional(),
    canDownload: z.boolean().optional(),
    allowNsfw: z.boolean().optional(),
  }) satisfies z.ZodType<CreateUserRequestDto>

  authGuard(event, { mustBeAbleToManage: true })
  const body = await parseBody(event, BodySchema)
  try {
    const user = await usersService().createUser(body)

    return { user: toUserDto(user) }
  } catch (err) {
    if (err instanceof Prisma.PrismaClientKnownRequestError && err.code === 'P2002') {
      throw createError({ statusCode: 409, statusMessage: 'Account name already taken' })
    }

    throw err
  }
})
