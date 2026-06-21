// SPDX-License-Identifier: AGPL-3.0-or-later
import { logout } from '../../utils/identity'

export default defineEventHandler(async (event) => {
  const session = await getUserSession(event)
  const sessionId = (session as { sessionId?: string }).sessionId
  if (sessionId) {
    await logout.execute({ sessionId })
  }

  await clearUserSession(event)

  return { ok: true }
})
