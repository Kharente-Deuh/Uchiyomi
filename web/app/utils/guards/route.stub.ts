// SPDX-License-Identifier: AGPL-3.0-or-later

import type { RouteLocationNormalized, RouteLocationNormalizedGeneric } from 'vue-router'

export function routeStub(partial: Partial<RouteLocationNormalizedGeneric>): RouteLocationNormalized {
  return {
    fullPath: '/',
    meta: {},
    ...partial,
  } as RouteLocationNormalized
}
