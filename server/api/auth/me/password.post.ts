// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ChangePasswordRequestDto } from '#shared/dto/identity/me.request'
import { z } from 'zod'
import * as Auth from '../../../domains/identity/auth/auth.domain'
import { changePassword } from '../../../utils/identity'

export default defineEventHandler(async (event) => {
  const cfg = useRuntimeConfig(event).auth
  const actor = event.context.authUser
  if (!actor) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthenticated' })
  }

  const Body = z.object({
    currentPassword: z.string().min(1),
    newPassword: z.string().min(cfg.minPasswordLength),
    logoutOtherDevices: z.boolean().optional(),
  })
    .refine(d => d.newPassword !== d.currentPassword, { message: 'newPasswordSameAsCurrent', path: ['newPassword'] }) satisfies z.ZodType<ChangePasswordRequestDto>

  const parsed = await readValidatedBody(event, Body.safeParse)
  if (!parsed.success) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid body' })
  }

  const session = await getUserSession(event)
  const currentSessionId = (session as { sessionId?: string }).sessionId
  if (!currentSessionId) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthenticated' })
  }

  const body = parsed.data
  try {
    await changePassword.execute({
      userId: actor.id,
      currentPassword: body.currentPassword,
      newPassword: body.newPassword,
      logoutOtherDevices: body.logoutOtherDevices ?? false,
      currentSessionId,
    })

    return { ok: true }
  } catch (err) {
    if (err instanceof Auth.AuthError) {
      throw createError({ statusCode: 400, statusMessage: 'Invalid password' })
    }

    throw err
  }
})
