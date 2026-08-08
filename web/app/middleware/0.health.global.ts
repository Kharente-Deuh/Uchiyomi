// SPDX-License-Identifier: AGPL-3.0-or-later

import { requiresHealthCheck, resolveHealthGuard } from '~/utils/guards'

export default defineNuxtRouteMiddleware(async (to) => {
  if (!requiresHealthCheck(to)) {
    return
  }

  const state = await createHealthApi().getServerStatus()
  const redirect = resolveHealthGuard(to, state)
  if (redirect) {
    return navigateTo(redirect)
  }
})
