// SPDX-License-Identifier: AGPL-3.0-or-later
export default defineNuxtRouteMiddleware((to) => {
  const authStore = useAuthStore()
  const { redirect } = resolveAuthGuard(to.path, authStore.isAuthenticated)

  if (redirect && redirect !== to.path) {
    return navigateTo(redirect)
  }
})
