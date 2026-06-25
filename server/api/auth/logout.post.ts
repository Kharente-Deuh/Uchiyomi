// SPDX-License-Identifier: AGPL-3.0-or-later

import { authService } from '~~/server/domains/identity/auth/application/auth.service'

export default defineEventHandler(async (event) => {
  const session = await getUserSession(event)
  const sessionId = (session as { sessionId?: string }).sessionId
  if (sessionId) {
    await authService().logout({ sessionId })
  }

  await clearUserSession(event)

  return { ok: true }
})
