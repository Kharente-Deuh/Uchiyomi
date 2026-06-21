// SPDX-License-Identifier: AGPL-3.0-or-later
import type { SetupRequestDto } from '#shared/dto/identity/auth.request'
import { z } from 'zod'
import { Prisma } from '../../../prisma/generated/client'
import * as Session from '../../domains/identity/sessions/session.domain'
import { toUserDto } from '../../domains/identity/users/infrastructure/transport/http/user-http.presenter'
import { sessionRepository, setupFirstAdmin } from '../../utils/identity'

function isSetupClosed(err: unknown): boolean {
  return (err instanceof Error && err.message.startsWith('setup_closed'))
    || (err instanceof Prisma.PrismaClientKnownRequestError && err.code === 'P2002')
}

export default defineEventHandler(async (event) => {
  const cfg = useRuntimeConfig(event).auth
  const Body = z.object({
    email: z.email(),
    displayName: z.string().min(1),
    password: z.string().min(cfg.minPasswordLength),
  }) satisfies z.ZodType<SetupRequestDto>

  const parsed = await readValidatedBody(event, Body.safeParse)
  if (!parsed.success) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid body' })
  }

  const body = parsed.data
  try {
    const admin = await setupFirstAdmin.execute(body)
    // Auto-login the new admin.
    const session = await sessionRepository.create({
      userId: admin.id,
      expiresAt: Session.newExpiry(new Date(), cfg.sessionTtlMs),
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
