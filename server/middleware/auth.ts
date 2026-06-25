// SPDX-License-Identifier: AGPL-3.0-or-later
import { sessionsService } from '../domains/identity/sessions/application/sessions.service'

// Loads the DB-backed session/user per request and attaches it to the context.
// The sealed cookie carries only { sessionId }; the session table is source of truth.
export default defineEventHandler(async (event) => {
  event.context.authUser = null

  try {
    const session = await getUserSession(event) // nuxt-auth-utils: reads the sealed cookie
    const sessionId = (session as { sessionId?: string }).sessionId
    if (!sessionId) {
      return
    }

    event.context.authUser = await sessionsService().getCurrentUser({ sessionId })
  } catch {
    // Invalid/expired/disabled — clear the stale cookie, stay unauthenticated.
    await clearUserSession(event)
  }
})
