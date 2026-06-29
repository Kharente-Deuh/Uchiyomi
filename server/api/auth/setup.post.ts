// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SetupRequestDto } from '#shared/dto/identity/auth.request'
import { z } from 'zod'
import { authService } from '~~/server/domains/identity/auth/application/auth.service'
import { sessionRepository } from '~~/server/domains/identity/sessions/application'
import { newSessionExpiry } from '~~/server/domains/identity/sessions/session.domain'
import { parseBody } from '~~/server/utils/request.util'
import { Prisma } from '../../../prisma/generated/client'
import { toUserDto } from '../../domains/identity/users/infrastructure/transport/http/user-http.presenter'
import { accountNameSchema } from '../../utils/account-name'

function isSetupClosed(err: unknown): boolean {
  return (err instanceof Error && err.message.startsWith('setup_closed'))
    || (err instanceof Prisma.PrismaClientKnownRequestError && err.code === 'P2002')
}

export default defineEventHandler(async (event) => {
  const cfg = useRuntimeConfig(event).auth
  const BodySchema = z.object({
    accountName: accountNameSchema,
    displayName: z.string().min(1),
    password: z.string().min(cfg.minPasswordLength),
  }) satisfies z.ZodType<SetupRequestDto>

  const body = await parseBody(event, BodySchema)
  try {
    const admin = await authService().setupFirstAdmin(body)
    const session = await sessionRepository.create({
      userId: admin.id,
      expiresAt: newSessionExpiry(new Date(), cfg.sessionTtlMs),
    })
    await setUserSession(event, { sessionId: session.id })

    return { user: toUserDto(admin) }
  } catch (err) {
    if (isSetupClosed(err)) {
      throw createError({ statusCode: 409, statusMessage: 'Setup already completed' })
    }

    throw err
  }
})
