// SPDX-License-Identifier: AGPL-3.0-or-later

/** Routes reachable without a session. `/setup` joins with M3.3/M3.4. */
export const PUBLIC_ROUTES = ['/login'] as const

type PublicRoute = (typeof PUBLIC_ROUTES)[number]

/**
 * UX routing decision only (client-side). Server-side session enforcement on
 * `/api/*` (M2.2) is the real security boundary.
 */
export function resolveAuthGuard(path: string, isAuthenticated: boolean): { redirect?: string } {
  const isPublic = (PUBLIC_ROUTES as readonly string[]).includes(path)

  if (isPublic) {
    return isAuthenticated && path === ('/login' satisfies PublicRoute) ? { redirect: '/' } : {}
  }

  return isAuthenticated ? {} : { redirect: '/login' }
}
