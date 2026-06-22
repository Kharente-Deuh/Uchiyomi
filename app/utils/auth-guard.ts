// SPDX-License-Identifier: AGPL-3.0-or-later

interface GuardState {
  isAuthenticated: boolean
  needsAdmin: boolean
}

/**
 * UX routing decision only (client-side). Server-side session enforcement on
 * `/api/*` (M2.2) is the real security boundary.
 */
export function resolveAuthGuard(path: string, state: GuardState): { redirect?: string } {
  // First run: nothing is reachable until the first admin exists.
  if (state.needsAdmin) {
    return path === '/setup' ? {} : { redirect: '/setup' }
  }

  // Setup is unreachable once an admin exists.
  if (path === '/setup') {
    return { redirect: state.isAuthenticated ? '/' : '/login' }
  }

  if (path === '/login') {
    return state.isAuthenticated ? { redirect: '/' } : {}
  }

  return state.isAuthenticated ? {} : { redirect: `/login?redirect=${encodeURIComponent(path)}` }
}
