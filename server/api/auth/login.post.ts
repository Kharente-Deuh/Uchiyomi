// SPDX-License-Identifier: AGPL-3.0-or-later

import type { LoginRequestDto } from '#shared/dto/identity/auth.request'
import { z } from 'zod'
import { authService, loginRateLimiter } from '~~/server/domains/identity/auth/application/auth.service'
import { AuthError } from '~~/server/domains/identity/auth/auth.domain'
import { parseBody } from '~~/server/utils/request.util'
import { accountNameSchema } from '../../utils/account-name'

const Body = z.object({ accountName: accountNameSchema, password: z.string() }) satisfies z.ZodType<LoginRequestDto>

export default defineEventHandler(async (event): Promise<void> => {
  const body = await parseBody(event, Body)
  // xForwardedFor: true is safe only because this app is deployed behind a
  // trusted reverse proxy that authoritatively sets X-Forwarded-For (ADR-0005,
  // single-instance self-hosted assumption). The rate-limit key includes the
  // (normalized) account name, so single-account brute force is capped even
  // when the IP portion cannot be fully trusted.
  const ip = getRequestIP(event, { xForwardedFor: true }) ?? 'unknown'
  const key = `${ip}:${body.accountName}`
  if (!loginRateLimiter.check({ key })) {
    throw createError({ statusCode: 429, statusMessage: 'Too many attempts' })
  }

  try {
    const session = await authService().login({
      accountName: body.accountName,
      password: body.password,
      userAgent: getHeader(event, 'user-agent') ?? undefined,
      ip,
    })
    loginRateLimiter.reset({ key })
    await setUserSession(event, { sessionId: session.id })
  } catch (err) {
    if (err instanceof AuthError) {
      throw createError({ statusCode: 401, statusMessage: 'Invalid credentials' })
    }

    throw err
  }
})
