// SPDX-License-Identifier: AGPL-3.0-or-later

import type { RouteLocationNormalized, RouteLocationRaw } from 'vue-router'
import type { AuthRouteGroup } from '~/constants/auth'
import { DEFAULT_PAGE } from '~/constants'
import { ADMIN_ROUTE_GROUP, AUTHENTICATED_ROUTE_GROUP, NOT_AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

export interface AuthGuardOpts {
  to: RouteLocationNormalized
  isAuthenticated: boolean
  isAdmin: boolean
}

function checkAuthenticatedRoute({ to, isAuthenticated }: AuthGuardOpts): RouteLocationRaw | undefined {
  if (isAuthenticated) {
    return
  }

  if (to.name === 'index') {
    return '/login'
  }

  return `/login?redirect=${encodeURIComponent(to.fullPath)}`
}

function checkNotAuthenticatedRoute({ isAuthenticated }: AuthGuardOpts): RouteLocationRaw | undefined {
  if (isAuthenticated) {
    return DEFAULT_PAGE
  }
}

function checkAdminRoute(opts: AuthGuardOpts): RouteLocationRaw | undefined {
  return checkAuthenticatedRoute(opts) ?? (opts.isAdmin ? undefined : DEFAULT_PAGE)
}

const checkFn: Record<AuthRouteGroup, (opts: AuthGuardOpts) => RouteLocationRaw | undefined> = {
  [ADMIN_ROUTE_GROUP]: checkAdminRoute,
  [AUTHENTICATED_ROUTE_GROUP]: checkAuthenticatedRoute,
  [NOT_AUTHENTICATED_ROUTE_GROUP]: checkNotAuthenticatedRoute,
}

export function requiresAuthCheck(to: RouteLocationNormalized): boolean {
  return !!to.meta.authGroups?.length
}

export function resolveAuthGuard(opts: AuthGuardOpts): RouteLocationRaw | undefined {
  const groups = opts.to.meta.authGroups ?? []

  for (const group of groups) {
    const redirect = checkFn[group](opts)
    if (redirect) {
      return redirect
    }
  }
}
