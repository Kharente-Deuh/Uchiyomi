// SPDX-License-Identifier: AGPL-3.0-or-later

export default defineNuxtRouteMiddleware((to) => {
  const authStore = useAuthStore()
  const { redirect } = resolveAuthGuard(to.path, {
    isAuthenticated: authStore.isAuthenticated,
    needsAdmin: authStore.needsAdmin,
  })

  if (redirect && redirect !== to.fullPath) {
    return navigateTo(redirect)
  }
})
