// SPDX-License-Identifier: AGPL-3.0-or-later

import { requiresAuthCheck, resolveAuthGuard } from '~/utils/guards'

export default defineNuxtRouteMiddleware(async (to) => {
  if (!requiresAuthCheck(to)) {
    return
  }

  const { isAuthenticated, isAdmin, fetchMe } = useAuth()

  if (await fetchMe() === 'unreachable') {
    return abortNavigation(createError({
      statusCode: 503,
      statusMessage: 'Unable to verify the current session',
    }))
  }

  const redirect = resolveAuthGuard({
    to,
    isAdmin: isAdmin.value,
    isAuthenticated: isAuthenticated.value,
  })

  if (redirect) {
    return navigateTo(redirect)
  }
})
