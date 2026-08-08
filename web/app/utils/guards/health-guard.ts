// SPDX-License-Identifier: AGPL-3.0-or-later

import type { RouteLocationNormalized, RouteLocationRaw } from 'vue-router'
import type { ServerStatusResponse } from '~/composables/health.api'
import { statusRedirect } from '../routes'

export function requiresHealthCheck(to: RouteLocationNormalized): boolean {
  return to.name !== 'status'
}

export function resolveHealthGuard(to: RouteLocationNormalized, state: ServerStatusResponse): RouteLocationRaw | undefined {
  return state.status === 'ok'
    ? undefined
    : statusRedirect(to)
}
