// SPDX-License-Identifier: AGPL-3.0-or-later
import type { CreateUserRequestDto } from '#shared/dto/identity/admin.request'
import { z } from 'zod'
import { createUser } from '~~/server/domains/identity/auth/application'
import { Prisma } from '../../../../prisma/generated/client'
import { toUserDto } from '../../../domains/identity/users/infrastructure/transport/http/user-http.presenter'
import { accountNameSchema } from '../../../utils/account-name'

export default defineEventHandler(async (event) => {
  const cfg = useRuntimeConfig(event).auth
  const Body = z.object({
    accountName: accountNameSchema,
    displayName: z.string().min(1),
    password: z.string().min(cfg.minPasswordLength),
    canManageExtensions: z.boolean().optional(),
    canDownload: z.boolean().optional(),
    allowNsfw: z.boolean().optional(),
  }) satisfies z.ZodType<CreateUserRequestDto>

  const actor = event.context.authUser
  if (!actor || !actor.canManageUsers()) {
    throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
  }

  const parsed = await readValidatedBody(event, Body.safeParse)
  if (!parsed.success) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid body' })
  }

  try {
    const user = await createUser.execute(parsed.data)

    return { user: toUserDto(user) }
  } catch (err) {
    if (err instanceof Prisma.PrismaClientKnownRequestError && err.code === 'P2002') {
      throw createError({ statusCode: 409, statusMessage: 'Account name already taken' })
    }

    throw err
  }
})
