// SPDX-License-Identifier: AGPL-3.0-or-later
import { onSessionLost } from '~/utils/api'

export default defineNuxtPlugin(() => {
  const authStore = useAuthStore()
  const session = useUserSession()

  onSessionLost(() => {
    // Only react if we believed we were authenticated — an expected 401 (boot
    // probe, login attempt) leaves the store unauthenticated and is a no-op.
    if (!authStore.isAuthenticated) {
      return
    }

    authStore.clear()
    session.clear()
    navigateTo({ path: '/login', query: { reason: 'expired' } })
  })
})
